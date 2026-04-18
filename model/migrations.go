package model

import (
	"fmt"
	"time"

	"github.com/QuantumNous/new-api/common"
	"gorm.io/gorm"
)

// AddForeignKeys adds foreign key constraints to improve data integrity
// Note: This should be run after the initial database setup
func AddForeignKeys() error {
	if common.UsingSQLite {
		// SQLite has limited foreign key support, skip for now
		common.SysLog("Skipping foreign key setup for SQLite")
		return nil
	}

	common.SysLog("Adding foreign key constraints...")

	// Add foreign keys for Token table
	if err := addForeignKey(DB, "tokens", "user_id", "users", "id", "CASCADE", "CASCADE"); err != nil {
		common.SysError("failed to add foreign key for tokens.user_id: " + err.Error())
	}

	// Add foreign keys for Log table (optional - logs should not be deleted when user is deleted)
	// if err := addForeignKey(LOG_DB, "logs", "user_id", "users", "id", "RESTRICT", "RESTRICT"); err != nil {
	// 	common.SysError("failed to add foreign key for logs.user_id: " + err.Error())
	// }

	// Add foreign keys for Ability table
	if err := addForeignKey(DB, "abilities", "channel_id", "channels", "id", "CASCADE", "CASCADE"); err != nil {
		common.SysError("failed to add foreign key for abilities.channel_id: " + err.Error())
	}

	// Add foreign keys for Task table
	if err := addForeignKey(DB, "tasks", "user_id", "users", "id", "CASCADE", "CASCADE"); err != nil {
		common.SysError("failed to add foreign key for tasks.user_id: " + err.Error())
	}

	common.SysLog("Foreign key constraints added successfully")
	return nil
}

// addForeignKey adds a foreign key constraint if it doesn't exist
func addForeignKey(db *gorm.DB, table, column, refTable, refColumn, onDelete, onUpdate string) error {
	// Check if foreign key already exists
	var count int64
	constraintName := fmt.Sprintf("fk_%s_%s", table, column)

	// Different queries for different databases
	switch db.Dialector.Name() {
	case "mysql":
		db.Raw("SELECT COUNT(*) FROM information_schema.TABLE_CONSTRAINTS WHERE CONSTRAINT_SCHEMA = DATABASE() AND TABLE_NAME = ? AND CONSTRAINT_NAME = ?", table, constraintName).Scan(&count)
	case "postgres":
		db.Raw("SELECT COUNT(*) FROM information_schema.TABLE_CONSTRAINTS WHERE TABLE_NAME = ? AND CONSTRAINT_NAME = ?", table, constraintName).Scan(&count)
	default:
		// For other databases, try to add and ignore errors
		count = 0
	}

	if count > 0 {
		common.SysLog(fmt.Sprintf("Foreign key %s already exists, skipping", constraintName))
		return nil
	}

	// Add foreign key constraint
	sql := fmt.Sprintf("ALTER TABLE %s ADD CONSTRAINT %s FOREIGN KEY (%s) REFERENCES %s(%s) ON DELETE %s ON UPDATE %s",
		table, constraintName, column, refTable, refColumn, onDelete, onUpdate)

	if err := db.Exec(sql).Error; err != nil {
		return err
	}

	common.SysLog(fmt.Sprintf("Added foreign key: %s", constraintName))
	return nil
}

// CreateIndexes creates additional indexes for better query performance
func CreateIndexes() error {
	common.SysLog("Creating additional indexes...")

	// Create composite index for logs table (common query pattern)
	if common.UsingMySQL {
		// MySQL specific optimizations
		DB.Exec("CREATE INDEX IF NOT EXISTS idx_logs_user_created ON logs(user_id, created_at)")
		DB.Exec("CREATE INDEX IF NOT EXISTS idx_logs_channel_created ON logs(channel_id, created_at)")
		DB.Exec("CREATE INDEX IF NOT EXISTS idx_logs_model_created ON logs(model_name, created_at)")
	}

	// Create index for token quota queries
	DB.Exec("CREATE INDEX IF NOT EXISTS idx_tokens_user_quota ON tokens(user_id, remain_quota)")

	// Create index for channel status queries
	DB.Exec("CREATE INDEX IF NOT EXISTS idx_channels_status_type ON channels(status, type)")

	common.SysLog("Additional indexes created successfully")
	return nil
}

// legacyTablesToDrop 是支付/订阅/兑换/midjourney/checkin 等已瘦身功能遗留的表。
// 这些表在老部署里会占空间且可能有过期外键，但 GORM AutoMigrate 不会主动删。
// dropLegacyTables 会一次性 DROP，并在 options 表写入 marker 防止重复执行。
var legacyTablesToDrop = []string{
	"subscription_plans",
	"subscription_orders",
	"user_subscriptions",
	"subscription_pre_consume_records",
	"top_ups",
	"topups",
	"redemptions",
	"midjourneys",
	"checkins",
}

const legacyDropMarkerKey = "_legacy_tables_dropped_at"

// dropLegacyTables 幂等地移除历史瘦身遗留表；marker 由 options 表记录。
// SQLite 也适用（CREATE/DROP TABLE IF EXISTS 通用）。
func dropLegacyTables() error {
	if DB == nil {
		return nil
	}
	// 检查 marker：已跑过则 skip
	var existing Option
	if err := DB.Where("`key` = ?", legacyDropMarkerKey).First(&existing).Error; err == nil && existing.Value != "" {
		return nil
	}

	dropped := make([]string, 0, len(legacyTablesToDrop))
	for _, tbl := range legacyTablesToDrop {
		// 安全：tbl 来自硬编码白名单，不是用户输入
		if !DB.Migrator().HasTable(tbl) {
			continue
		}
		if err := DB.Migrator().DropTable(tbl); err != nil {
			common.SysError(fmt.Sprintf("dropLegacyTables: failed to drop %s: %v", tbl, err))
			continue
		}
		dropped = append(dropped, tbl)
	}

	// 写 marker 防重（即使本次无表可删，也落一次防后续重试）
	if err := DB.Save(&Option{Key: legacyDropMarkerKey, Value: fmt.Sprintf("%d", time.Now().Unix())}).Error; err != nil {
		common.SysError("dropLegacyTables: failed to write marker: " + err.Error())
	}

	if len(dropped) > 0 {
		common.SysLog(fmt.Sprintf("dropLegacyTables: dropped %d legacy tables: %v", len(dropped), dropped))
	}
	return nil
}

// RunDatabaseOptimizations runs all database optimization tasks
func RunDatabaseOptimizations() {
	go func() {
		// Wait for database to be fully initialized
		if DB == nil {
			return
		}

		// 清理瘦身遗留的废弃表（订阅 / 充值 / 兑换 / midjourney / checkin）
		if err := dropLegacyTables(); err != nil {
			common.SysError("Failed to drop legacy tables: " + err.Error())
		}

		// Add foreign keys
		if err := AddForeignKeys(); err != nil {
			common.SysError("Failed to add foreign keys: " + err.Error())
		}

		// Create additional indexes
		if err := CreateIndexes(); err != nil {
			common.SysError("Failed to create indexes: " + err.Error())
		}
	}()
}
