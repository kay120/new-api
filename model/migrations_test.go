package model

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ===========================================================================
// dropLegacyTables
// ===========================================================================

func TestDropLegacyTables_DropsPresentTables(t *testing.T) {
	// 清理可能残留的 marker 和测试表
	DB.Exec("DELETE FROM options WHERE `key` = ?", legacyDropMarkerKey)
	for _, tbl := range legacyTablesToDrop {
		DB.Exec("DROP TABLE IF EXISTS `" + tbl + "`")
	}
	// 确保 options 表存在（TestMain 里未迁移 Option，这里补）
	require.NoError(t, DB.AutoMigrate(&Option{}))

	// 手工建一张 legacy 表模拟老部署数据
	require.NoError(t, DB.Exec(
		"CREATE TABLE `user_subscriptions` (`id` INTEGER PRIMARY KEY, `user_id` INTEGER)",
	).Error)
	require.NoError(t, DB.Exec(
		"CREATE TABLE `top_ups` (`id` INTEGER PRIMARY KEY)",
	).Error)

	require.True(t, DB.Migrator().HasTable("user_subscriptions"))
	require.True(t, DB.Migrator().HasTable("top_ups"))

	require.NoError(t, dropLegacyTables())

	assert.False(t, DB.Migrator().HasTable("user_subscriptions"), "user_subscriptions 应被删除")
	assert.False(t, DB.Migrator().HasTable("top_ups"), "top_ups 应被删除")

	// marker 应被写入
	var opt Option
	require.NoError(t, DB.Where("`key` = ?", legacyDropMarkerKey).First(&opt).Error)
	assert.NotEmpty(t, opt.Value, "应记录 unix 时间戳")
}

func TestDropLegacyTables_Idempotent(t *testing.T) {
	// 准备：已有 marker
	DB.Exec("DELETE FROM options WHERE `key` = ?", legacyDropMarkerKey)
	require.NoError(t, DB.AutoMigrate(&Option{}))
	require.NoError(t, DB.Save(&Option{Key: legacyDropMarkerKey, Value: "1700000000"}).Error)

	// 重新建一张 legacy 表
	DB.Exec("DROP TABLE IF EXISTS `subscription_plans`")
	require.NoError(t, DB.Exec(
		"CREATE TABLE `subscription_plans` (`id` INTEGER PRIMARY KEY)",
	).Error)

	require.NoError(t, dropLegacyTables())

	// marker 已存在 → skip，表仍然在
	assert.True(t, DB.Migrator().HasTable("subscription_plans"),
		"marker 存在时不应再次 drop")

	// 清理以免影响其他用例
	DB.Exec("DROP TABLE IF EXISTS `subscription_plans`")
}

func TestDropLegacyTables_NoTablesPresent(t *testing.T) {
	// 清空 marker 和所有 legacy 表
	DB.Exec("DELETE FROM options WHERE `key` = ?", legacyDropMarkerKey)
	for _, tbl := range legacyTablesToDrop {
		DB.Exec("DROP TABLE IF EXISTS `" + tbl + "`")
	}
	require.NoError(t, DB.AutoMigrate(&Option{}))

	// 无表也不应报错，marker 仍应被写入
	require.NoError(t, dropLegacyTables())

	var opt Option
	require.NoError(t, DB.Where("`key` = ?", legacyDropMarkerKey).First(&opt).Error)
	assert.NotEmpty(t, opt.Value)
}
