package ntp_time_provider

import (
	"context"
	"github.com/beevik/ntp"
	"github.com/storage-lock/go-events"
	"github.com/storage-lock/go-storage"
	"time"
)

var queryTime = ntp.Time

// NTPTimeProvider 基于NTP的时间源实现
type NTPTimeProvider struct {
	e          *events.Event
	ntpServers []string
}

var _ storage.TimeProvider = &NTPTimeProvider{}

// NewNTPTimeProvider 如果是在云环境内网的话，手动指定一个内网的ntp服务器速度会更快，云服务商一般都会提供内网的ntp服务器
func NewNTPTimeProvider(e *events.Event, ntpServers ...string) *NTPTimeProvider {
	if len(ntpServers) == 0 {
		ntpServers = DefaultNtpServers
	}
	return &NTPTimeProvider{
		e:          e,
		ntpServers: ntpServers,
	}
}

func (x *NTPTimeProvider) SetEvent(e *events.Event) *NTPTimeProvider {
	x.e = e
	return x
}

// ntpQueryResult 承载单次 ntp.Time 调用的结果，用于配合 ctx 做提前返回
type ntpQueryResult struct {
	now time.Time
	err error
}

// GetTime 从NTP获取时间，当不方便从Storage获取时间的时候可以使用NTP作为时间源
//
// 漏洞修复：原实现完全忽略传入的 ctx，依次查询 DefaultNtpServers（默认 9 个），
// 每个失败都要等 ntp 库自身的 5s 超时才切换下一个，最坏情况一次 GetTime 阻塞
// 近 45s，调用方即使传入带超时/取消的 ctx 也完全无法提前中断。现在每个 server
// 的查询放到独立 goroutine 执行，用 select 监听 ctx.Done()，一旦 ctx 被取消/超时
// 立即返回，不再傻等剩余服务器逐个超时。
func (x *NTPTimeProvider) GetTime(ctx context.Context) (time.Time, error) {

	for _, server := range x.ntpServers {

		select {
		case <-ctx.Done():
			return time.Time{}, ctx.Err()
		default:
		}

		resultCh := make(chan ntpQueryResult, 1)
		go func(server string) {
			now, err := queryTime(server)
			resultCh <- ntpQueryResult{now: now, err: err}
		}(server)

		var now time.Time
		var err error
		select {
		case <-ctx.Done():
			return time.Time{}, ctx.Err()
		case r := <-resultCh:
			now, err = r.now, r.err
		}

		if err != nil {
			if x.e != nil {
				x.e.AddAction(events.NewAction(ActionNtpError).SetErr(err)).Publish(ctx)
			}
			continue
		}

		if now.IsZero() {
			if x.e != nil {
				x.e.AddAction(events.NewAction(ActionNtpZero)).Publish(ctx)
			}
			continue
		}

		if x.e != nil {
			x.e.AddAction(events.NewAction(ActionNtpSuccess).AddPayload("time", now)).Publish(ctx)
		}

		return now, nil
	}

	return time.Time{}, ErrTimeProviderUnavailable
}

// ------------------------------------------------- --------------------------------------------------------------------
