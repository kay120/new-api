package middleware

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/QuantumNous/new-api/common"
	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
)

var inMemoryRateLimiter common.InMemoryRateLimiter

var defNext = func(c *gin.Context) {
	c.Next()
}

// slidingWindowScript 原子化滑动窗口限流。
// KEYS[1] = 限流 key
// ARGV[1] = max（最大请求数）
// ARGV[2] = window_ns（窗口纳秒数）
// ARGV[3] = now_ns（当前时间纳秒）
// ARGV[4] = expire_sec（key TTL 秒）
// 返回：1=允许，0=拒绝
var slidingWindowScript = redis.NewScript(`
local len = redis.call('LLEN', KEYS[1])
local max = tonumber(ARGV[1])

if len < max then
    redis.call('LPUSH', KEYS[1], ARGV[3])
    redis.call('EXPIRE', KEYS[1], ARGV[4])
    return 1
end

local oldest = tonumber(redis.call('LINDEX', KEYS[1], -1))
local now = tonumber(ARGV[3])
local window = tonumber(ARGV[2])

if now - oldest < window then
    redis.call('EXPIRE', KEYS[1], ARGV[4])
    return 0
end

redis.call('LPUSH', KEYS[1], ARGV[3])
redis.call('LTRIM', KEYS[1], 0, max - 1)
redis.call('EXPIRE', KEYS[1], ARGV[4])
return 1
`)

// redisSlidingWindow 一次 RTT 原子执行限流检查与记录。
func redisSlidingWindow(ctx context.Context, key string, max int, durationSec int64) (bool, error) {
	expireSec := int64(common.RateLimitKeyExpirationDuration / time.Second)
	res, err := slidingWindowScript.Run(
		ctx, common.RDB,
		[]string{key},
		max, durationSec*int64(time.Second), time.Now().UnixNano(), expireSec,
	).Int()
	if err != nil {
		return false, err
	}
	return res == 1, nil
}

func redisRateLimiter(c *gin.Context, maxRequestNum int, duration int64, mark string) {
	key := "rateLimit:" + mark + c.ClientIP()
	allowed, err := redisSlidingWindow(c.Request.Context(), key, maxRequestNum, duration)
	if err != nil {
		common.SysError("rate limit redis error: " + err.Error())
		c.Status(http.StatusInternalServerError)
		c.Abort()
		return
	}
	if !allowed {
		c.Status(http.StatusTooManyRequests)
		c.Abort()
	}
}

func memoryRateLimiter(c *gin.Context, maxRequestNum int, duration int64, mark string) {
	key := mark + c.ClientIP()
	if !inMemoryRateLimiter.Request(key, maxRequestNum, duration) {
		c.Status(http.StatusTooManyRequests)
		c.Abort()
		return
	}
}

func rateLimitFactory(maxRequestNum int, duration int64, mark string) func(c *gin.Context) {
	if common.RedisEnabled {
		return func(c *gin.Context) {
			redisRateLimiter(c, maxRequestNum, duration, mark)
		}
	}
	// It's safe to call multi times.
	inMemoryRateLimiter.Init(common.RateLimitKeyExpirationDuration)
	return func(c *gin.Context) {
		memoryRateLimiter(c, maxRequestNum, duration, mark)
	}
}

func GlobalWebRateLimit() func(c *gin.Context) {
	if common.GlobalWebRateLimitEnable {
		return rateLimitFactory(common.GlobalWebRateLimitNum, common.GlobalWebRateLimitDuration, "GW")
	}
	return defNext
}

func GlobalAPIRateLimit() func(c *gin.Context) {
	if common.GlobalApiRateLimitEnable {
		return rateLimitFactory(common.GlobalApiRateLimitNum, common.GlobalApiRateLimitDuration, "GA")
	}
	return defNext
}

func CriticalRateLimit() func(c *gin.Context) {
	if common.CriticalRateLimitEnable {
		return rateLimitFactory(common.CriticalRateLimitNum, common.CriticalRateLimitDuration, "CT")
	}
	return defNext
}

func DownloadRateLimit() func(c *gin.Context) {
	return rateLimitFactory(common.DownloadRateLimitNum, common.DownloadRateLimitDuration, "DW")
}

func UploadRateLimit() func(c *gin.Context) {
	return rateLimitFactory(common.UploadRateLimitNum, common.UploadRateLimitDuration, "UP")
}

// userRateLimitFactory creates a rate limiter keyed by authenticated user ID
// instead of client IP, making it resistant to proxy rotation attacks.
// Must be used AFTER authentication middleware (UserAuth).
func userRateLimitFactory(maxRequestNum int, duration int64, mark string) func(c *gin.Context) {
	if common.RedisEnabled {
		return func(c *gin.Context) {
			userId := c.GetInt("id")
			if userId == 0 {
				c.Status(http.StatusUnauthorized)
				c.Abort()
				return
			}
			key := fmt.Sprintf("rateLimit:%s:user:%d", mark, userId)
			userRedisRateLimiter(c, maxRequestNum, duration, key)
		}
	}
	// It's safe to call multi times.
	inMemoryRateLimiter.Init(common.RateLimitKeyExpirationDuration)
	return func(c *gin.Context) {
		userId := c.GetInt("id")
		if userId == 0 {
			c.Status(http.StatusUnauthorized)
			c.Abort()
			return
		}
		key := fmt.Sprintf("%s:user:%d", mark, userId)
		if !inMemoryRateLimiter.Request(key, maxRequestNum, duration) {
			c.Status(http.StatusTooManyRequests)
			c.Abort()
			return
		}
	}
}

// userRedisRateLimiter accepts a pre-built key (to support user-ID-based keys).
func userRedisRateLimiter(c *gin.Context, maxRequestNum int, duration int64, key string) {
	allowed, err := redisSlidingWindow(c.Request.Context(), key, maxRequestNum, duration)
	if err != nil {
		common.SysError("user rate limit redis error: " + err.Error())
		c.Status(http.StatusInternalServerError)
		c.Abort()
		return
	}
	if !allowed {
		c.Status(http.StatusTooManyRequests)
		c.Abort()
	}
}

// ---------------------------------------------------------------------------
// Login brute-force guard — per-username failure counter
// ---------------------------------------------------------------------------

const (
	loginFailKeyPrefix = "loginFail:"
	// 阈值：同一用户名 5 分钟内连续 8 次失败即锁定 15 分钟。
	// 之所以比 IP 级 CriticalRateLimit 宽一点，是为了不误伤真人偶尔手滑；
	// 锁定期足够长以阻止典型字典爆破。
	loginFailThreshold  = 8
	loginFailWindowSec  = 5 * 60
	loginFailLockoutSec = 15 * 60
)

func loginFailKey(username string) string {
	return loginFailKeyPrefix + strings.ToLower(strings.TrimSpace(username))
}

// LoginBruteForceLocked 判断指定用户名当前是否因连续失败被锁定。
// Redis 未启用时降级为 false（不阻挡），此时依赖 IP 级 CriticalRateLimit 兜底。
func LoginBruteForceLocked(username string) bool {
	if !common.RedisEnabled || username == "" {
		return false
	}
	ctx := context.Background()
	val, err := common.RDB.Get(ctx, loginFailKey(username)).Int()
	if err != nil {
		return false
	}
	return val >= loginFailThreshold
}

// RecordLoginFailure 在验证失败后记一次失败，达到阈值时延长 TTL 为锁定期。
func RecordLoginFailure(username string) {
	if !common.RedisEnabled || username == "" {
		return
	}
	ctx := context.Background()
	key := loginFailKey(username)
	count, err := common.RDB.Incr(ctx, key).Result()
	if err != nil {
		common.SysError("login brute-force counter incr failed: " + err.Error())
		return
	}
	if count == 1 {
		common.RDB.Expire(ctx, key, time.Duration(loginFailWindowSec)*time.Second)
	} else if count >= int64(loginFailThreshold) {
		common.RDB.Expire(ctx, key, time.Duration(loginFailLockoutSec)*time.Second)
	}
}

// ClearLoginFailure 登录成功后清零失败计数。
func ClearLoginFailure(username string) {
	if !common.RedisEnabled || username == "" {
		return
	}
	common.RDB.Del(context.Background(), loginFailKey(username))
}

// SearchRateLimit returns a per-user rate limiter for search endpoints.
// Configurable via SEARCH_RATE_LIMIT_ENABLE / SEARCH_RATE_LIMIT / SEARCH_RATE_LIMIT_DURATION.
func SearchRateLimit() func(c *gin.Context) {
	if !common.SearchRateLimitEnable {
		return defNext
	}
	return userRateLimitFactory(common.SearchRateLimitNum, common.SearchRateLimitDuration, "SR")
}
