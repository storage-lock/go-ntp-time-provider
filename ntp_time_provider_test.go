package ntp_time_provider

import (
	"context"
	"fmt"
	"github.com/storage-lock/go-events"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestNTPTimeProvider_GetTime(t *testing.T) {

	e := events.NewEvent("xxx").AddListeners(events.NewListenerWrapper("test-listener", func(ctx context.Context, e *events.Event) {
		fmt.Println(e.ToJsonString())
	}))

	// 默认列表里的公共 NTP 服务器不保证每一台当前都可达，逐个断言会让测试
	// 取决于测试机器的网络。这里验证的是：至少有一台能给出时间，并且一台都
	// 不可达时返回错误而不是零时间。
	provider := NewNTPTimeProvider(e.Fork(), DefaultNtpServers...)
	got, err := provider.GetTime(context.Background())
	assert.NoError(t, err)
	assert.False(t, got.IsZero())

	dead := NewNTPTimeProvider(e.Fork(), "192.0.2.1")
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	_, err = dead.GetTime(ctx)
	assert.Error(t, err)
}

// TestNTPTimeProvider_GetTime_RespectsContextCancellation 验证漏洞修复：
// GetTime 曾经完全忽略 ctx，遇到不可达的服务器要傻等库自身的默认超时（5s/个，
// 最坏情况下 9 个服务器约 45s）才会切换下一个。这里用 TEST-NET-1（RFC 5737，
// 保证路由黑洞，不依赖真实网络当下是否可达任何 NTP 服务器）模拟不可达场景，
// 验证 ctx 一旦超时，GetTime 应该马上返回，而不是等库自身的超时逻辑。
func TestNTPTimeProvider_GetTime_RespectsContextCancellation(t *testing.T) {
	provider := NewNTPTimeProvider(nil, "192.0.2.1", "192.0.2.2", "192.0.2.3")
	ctx, cancel := context.WithTimeout(context.Background(), 300*time.Millisecond)
	defer cancel()

	start := time.Now()
	_, err := provider.GetTime(ctx)
	elapsed := time.Since(start)

	assert.NotNil(t, err)
	assert.Less(t, elapsed, 2*time.Second, "ctx 超时后 GetTime 应很快返回，而不是等每个 server 的默认 5s 超时")
}
