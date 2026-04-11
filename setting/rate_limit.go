package setting

import (
	"encoding/json"
	"fmt"
	"math"
	"strings"
	"sync"
	"time"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/dto"
)

var ModelRequestRateLimitEnabled = false
var ModelRequestRateLimitDurationMinutes = 1
var ModelRequestRateLimitCount = 0
var ModelRequestRateLimitSuccessCount = 1000
var ModelRequestRateLimitGroup = map[string][2]int{}
var ModelRequestRateLimitMutex sync.RWMutex

// DefaultRateLimitSchedule 全局默认分时段限流配置
var DefaultRateLimitSchedule []dto.TimeRateLimit
var DefaultRateLimitScheduleMutex sync.RWMutex

func ModelRequestRateLimitGroup2JSONString() string {
	ModelRequestRateLimitMutex.RLock()
	defer ModelRequestRateLimitMutex.RUnlock()

	jsonBytes, err := json.Marshal(ModelRequestRateLimitGroup)
	if err != nil {
		common.SysLog("error marshalling model ratio: " + err.Error())
	}
	return string(jsonBytes)
}

func UpdateModelRequestRateLimitGroupByJSONString(jsonStr string) error {
	ModelRequestRateLimitMutex.RLock()
	defer ModelRequestRateLimitMutex.RUnlock()

	ModelRequestRateLimitGroup = make(map[string][2]int)
	return json.Unmarshal([]byte(jsonStr), &ModelRequestRateLimitGroup)
}

func GetGroupRateLimit(group string) (totalCount, successCount int, found bool) {
	ModelRequestRateLimitMutex.RLock()
	defer ModelRequestRateLimitMutex.RUnlock()

	if ModelRequestRateLimitGroup == nil {
		return 0, 0, false
	}

	limits, found := ModelRequestRateLimitGroup[group]
	if !found {
		return 0, 0, false
	}
	return limits[0], limits[1], true
}

func CheckModelRequestRateLimitGroup(jsonStr string) error {
	checkModelRequestRateLimitGroup := make(map[string][2]int)
	err := json.Unmarshal([]byte(jsonStr), &checkModelRequestRateLimitGroup)
	if err != nil {
		return err
	}
	for group, limits := range checkModelRequestRateLimitGroup {
		if limits[0] < 0 || limits[1] < 1 {
			return fmt.Errorf("group %s has negative rate limit values: [%d, %d]", group, limits[0], limits[1])
		}
		if limits[0] > math.MaxInt32 || limits[1] > math.MaxInt32 {
			return fmt.Errorf("group %s [%d, %d] has max rate limits value 2147483647", group, limits[0], limits[1])
		}
	}

	return nil
}

// DefaultRateLimitSchedule2JSONString 全局默认分时段限流配置序列化
func DefaultRateLimitSchedule2JSONString() string {
	DefaultRateLimitScheduleMutex.RLock()
	defer DefaultRateLimitScheduleMutex.RUnlock()

	jsonBytes, err := json.Marshal(DefaultRateLimitSchedule)
	if err != nil {
		common.SysLog("error marshalling default rate limit schedule: " + err.Error())
		return "[]"
	}
	return string(jsonBytes)
}

// UpdateDefaultRateLimitScheduleByJSONString 更新全局默认分时段限流配置
func UpdateDefaultRateLimitScheduleByJSONString(jsonStr string) error {
	DefaultRateLimitScheduleMutex.Lock()
	defer DefaultRateLimitScheduleMutex.Unlock()

	var schedule []dto.TimeRateLimit
	if err := json.Unmarshal([]byte(jsonStr), &schedule); err != nil {
		return fmt.Errorf("invalid rate limit schedule JSON: %w", err)
	}
	for i, rl := range schedule {
		if err := validateTimeRange(rl.TimeRange); err != nil {
			return fmt.Errorf("schedule[%d]: %w", i, err)
		}
		if rl.RPM < 0 || rl.TPM < 0 {
			return fmt.Errorf("schedule[%d]: rpm and tpm must be >= 0", i)
		}
	}
	DefaultRateLimitSchedule = schedule
	return nil
}

// CheckDefaultRateLimitSchedule 校验分时段限流配置 JSON 格式
func CheckDefaultRateLimitSchedule(jsonStr string) error {
	var schedule []dto.TimeRateLimit
	if err := json.Unmarshal([]byte(jsonStr), &schedule); err != nil {
		return fmt.Errorf("invalid JSON: %w", err)
	}
	for i, rl := range schedule {
		if err := validateTimeRange(rl.TimeRange); err != nil {
			return fmt.Errorf("schedule[%d]: %w", i, err)
		}
		if rl.RPM < 0 || rl.TPM < 0 {
			return fmt.Errorf("schedule[%d]: rpm and tpm must be >= 0", i)
		}
	}
	return nil
}

// GetCurrentTimeRateLimit 从分时段配置中获取当前时间对应的限流配置
func GetCurrentTimeRateLimit(schedule []dto.TimeRateLimit) *dto.TimeRateLimit {
	if len(schedule) == 0 {
		return nil
	}
	now := time.Now().Format("15:04")
	for _, rl := range schedule {
		start, end, err := parseTimeRange(rl.TimeRange)
		if err != nil {
			continue
		}
		if isTimeInRange(now, start, end) {
			return &rl
		}
	}
	return nil
}

// GetEffectiveRateLimitSchedule 获取有效的分时段限流配置（用户配置优先，否则全局默认）
func GetEffectiveRateLimitSchedule(userSchedule []dto.TimeRateLimit) []dto.TimeRateLimit {
	if len(userSchedule) > 0 {
		return userSchedule
	}
	DefaultRateLimitScheduleMutex.RLock()
	defer DefaultRateLimitScheduleMutex.RUnlock()
	return DefaultRateLimitSchedule
}

// parseTimeRange 解析 "HH:MM-HH:MM" 格式的时间段
func parseTimeRange(timeRange string) (start, end string, err error) {
	parts := strings.SplitN(timeRange, "-", 2)
	if len(parts) != 2 {
		return "", "", fmt.Errorf("invalid time range format: %q, expected HH:MM-HH:MM", timeRange)
	}
	start = strings.TrimSpace(parts[0])
	end = strings.TrimSpace(parts[1])
	return start, end, nil
}

// validateTimeRange 校验时间段格式
func validateTimeRange(timeRange string) error {
	start, end, err := parseTimeRange(timeRange)
	if err != nil {
		return err
	}
	if !isValidTime(start) || !isValidTime(end) {
		return fmt.Errorf("invalid time format in %q, expected HH:MM", timeRange)
	}
	return nil
}

// isValidTime 校验 HH:MM 格式
func isValidTime(t string) bool {
	if len(t) != 5 || t[2] != ':' {
		return false
	}
	_, err := time.Parse("15:04", t)
	return err == nil
}

// isTimeInRange 判断时间是否在范围内，支持跨天（如 22:00-08:00）
func isTimeInRange(now, start, end string) bool {
	if start <= end {
		// 同一天内：09:00-18:00
		return now >= start && now < end
	}
	// 跨天：22:00-08:00
	return now >= start || now < end
}
