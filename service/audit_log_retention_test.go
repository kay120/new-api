package service

import (
	"os"
	"testing"
	"time"

	"github.com/QuantumNous/new-api/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func seedAuditLog(t *testing.T, id int, createdAt int64) {
	t.Helper()
	row := &model.AuditLog{
		Id:        id,
		UserId:    1,
		Username:  "admin",
		Action:    "test_action",
		Resource:  "test",
		CreatedAt: createdAt,
	}
	require.NoError(t, model.DB.Create(row).Error)
}

func TestAuditLogRetention_GetDays_DefaultWhenUnset(t *testing.T) {
	t.Setenv("AUDIT_LOG_RETENTION_DAYS", "")
	assert.Equal(t, defaultAuditLogRetentionDays, getAuditLogRetentionDays())
}

func TestAuditLogRetention_GetDays_CustomEnv(t *testing.T) {
	t.Setenv("AUDIT_LOG_RETENTION_DAYS", "30")
	assert.Equal(t, 30, getAuditLogRetentionDays())
}

func TestAuditLogRetention_GetDays_InvalidEnvFallsBack(t *testing.T) {
	t.Setenv("AUDIT_LOG_RETENTION_DAYS", "not-a-number")
	assert.Equal(t, defaultAuditLogRetentionDays, getAuditLogRetentionDays())
}

func TestAuditLogRetention_GetDays_ZeroMeansDisabled(t *testing.T) {
	t.Setenv("AUDIT_LOG_RETENTION_DAYS", "0")
	assert.Equal(t, 0, getAuditLogRetentionDays())
}

func TestPurgeAuditLogsOlderThan_DeletesOnlyOldRows(t *testing.T) {
	require.NoError(t, model.DB.AutoMigrate(&model.AuditLog{}))
	model.DB.Exec("DELETE FROM audit_logs")

	now := time.Now().Unix()
	seedAuditLog(t, 1001, now-100*24*3600) // 100 天前
	seedAuditLog(t, 1002, now-45*24*3600)  // 45 天前
	seedAuditLog(t, 1003, now-1*24*3600)   // 1 天前

	cutoff := now - 60*24*3600
	deleted, err := model.PurgeAuditLogsOlderThan(cutoff)
	require.NoError(t, err)
	assert.EqualValues(t, 1, deleted, "仅 100 天前那条应被删")

	var remaining int64
	require.NoError(t, model.DB.Model(&model.AuditLog{}).Count(&remaining).Error)
	assert.EqualValues(t, 2, remaining)
}

func TestRunAuditLogPurgeOnce_RespectsDisabled(t *testing.T) {
	t.Setenv("AUDIT_LOG_RETENTION_DAYS", "0")
	require.NoError(t, model.DB.AutoMigrate(&model.AuditLog{}))
	model.DB.Exec("DELETE FROM audit_logs")

	seedAuditLog(t, 2001, time.Now().Unix()-365*24*3600)

	runAuditLogPurgeOnce()

	var remaining int64
	require.NoError(t, model.DB.Model(&model.AuditLog{}).Count(&remaining).Error)
	assert.EqualValues(t, 1, remaining, "禁用时不应删除任何行")
}

func TestRunAuditLogPurgeOnce_DeletesPerConfig(t *testing.T) {
	_ = os.Setenv("AUDIT_LOG_RETENTION_DAYS", "10")
	t.Cleanup(func() { _ = os.Unsetenv("AUDIT_LOG_RETENTION_DAYS") })
	require.NoError(t, model.DB.AutoMigrate(&model.AuditLog{}))
	model.DB.Exec("DELETE FROM audit_logs")

	now := time.Now().Unix()
	seedAuditLog(t, 3001, now-20*24*3600) // 保留期外
	seedAuditLog(t, 3002, now-5*24*3600)  // 保留期内

	runAuditLogPurgeOnce()

	var remaining int64
	require.NoError(t, model.DB.Model(&model.AuditLog{}).Count(&remaining).Error)
	assert.EqualValues(t, 1, remaining, "保留期外的行应被清理")
}
