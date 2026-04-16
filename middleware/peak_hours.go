package middleware

import (
	"strconv"
	"strings"
	"time"

	"github.com/QuantumNous/new-api/common"
)

func IsPeakHoursBlocked(group string) bool {
	if !common.PeakHoursEnabled {
		return false
	}
	if !isGroupRestricted(group) {
		return false
	}
	return isCurrentTimePeak()
}

func isGroupRestricted(group string) bool {
	restricted := common.PeakHoursRestrictedGroups
	if restricted == "" {
		return false
	}
	for _, g := range strings.Split(restricted, ",") {
		if strings.TrimSpace(g) == group {
			return true
		}
	}
	return false
}

func isCurrentTimePeak() bool {
	return isTimePeak(time.Now())
}

func isTimePeak(now time.Time) bool {
	weekday := int(now.Weekday())
	if weekday == 0 {
		weekday = 7 // Sunday = 7
	}

	if !isWeekdayRestricted(weekday) {
		return false
	}

	startH, startM := parseHHMM(common.PeakHoursStart)
	endH, endM := parseHHMM(common.PeakHoursEnd)

	startMin := startH*60 + startM
	endMin := endH*60 + endM
	nowMin := now.Hour()*60 + now.Minute()

	if startMin <= endMin {
		return nowMin >= startMin && nowMin < endMin
	}
	// 跨午夜: e.g. 22:00-06:00
	return nowMin >= startMin || nowMin < endMin
}

func isWeekdayRestricted(weekday int) bool {
	for _, d := range strings.Split(common.PeakHoursWeekdays, ",") {
		v, err := strconv.Atoi(strings.TrimSpace(d))
		if err == nil && v == weekday {
			return true
		}
	}
	return false
}

func parseHHMM(s string) (int, int) {
	parts := strings.SplitN(strings.TrimSpace(s), ":", 2)
	if len(parts) != 2 {
		return 0, 0
	}
	h, _ := strconv.Atoi(parts[0])
	m, _ := strconv.Atoi(parts[1])
	return h, m
}
