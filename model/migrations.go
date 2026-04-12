package model

import (
	"fmt"

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

	// Add foreign keys for TopUp table
	if err := addForeignKey(DB, "topups", "user_id", "users", "id", "CASCADE", "CASCADE"); err != nil {
		common.SysError("failed to add foreign key for topups.user_id: " + err.Error())
	}

	// Add foreign keys for Redemption table
	if err := addForeignKey(DB, "redemptions", "user_id", "users", "id", "SET NULL", "CASCADE"); err != nil {
		common.SysError("failed to add foreign key for redemptions.user_id: " + err.Error())
	}

	// Add foreign keys for CheckIn table
	if err := addForeignKey(DB, "checkins", "user_id", "users", "id", "CASCADE", "CASCADE"); err != nil {
		common.SysError("failed to add foreign key for checkins.user_id: " + err.Error())
	}

	// Add foreign keys for Task table
	if err := addForeignKey(DB, "tasks", "user_id", "users", "id", "CASCADE", "CASCADE"); err != nil {
		common.SysError("failed to add foreign key for tasks.user_id: " + err.Error())
	}

	// Add foreign keys for Midjourney table
	if err := addForeignKey(DB, "midjourneys", "user_id", "users", "id", "CASCADE", "CASCADE"); err != nil {
		common.SysError("failed to add foreign key for midjourneys.user_id: " + err.Error())
	}

	// Add foreign keys for Subscription table
	if err := addForeignKey(DB, "subscriptions", "user_id", "users", "id", "CASCADE", "CASCADE"); err != nil {
		common.SysError("failed to add foreign key for subscriptions.user_id: " + err.Error())
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

// RunDatabaseOptimizations runs all database optimization tasks
func RunDatabaseOptimizations() {
	go func() {
		// Wait for database to be fully initialized
		if DB == nil {
			return
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
