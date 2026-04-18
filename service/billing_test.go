package service

import (
	"errors"
	"sync/atomic"
	"testing"

	relaycommon "github.com/QuantumNous/new-api/relay/common"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// fakeBillingSettler 实现 relaycommon.BillingSettler，用于 SettleBilling 单元测试。
type fakeBillingSettler struct {
	preConsumed int
	settleErr   error
	settleCalls atomic.Int32
	lastActual  atomic.Int64
}

func (f *fakeBillingSettler) Settle(actualQuota int) error {
	f.settleCalls.Add(1)
	f.lastActual.Store(int64(actualQuota))
	return f.settleErr
}
func (f *fakeBillingSettler) Refund(_ *gin.Context)   {}
func (f *fakeBillingSettler) NeedsRefund() bool       { return false }
func (f *fakeBillingSettler) GetPreConsumedQuota() int { return f.preConsumed }

// ===========================================================================
// PreConsumeBilling
// ===========================================================================

func TestPreConsumeBilling_AttachesSessionToRelayInfo(t *testing.T) {
	info := newRelayInfo(1, 1, "sk-x")
	apiErr := PreConsumeBilling(newTestGinContext(), 0, info)
	require.Nil(t, apiErr)
	require.NotNil(t, info.Billing)
}

func TestPreConsumeBilling_NilRelayInfo(t *testing.T) {
	apiErr := PreConsumeBilling(newTestGinContext(), 0, nil)
	require.NotNil(t, apiErr, "relayInfo nil 应返回错误")
}

// ===========================================================================
// SettleBilling — 有 BillingSession
// ===========================================================================

func TestSettleBilling_WithSession_DeltaZero(t *testing.T) {
	c := newTestGinContext()
	fs := &fakeBillingSettler{preConsumed: 100}
	info := newRelayInfo(1, 1, "sk-x")
	info.Billing = fs

	require.NoError(t, SettleBilling(c, info, 100))
	assert.EqualValues(t, 1, fs.settleCalls.Load())
	assert.EqualValues(t, 100, fs.lastActual.Load())
}

func TestSettleBilling_WithSession_DeltaPositive(t *testing.T) {
	c := newTestGinContext()
	fs := &fakeBillingSettler{preConsumed: 100}
	info := newRelayInfo(1, 1, "sk-x")
	info.Billing = fs

	require.NoError(t, SettleBilling(c, info, 250))
	assert.EqualValues(t, 250, fs.lastActual.Load())
}

func TestSettleBilling_WithSession_DeltaNegative(t *testing.T) {
	c := newTestGinContext()
	fs := &fakeBillingSettler{preConsumed: 500}
	info := newRelayInfo(1, 1, "sk-x")
	info.Billing = fs

	require.NoError(t, SettleBilling(c, info, 200))
	assert.EqualValues(t, 200, fs.lastActual.Load())
}

func TestSettleBilling_WithSession_ActualZero_SkipsNotify(t *testing.T) {
	c := newTestGinContext()
	fs := &fakeBillingSettler{preConsumed: 0}
	info := newRelayInfo(1, 1, "sk-x")
	info.Billing = fs

	require.NoError(t, SettleBilling(c, info, 0))
	// 核心路径：Settle 被调用；0 消耗不触发告警异步任务
	assert.EqualValues(t, 1, fs.settleCalls.Load())
}

func TestSettleBilling_WithSession_SettleError(t *testing.T) {
	c := newTestGinContext()
	fs := &fakeBillingSettler{preConsumed: 100, settleErr: errors.New("db down")}
	info := newRelayInfo(1, 1, "sk-x")
	info.Billing = fs

	err := SettleBilling(c, info, 300)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "db down")
}

// ===========================================================================
// SettleBilling — 无 BillingSession 的回退路径
// ===========================================================================

func TestSettleBilling_NoSession_DeltaZero(t *testing.T) {
	c := newTestGinContext()
	info := newRelayInfo(1, 1, "sk-x")
	info.FinalPreConsumedQuota = 500

	// 无 session + delta=0 应直接 return nil，不走 PostConsumeQuota
	require.NoError(t, SettleBilling(c, info, 500))
}

func TestSettleBilling_NoSession_DeltaNonZero_CallsPostConsume(t *testing.T) {
	truncate(t)
	const userID, tokenID = 501, 501
	seedUser(t, userID, 10000)
	seedToken(t, tokenID, userID, "sk-settle-bill", 5000)

	c := newTestGinContext()
	info := &relaycommon.RelayInfo{
		UserId:                userID,
		TokenId:               tokenID,
		TokenKey:              "sk-settle-bill",
		FinalPreConsumedQuota: 300,
	}

	// delta = 500 - 300 = 200，应走 PostConsumeQuota 扣钱包 + token
	require.NoError(t, SettleBilling(c, info, 500))
	assert.Equal(t, 10000-200, getUserQuota(t, userID))
	assert.Equal(t, 5000-200, getTokenRemainQuota(t, tokenID))
}

func TestSettleBilling_NoSession_Refund(t *testing.T) {
	truncate(t)
	const userID, tokenID = 502, 502
	seedUser(t, userID, 10000)
	seedToken(t, tokenID, userID, "sk-settle-refund", 5000)

	c := newTestGinContext()
	info := &relaycommon.RelayInfo{
		UserId:                userID,
		TokenId:               tokenID,
		TokenKey:              "sk-settle-refund",
		FinalPreConsumedQuota: 500,
	}

	// delta = 200 - 500 = -300，PostConsumeQuota 退还
	require.NoError(t, SettleBilling(c, info, 200))
	assert.Equal(t, 10000+300, getUserQuota(t, userID))
	assert.Equal(t, 5000+300, getTokenRemainQuota(t, tokenID))
}
