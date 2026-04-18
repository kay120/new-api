package service

import (
	"errors"
	"net/http/httptest"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/model"
	relaycommon "github.com/QuantumNous/new-api/relay/common"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// fakeFunding 是 FundingSource 的可观测桩，用于隔离单元测试。
type fakeFunding struct {
	sourceName string

	preConsumeErr error
	settleErr     error
	refundErr     error

	preConsumeCalls atomic.Int32
	settleCalls     atomic.Int32
	refundCalls     atomic.Int32

	lastPreConsumeAmt atomic.Int64
	lastSettleDelta   atomic.Int64

	refundWG sync.WaitGroup
}

func newFakeFunding(name string) *fakeFunding {
	return &fakeFunding{sourceName: name}
}

func (f *fakeFunding) Source() string { return f.sourceName }

func (f *fakeFunding) PreConsume(amount int) error {
	f.preConsumeCalls.Add(1)
	f.lastPreConsumeAmt.Store(int64(amount))
	return f.preConsumeErr
}

func (f *fakeFunding) Settle(delta int) error {
	f.settleCalls.Add(1)
	f.lastSettleDelta.Store(int64(delta))
	return f.settleErr
}

func (f *fakeFunding) Refund() error {
	defer f.refundWG.Done()
	f.refundCalls.Add(1)
	return f.refundErr
}

// expectRefund 标记即将有一次 Refund 调用，Refund() 完成后 refundWG 会递减。
func (f *fakeFunding) expectRefund() {
	f.refundWG.Add(1)
}

// waitRefund 等待所有已排队的 Refund 调用完成。
func (f *fakeFunding) waitRefund(t *testing.T, timeout time.Duration) {
	t.Helper()
	done := make(chan struct{})
	go func() {
		f.refundWG.Wait()
		close(done)
	}()
	select {
	case <-done:
	case <-time.After(timeout):
		t.Fatalf("timeout waiting for Refund goroutine")
	}
}

func newTestGinContext() *gin.Context {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	return c
}

func newRelayInfo(userId, tokenId int, tokenKey string) *relaycommon.RelayInfo {
	return &relaycommon.RelayInfo{
		UserId:   userId,
		TokenId:  tokenId,
		TokenKey: tokenKey,
	}
}

// ===========================================================================
// NewBillingSession 构造
// ===========================================================================

func TestNewBillingSession_NilRelayInfo(t *testing.T) {
	_, apiErr := NewBillingSession(nil, nil, 0)
	require.NotNil(t, apiErr)
}

func TestNewBillingSession_SelfUseMode(t *testing.T) {
	info := newRelayInfo(1, 1, "sk-x")
	session, apiErr := NewBillingSession(nil, info, 0)
	require.Nil(t, apiErr)
	require.NotNil(t, session)
	assert.Nil(t, session.funding, "自用模式下 funding 应为 nil")
	assert.Equal(t, 0, session.GetPreConsumedQuota())
}

// ===========================================================================
// Settle
// ===========================================================================

func TestSettle_SelfUseMode_NoFunding(t *testing.T) {
	session, _ := NewBillingSession(nil, newRelayInfo(1, 1, "sk-x"), 0)
	require.NoError(t, session.Settle(12345))
	assert.True(t, session.settled)
}

func TestSettle_Idempotent(t *testing.T) {
	session, _ := NewBillingSession(nil, newRelayInfo(1, 1, "sk-x"), 0)
	require.NoError(t, session.Settle(100))
	// 第二次调用应直接返回 nil，不重复动作
	require.NoError(t, session.Settle(200))
}

func TestSettle_WalletFunding_ZeroDelta(t *testing.T) {
	info := newRelayInfo(1, 1, "sk-x")
	info.IsPlayground = true // 跳过 token 调整
	ff := newFakeFunding(BillingSourceWallet)
	session := &BillingSession{relayInfo: info, funding: ff, preConsumedQuota: 500}

	require.NoError(t, session.Settle(500))
	assert.Equal(t, int32(0), ff.settleCalls.Load(), "delta=0 时不应调用 funding.Settle")
	assert.True(t, session.settled)
}

func TestSettle_WalletFunding_PositiveDelta(t *testing.T) {
	truncate(t)
	const userID, tokenID = 101, 101
	seedUser(t, userID, 10000)
	seedToken(t, tokenID, userID, "sk-settle-pos", 8000)

	info := newRelayInfo(userID, tokenID, "sk-settle-pos")
	ff := newFakeFunding(BillingSourceWallet)
	session := &BillingSession{relayInfo: info, funding: ff, preConsumedQuota: 1000}

	require.NoError(t, session.Settle(1500))
	assert.Equal(t, int32(1), ff.settleCalls.Load())
	assert.EqualValues(t, 500, ff.lastSettleDelta.Load(), "补扣 delta 应为 500")
	assert.Equal(t, 8000-500, getTokenRemainQuota(t, tokenID), "令牌应多扣 500")
}

func TestSettle_WalletFunding_NegativeDelta(t *testing.T) {
	truncate(t)
	const userID, tokenID = 102, 102
	seedUser(t, userID, 10000)
	seedToken(t, tokenID, userID, "sk-settle-neg", 8000)

	info := newRelayInfo(userID, tokenID, "sk-settle-neg")
	ff := newFakeFunding(BillingSourceWallet)
	session := &BillingSession{relayInfo: info, funding: ff, preConsumedQuota: 2000}

	require.NoError(t, session.Settle(500))
	assert.EqualValues(t, -1500, ff.lastSettleDelta.Load())
	assert.Equal(t, 8000+1500, getTokenRemainQuota(t, tokenID), "令牌应退还 1500")
}

func TestSettle_FundingFailsLeavesNotSettled(t *testing.T) {
	info := newRelayInfo(1, 1, "sk-x")
	info.IsPlayground = true
	ff := newFakeFunding(BillingSourceWallet)
	ff.settleErr = errors.New("db timeout")
	session := &BillingSession{relayInfo: info, funding: ff, preConsumedQuota: 100}

	err := session.Settle(300)
	require.Error(t, err)
	assert.False(t, session.settled)
	assert.False(t, session.fundingSettled)
}

// ===========================================================================
// Refund
// ===========================================================================

func TestRefund_SelfUseMode_NoOp(t *testing.T) {
	c := newTestGinContext()
	session, _ := NewBillingSession(nil, newRelayInfo(1, 1, "sk-x"), 0)
	session.Refund(c)
	assert.False(t, session.refunded, "自用模式 Refund 应直接返回，不标记 refunded")
}

func TestRefund_AfterSettle_NoOp(t *testing.T) {
	c := newTestGinContext()
	ff := newFakeFunding(BillingSourceWallet)
	session := &BillingSession{
		relayInfo: newRelayInfo(1, 1, "sk-x"),
		funding:   ff,
		settled:   true,
	}
	session.Refund(c)
	assert.Equal(t, int32(0), ff.refundCalls.Load())
}

func TestRefund_NoTokenConsumed_NoOp(t *testing.T) {
	c := newTestGinContext()
	ff := newFakeFunding(BillingSourceWallet)
	session := &BillingSession{
		relayInfo:     newRelayInfo(1, 1, "sk-x"),
		funding:       ff,
		tokenConsumed: 0,
	}
	session.Refund(c)
	assert.Equal(t, int32(0), ff.refundCalls.Load())
	assert.False(t, session.refunded)
}

func TestRefund_HappyPath(t *testing.T) {
	truncate(t)
	c := newTestGinContext()
	const userID, tokenID = 201, 201
	seedUser(t, userID, 10000)
	seedToken(t, tokenID, userID, "sk-refund", 5000)

	ff := newFakeFunding(BillingSourceWallet)
	ff.expectRefund()
	session := &BillingSession{
		relayInfo:     newRelayInfo(userID, tokenID, "sk-refund"),
		funding:       ff,
		tokenConsumed: 300,
	}

	session.Refund(c)
	assert.True(t, session.refunded)

	ff.waitRefund(t, 2*time.Second)
	assert.Equal(t, int32(1), ff.refundCalls.Load())

	// Refund 的 token 调整在 funding.Refund 之后异步进行，poll 到值变化
	waitUntil(t, 2*time.Second, func() bool {
		return getTokenRemainQuota(t, tokenID) == 5000+300
	}, "异步退还 300 token 额度")
}

// waitUntil 每 10ms 检查一次 pred，直到返回 true 或超时。
func waitUntil(t *testing.T, timeout time.Duration, pred func() bool, msg string) {
	t.Helper()
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if pred() {
			return
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Fatalf("timeout waiting for condition: %s", msg)
}

func TestRefund_Idempotent(t *testing.T) {
	c := newTestGinContext()
	ff := newFakeFunding(BillingSourceWallet)
	ff.expectRefund()
	session := &BillingSession{
		relayInfo:     newRelayInfo(1, 0, "sk-x"), // tokenId=0 跳过 token 退还
		funding:       ff,
		tokenConsumed: 100,
	}
	session.Refund(c)
	session.Refund(c) // 第二次应被 refunded flag 拦截
	session.Refund(c)

	ff.waitRefund(t, 2*time.Second)
	assert.Equal(t, int32(1), ff.refundCalls.Load(), "Refund 必须幂等")
}

func TestNeedsRefund_FundingSettledBlocksRefund(t *testing.T) {
	session := &BillingSession{
		relayInfo:      newRelayInfo(1, 1, "sk-x"),
		funding:        newFakeFunding(BillingSourceWallet),
		tokenConsumed:  100,
		fundingSettled: true,
	}
	assert.False(t, session.NeedsRefund(),
		"fundingSettled 后不应再触发退款，防止资金来源被重复修改")
}

// ===========================================================================
// shouldTrust
// ===========================================================================

func TestShouldTrust_ForcePreConsumeDisables(t *testing.T) {
	info := newRelayInfo(1, 1, "sk-x")
	info.ForcePreConsume = true
	info.TokenUnlimited = true
	info.UserQuota = 999999999
	session := &BillingSession{relayInfo: info, funding: newFakeFunding(BillingSourceWallet)}

	c := newTestGinContext()
	assert.False(t, session.shouldTrust(c), "ForcePreConsume=true 必须关闭信任旁路")
}

func TestShouldTrust_TokenInsufficient(t *testing.T) {
	info := newRelayInfo(1, 1, "sk-x")
	info.UserQuota = 9999999999 // 远超 trust quota
	session := &BillingSession{relayInfo: info, funding: newFakeFunding(BillingSourceWallet)}

	c := newTestGinContext()
	c.Set("token_quota", 1) // 远小于 trust quota
	assert.False(t, session.shouldTrust(c))
}

func TestShouldTrust_WalletInsufficient(t *testing.T) {
	info := newRelayInfo(1, 1, "sk-x")
	info.UserQuota = 1 // 远小于 trust quota
	info.TokenUnlimited = true
	session := &BillingSession{relayInfo: info, funding: newFakeFunding(BillingSourceWallet)}

	c := newTestGinContext()
	assert.False(t, session.shouldTrust(c))
}

func TestShouldTrust_AllSufficient(t *testing.T) {
	trust := common.GetTrustQuota()
	info := newRelayInfo(1, 1, "sk-x")
	info.TokenUnlimited = true
	info.UserQuota = trust + 1
	session := &BillingSession{relayInfo: info, funding: newFakeFunding(BillingSourceWallet)}

	c := newTestGinContext()
	assert.True(t, session.shouldTrust(c))
}

func TestShouldTrust_TokenUnlimitedBypassesTokenQuota(t *testing.T) {
	trust := common.GetTrustQuota()
	info := newRelayInfo(1, 1, "sk-x")
	info.TokenUnlimited = true
	info.UserQuota = trust + 1
	session := &BillingSession{relayInfo: info, funding: newFakeFunding(BillingSourceWallet)}

	c := newTestGinContext()
	c.Set("token_quota", 0) // 会被 TokenUnlimited 旁路忽略
	assert.True(t, session.shouldTrust(c))
}

// ===========================================================================
// preConsume
// ===========================================================================

func TestPreConsume_WalletSuccess_UpdatesRelayInfoFields(t *testing.T) {
	truncate(t)
	const userID, tokenID = 301, 301
	seedUser(t, userID, 10000)
	seedToken(t, tokenID, userID, "sk-preconsume", 5000)

	info := newRelayInfo(userID, tokenID, "sk-preconsume")
	ff := newFakeFunding(BillingSourceWallet)
	session := &BillingSession{relayInfo: info, funding: ff}

	c := newTestGinContext()
	c.Set("token_quota", 0)
	apiErr := session.preConsume(c, 1000)
	require.Nil(t, apiErr)

	assert.Equal(t, 1000, session.preConsumedQuota)
	assert.Equal(t, 1000, session.tokenConsumed)
	assert.EqualValues(t, 1000, ff.lastPreConsumeAmt.Load())
	assert.Equal(t, BillingSourceWallet, info.BillingSource, "preConsume 成功后应同步 BillingSource")
	assert.Equal(t, 1000, info.FinalPreConsumedQuota)
	assert.Equal(t, 5000-1000, getTokenRemainQuota(t, tokenID))
}

func TestPreConsume_WalletFailureRollsBackToken(t *testing.T) {
	truncate(t)
	const userID, tokenID = 302, 302
	seedUser(t, userID, 10000)
	seedToken(t, tokenID, userID, "sk-preconsume-rollback", 5000)

	info := newRelayInfo(userID, tokenID, "sk-preconsume-rollback")
	ff := newFakeFunding(BillingSourceWallet)
	ff.preConsumeErr = errors.New("wallet exhausted")
	session := &BillingSession{relayInfo: info, funding: ff}

	c := newTestGinContext()
	c.Set("token_quota", 0)
	apiErr := session.preConsume(c, 1000)
	require.NotNil(t, apiErr, "预扣资金来源失败应返回 error")

	assert.Equal(t, 0, session.tokenConsumed, "令牌预扣必须被回滚")
	assert.Equal(t, 5000, getTokenRemainQuota(t, tokenID), "token_remain 应恢复到预扣前")
}

func TestPreConsume_TrustBypassSkipsAllDebits(t *testing.T) {
	truncate(t)
	trust := common.GetTrustQuota()
	const userID, tokenID = 303, 303
	seedUser(t, userID, trust*2)
	seedToken(t, tokenID, userID, "sk-trust", trust*2)

	info := newRelayInfo(userID, tokenID, "sk-trust")
	info.TokenUnlimited = true
	info.UserQuota = trust * 2
	ff := newFakeFunding(BillingSourceWallet)
	session := &BillingSession{relayInfo: info, funding: ff}

	c := newTestGinContext()
	apiErr := session.preConsume(c, 9999)
	require.Nil(t, apiErr)

	assert.Equal(t, 0, session.preConsumedQuota, "信任旁路下 preConsumedQuota 应为 0")
	assert.Equal(t, 0, session.tokenConsumed)
	assert.EqualValues(t, 0, ff.lastPreConsumeAmt.Load(), "信任旁路下 funding 预扣应为 0")
	assert.Equal(t, trust*2, getTokenRemainQuota(t, tokenID), "令牌不应被动扣")
}

func TestPreConsume_ZeroQuotaIsNoop(t *testing.T) {
	truncate(t)
	const userID, tokenID = 304, 304
	seedUser(t, userID, 10000)
	seedToken(t, tokenID, userID, "sk-zero", 5000)

	info := newRelayInfo(userID, tokenID, "sk-zero")
	ff := newFakeFunding(BillingSourceWallet)
	session := &BillingSession{relayInfo: info, funding: ff}

	c := newTestGinContext()
	c.Set("token_quota", 0)
	apiErr := session.preConsume(c, 0)
	require.Nil(t, apiErr)

	assert.Equal(t, 0, session.tokenConsumed)
	assert.EqualValues(t, 0, ff.lastPreConsumeAmt.Load())
	assert.Equal(t, 1, int(ff.preConsumeCalls.Load()), "funding.PreConsume 仍应被调用（金额=0）")
}

// ===========================================================================
// GetPreConsumedQuota
// ===========================================================================

func TestGetPreConsumedQuota(t *testing.T) {
	session := &BillingSession{preConsumedQuota: 42}
	assert.Equal(t, 42, session.GetPreConsumedQuota())
}

// ensure imports used
var _ = model.DB
