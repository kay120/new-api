package middleware

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/QuantumNous/new-api/common"
	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
)

const tpmRateLimitKeyPrefix = "tpmLimit:"

// TPMRateLimit TPM（每分钟 token 数）限流中间件
// TPM 限流采用滑动窗口计数器：
// - 请求前：检查过去 60 秒内的累计 token 数是否超过 TPM 限制
// - 请求后：将实际消耗的 token 数记录到 Redis
func TPMRateLimit() gin.HandlerFunc {
	return func(c *gin.Context) {
		tpmLimit, exists := c.Get("time_rate_limit_tpm")
		if !exists {
			c.Next()
			return
		}
		tpmLimitVal, ok := tpmLimit.(int)
		if !ok || tpmLimitVal <= 0 {
			c.Next()
			return
		}

		// 需要 Redis 支持 TPM 限流
		if !common.RedisEnabled {
			c.Next()
			return
		}

		userId := strconv.Itoa(c.GetInt("id"))
		key := tpmRateLimitKeyPrefix + userId
		ctx := context.Background()

		// 请求前：检查过去 60 秒的累计 token 数
		currentTPM, err := getTPMCount(ctx, key)
		if err != nil {
			// Redis 错误不阻断请求
			common.SysLog("TPM rate limit check error: " + err.Error())
			c.Next()
			return
		}

		if currentTPM >= int64(tpmLimitVal) {
			abortWithOpenAiMessage(c, http.StatusTooManyRequests,
				fmt.Sprintf("您已达到 TPM 限制：每分钟最多使用 %d tokens，当前已使用 %d tokens", tpmLimitVal, currentTPM))
			return
		}

		// 将 TPM key 存入 context，供请求完成后记录使用
		c.Set("tpm_limit_key", key)

		c.Next()

		// 请求后：记录实际消耗的 token 数
		if c.Writer.Status() < 400 {
			totalTokens := c.GetInt("usage_total_tokens")
			if totalTokens > 0 {
				recordTPMUsage(ctx, key, totalTokens)
			}
		}
	}
}

// getTPMCount 获取过去 60 秒内的累计 token 数
func getTPMCount(ctx context.Context, key string) (int64, error) {
	now := time.Now().UnixMilli()
	windowStart := now - 60*1000 // 60 秒前

	// 使用 sorted set，score 为时间戳
	// 先清理过期数据
	common.RDB.ZRemRangeByScore(ctx, key, "-inf", strconv.FormatInt(windowStart, 10))

	// 获取窗口内所有记录的 value 之和
	members, err := common.RDB.ZRangeByScore(ctx, key, &redis.ZRangeBy{
		Min: strconv.FormatInt(windowStart, 10),
		Max: "+inf",
	}).Result()
	if err != nil {
		return 0, err
	}

	var total int64
	for _, member := range members {
		// member 格式: "tokens:timestamp"
		parts := strings.SplitN(member, ":", 2)
		if len(parts) >= 1 {
			tokens, _ := strconv.ParseInt(parts[0], 10, 64)
			total += tokens
		}
	}
	return total, nil
}

// recordTPMUsage 记录一次请求的 token 消耗
func recordTPMUsage(ctx context.Context, key string, tokens int) {
	now := time.Now().UnixMilli()
	// 使用 unique member = "tokens:timestamp" 避免去重
	member := fmt.Sprintf("%d:%d", tokens, now)
	common.RDB.ZAdd(ctx, key, &redis.Z{
		Score:  float64(now),
		Member: member,
	})
	// 设置过期时间，防止 key 堆积
	common.RDB.Expire(ctx, key, 2*time.Minute)
}
