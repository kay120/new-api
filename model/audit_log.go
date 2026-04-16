package model

import (
	"strings"

	"github.com/QuantumNous/new-api/common"
)

// AuditLog 管理员操作审计日志
type AuditLog struct {
	Id         int    `json:"id"`
	UserId     int    `json:"user_id" gorm:"index"`             // 操作者用户 ID
	Username   string `json:"username" gorm:"type:varchar(64)"` // 操作者用户名
	Action     string `json:"action" gorm:"type:varchar(64);index"` // 操作类型: create_channel, update_user, delete_token 等
	Resource   string `json:"resource" gorm:"type:varchar(64)"` // 资源类型: channel, user, token, group, option
	ResourceId string `json:"resource_id" gorm:"type:varchar(64)"` // 资源 ID
	Detail     string `json:"detail" gorm:"type:text"`          // 操作详情（JSON）
	IP         string `json:"ip" gorm:"type:varchar(64)"`       // 操作者 IP
	CreatedAt  int64  `json:"created_at" gorm:"bigint;index"`
}

// RecordAuditLog 记录一条审计日志（异步）
func RecordAuditLog(userId int, username, action, resource, resourceId, detail, ip string) {
	go func() {
		log := &AuditLog{
			UserId:     userId,
			Username:   username,
			Action:     action,
			Resource:   resource,
			ResourceId: resourceId,
			Detail:     detail,
			IP:         ip,
			CreatedAt:  common.GetTimestamp(),
		}
		if err := DB.Create(log).Error; err != nil {
			common.SysError("failed to record audit log: " + err.Error())
		}
	}()
}

// QueryAuditLogs 查询审计日志（分页）
func QueryAuditLogs(page, pageSize int, filters map[string]string) ([]AuditLog, int64, error) {
	var logs []AuditLog
	var total int64
	q := DB.Model(&AuditLog{})
	if v, ok := filters["username"]; ok && v != "" {
		q = q.Where("username LIKE ?", "%"+strings.TrimSpace(v)+"%")
	}
	if v, ok := filters["action"]; ok && v != "" {
		q = q.Where("action = ?", v)
	}
	if v, ok := filters["resource"]; ok && v != "" {
		q = q.Where("resource = ?", v)
	}
	q.Count(&total)
	err := q.Order("id DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&logs).Error
	return logs, total, err
}
