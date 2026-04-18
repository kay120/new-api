package userctl

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"testing"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/model"
	"github.com/gin-gonic/gin"
	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

// setupUserAllowedChannelsDB 准备 User + Channel 表的测试库。
func setupUserAllowedChannelsDB(t *testing.T) *gorm.DB {
	t.Helper()
	gin.SetMode(gin.TestMode)
	common.UsingSQLite = true
	common.UsingMySQL = false
	common.UsingPostgreSQL = false
	common.RedisEnabled = false
	model.InitColumnNamesForTesting()

	dsn := fmt.Sprintf("file:%s?mode=memory&cache=shared", strings.ReplaceAll(t.Name(), "/", "_"))
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{})
	require.NoError(t, err)
	model.DB = db
	model.LOG_DB = db

	require.NoError(t, db.AutoMigrate(&model.User{}, &model.Channel{}, &model.Ability{}))
	t.Cleanup(func() {
		if sqlDB, err := db.DB(); err == nil {
			_ = sqlDB.Close()
		}
	})
	return db
}

func seedUserWithAllowedChannels(t *testing.T, db *gorm.DB, id int, allowed string) {
	t.Helper()
	u := &model.User{
		Id:              id,
		Username:        "u",
		Role:            common.RoleCommonUser,
		Status:          common.UserStatusEnabled,
		Group:           "default",
		AllowedChannels: allowed,
	}
	require.NoError(t, db.Create(u).Error)
}

func seedEnabledChannel(t *testing.T, db *gorm.DB, id int, status int, models string) {
	t.Helper()
	ch := &model.Channel{
		Id:     id,
		Type:   1,
		Name:   fmt.Sprintf("ch_%d", id),
		Key:    "sk-x",
		Status: status,
		Models: models,
		Group:  "default",
	}
	require.NoError(t, db.Create(ch).Error)
}

type userModelsResponse struct {
	Success bool     `json:"success"`
	Message string   `json:"message"`
	Data    []string `json:"data"`
}

// ===========================================================================
// GetUserModels with allowed_channels
// ===========================================================================

func TestGetUserModels_AllowedChannelsWildcard(t *testing.T) {
	db := setupUserAllowedChannelsDB(t)
	seedUserWithAllowedChannels(t, db, 1, "11:*")
	seedEnabledChannel(t, db, 11, 1, "gpt-4,gpt-3.5,claude-3")

	ctx, rec := newAuthenticatedContext(t, http.MethodGet, "/api/user/models", nil, 1)
	GetUserModels(ctx)

	var resp userModelsResponse
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
	require.True(t, resp.Success, "msg=%s", resp.Message)
	assert.ElementsMatch(t, []string{"gpt-4", "gpt-3.5", "claude-3"}, resp.Data,
		"通配符应返回该渠道全部 models")
}

func TestGetUserModels_AllowedChannelsExactMatch(t *testing.T) {
	db := setupUserAllowedChannelsDB(t)
	seedUserWithAllowedChannels(t, db, 1, "12:gpt-4")
	seedEnabledChannel(t, db, 12, 1, "gpt-4,gpt-3.5,claude-3")

	ctx, rec := newAuthenticatedContext(t, http.MethodGet, "/api/user/models", nil, 1)
	GetUserModels(ctx)

	var resp userModelsResponse
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
	require.True(t, resp.Success)
	assert.Equal(t, []string{"gpt-4"}, resp.Data,
		"精确匹配只应返回显式列出的 model")
}

func TestGetUserModels_AllowedChannelsSkipsDisabled(t *testing.T) {
	db := setupUserAllowedChannelsDB(t)
	seedUserWithAllowedChannels(t, db, 1, "13:*,14:*")
	seedEnabledChannel(t, db, 13, 2, "gpt-4") // 禁用
	seedEnabledChannel(t, db, 14, 1, "claude-3") // 启用

	ctx, rec := newAuthenticatedContext(t, http.MethodGet, "/api/user/models", nil, 1)
	GetUserModels(ctx)

	var resp userModelsResponse
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
	require.True(t, resp.Success)
	assert.Equal(t, []string{"claude-3"}, resp.Data,
		"被禁用的渠道不应贡献 models")
}

func TestGetUserModels_MalformedEntriesIgnored(t *testing.T) {
	db := setupUserAllowedChannelsDB(t)
	// 混合合法 / 缺冒号 / 非数字 ID / 冒号为空
	seedUserWithAllowedChannels(t, db, 1, "15:gpt-4,noise,abc:gpt-4,:gpt-4")
	seedEnabledChannel(t, db, 15, 1, "gpt-4,gpt-3.5")

	ctx, rec := newAuthenticatedContext(t, http.MethodGet, "/api/user/models", nil, 1)
	GetUserModels(ctx)

	var resp userModelsResponse
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
	require.True(t, resp.Success)
	assert.Equal(t, []string{"gpt-4"}, resp.Data,
		"非法条目必须被静默跳过，仅解析合法的 15:gpt-4")
}

func TestGetUserModels_DedupesAcrossChannels(t *testing.T) {
	db := setupUserAllowedChannelsDB(t)
	seedUserWithAllowedChannels(t, db, 1, "16:*,17:*")
	seedEnabledChannel(t, db, 16, 1, "gpt-4,gpt-3.5")
	seedEnabledChannel(t, db, 17, 1, "gpt-4,claude-3")

	ctx, rec := newAuthenticatedContext(t, http.MethodGet, "/api/user/models", nil, 1)
	GetUserModels(ctx)

	var resp userModelsResponse
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
	require.True(t, resp.Success)
	assert.ElementsMatch(t, []string{"gpt-4", "gpt-3.5", "claude-3"}, resp.Data,
		"gpt-4 在两个渠道都存在，应去重")
}
