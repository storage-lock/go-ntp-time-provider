package ntp_time_provider

import (
	"context"
	"github.com/beevik/ntp"
	"github.com/storage-lock/go-events"
	"github.com/storage-lock/go-storage"
	"time"
)

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

// GetTime 从NTP获取时间，当不方便从Storage获取时间的时候可以使用NTP作为时间源
func (x *NTPTimeProvider) GetTime(ctx context.Context) (time.Time, error) {

	var lastError error
	for _, server := range x.ntpServers {

		now, err := ntp.Time(server)

		if err != nil {
			if x.e != nil {
				x.e.AddAction(events.NewAction(ActionNtpError).SetErr(err)).Publish(ctx)
			}
			lastError = err
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

	if lastError != nil {
		return time.Time{}, ErrTimeProviderUnavailable
	} else {
		return time.Time{}, ErrTimeProviderUnavailable
	}
}

// ------------------------------------------------- --------------------------------------------------------------------
