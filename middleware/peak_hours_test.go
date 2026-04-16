package middleware

import (
	"testing"
	"time"

	"github.com/QuantumNous/new-api/common"
)

func resetPeakConfig() {
	common.PeakHoursEnabled = false
	common.PeakHoursStart = "09:00"
	common.PeakHoursEnd = "18:00"
	common.PeakHoursWeekdays = "1,2,3,4,5"
	common.PeakHoursRestrictedGroups = ""
}

func TestIsPeakHoursBlocked_Disabled(t *testing.T) {
	resetPeakConfig()
	common.PeakHoursEnabled = false
	common.PeakHoursRestrictedGroups = "default"
	if IsPeakHoursBlocked("default") {
		t.Fatal("should not block when disabled")
	}
}

func TestIsPeakHoursBlocked_UnrestrictedGroup(t *testing.T) {
	resetPeakConfig()
	common.PeakHoursEnabled = true
	common.PeakHoursRestrictedGroups = "internal"
	if IsPeakHoursBlocked("vip") {
		t.Fatal("should not block unrestricted group")
	}
}

func TestIsTimePeak_WorkHours(t *testing.T) {
	resetPeakConfig()
	// 周三 10:30 = 峰时
	wed := time.Date(2026, 4, 15, 10, 30, 0, 0, time.Local) // Wednesday
	common.PeakHoursStart = "09:00"
	common.PeakHoursEnd = "18:00"
	common.PeakHoursWeekdays = "1,2,3,4,5"
	if !isTimePeak(wed) {
		t.Fatal("Wed 10:30 should be peak")
	}
}

func TestIsTimePeak_Evening(t *testing.T) {
	resetPeakConfig()
	// 周三 20:00 = 非峰时
	wed := time.Date(2026, 4, 15, 20, 0, 0, 0, time.Local)
	common.PeakHoursStart = "09:00"
	common.PeakHoursEnd = "18:00"
	common.PeakHoursWeekdays = "1,2,3,4,5"
	if isTimePeak(wed) {
		t.Fatal("Wed 20:00 should not be peak")
	}
}

func TestIsTimePeak_Weekend(t *testing.T) {
	resetPeakConfig()
	// 周六 10:30 = 非峰时（周末不在限制日）
	sat := time.Date(2026, 4, 18, 10, 30, 0, 0, time.Local) // Saturday
	common.PeakHoursStart = "09:00"
	common.PeakHoursEnd = "18:00"
	common.PeakHoursWeekdays = "1,2,3,4,5"
	if isTimePeak(sat) {
		t.Fatal("Sat should not be peak (weekday 6 not in list)")
	}
}

func TestIsTimePeak_CrossMidnight(t *testing.T) {
	resetPeakConfig()
	// 跨午夜场景: 22:00-06:00
	common.PeakHoursStart = "22:00"
	common.PeakHoursEnd = "06:00"
	common.PeakHoursWeekdays = "1,2,3,4,5"

	// 周三 23:00 应为峰时
	late := time.Date(2026, 4, 15, 23, 0, 0, 0, time.Local)
	if !isTimePeak(late) {
		t.Fatal("23:00 should be peak in 22:00-06:00 range")
	}

	// 周三 03:00 应为峰时
	early := time.Date(2026, 4, 15, 3, 0, 0, 0, time.Local)
	if !isTimePeak(early) {
		t.Fatal("03:00 should be peak in 22:00-06:00 range")
	}

	// 周三 12:00 应为非峰时
	noon := time.Date(2026, 4, 15, 12, 0, 0, 0, time.Local)
	if isTimePeak(noon) {
		t.Fatal("12:00 should not be peak in 22:00-06:00 range")
	}
}

func TestParseHHMM(t *testing.T) {
	h, m := parseHHMM("09:30")
	if h != 9 || m != 30 {
		t.Fatalf("expected 9:30, got %d:%d", h, m)
	}
	h, m = parseHHMM("invalid")
	if h != 0 || m != 0 {
		t.Fatalf("expected 0:0 for invalid, got %d:%d", h, m)
	}
}
