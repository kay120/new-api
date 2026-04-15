package service

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/setting/system_setting"
)

func withSSRFDisabled(t *testing.T) {
	t.Helper()
	fs := system_setting.GetFetchSetting()
	orig := fs.EnableSSRFProtection
	fs.EnableSSRFProtection = false
	t.Cleanup(func() {
		fs.EnableSSRFProtection = orig
	})
}

func resetChannelAlertConfig() {
	common.ChannelAlertWebhookEnabled = false
	common.ChannelAlertWebhookURL = ""
	common.ChannelAlertWebhookSecret = ""
}

func TestSendChannelAlert_Disabled_NoOp(t *testing.T) {
	resetChannelAlertConfig()
	defer resetChannelAlertConfig()

	called := false
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
	}))
	defer srv.Close()

	common.ChannelAlertWebhookEnabled = false
	common.ChannelAlertWebhookURL = srv.URL

	SendChannelAlert("disabled", 1, "test-channel", "quota exhausted")

	time.Sleep(50 * time.Millisecond)
	if called {
		t.Fatal("webhook should not be called when disabled")
	}
}

func TestSendChannelAlert_EmptyURL_NoOp(t *testing.T) {
	resetChannelAlertConfig()
	defer resetChannelAlertConfig()

	common.ChannelAlertWebhookEnabled = true
	common.ChannelAlertWebhookURL = ""

	// Should not panic or error
	SendChannelAlert("disabled", 1, "test", "reason")
}

func TestSendChannelAlert_PostsPayload(t *testing.T) {
	resetChannelAlertConfig()
	defer resetChannelAlertConfig()
	withSSRFDisabled(t)
	InitHttpClient()

	var (
		mu       sync.Mutex
		received string
		done     = make(chan struct{}, 1)
	)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		b, _ := io.ReadAll(r.Body)
		mu.Lock()
		received = string(b)
		mu.Unlock()
		w.WriteHeader(http.StatusOK)
		done <- struct{}{}
	}))
	defer srv.Close()

	common.ChannelAlertWebhookEnabled = true
	common.ChannelAlertWebhookURL = srv.URL
	common.ChannelAlertWebhookSecret = ""

	SendChannelAlert("disabled", 42, "openai-main", "401 invalid_api_key")

	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("webhook was not called within timeout")
	}

	mu.Lock()
	payload := received
	mu.Unlock()

	if !strings.Contains(payload, "openai-main") {
		t.Fatalf("payload missing channel name: %s", payload)
	}
	if !strings.Contains(payload, "channel_disabled") {
		t.Fatalf("payload missing event type: %s", payload)
	}
	if !strings.Contains(payload, "401 invalid_api_key") {
		t.Fatalf("payload missing reason: %s", payload)
	}
}
