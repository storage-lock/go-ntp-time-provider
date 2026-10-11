package ntp_time_provider

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/storage-lock/go-events"
)

func TestNTPProviderBranches(t *testing.T) {
	old := DefaultNtpServers
	SetDefaultNtpServers([]string{"default.example"})
	t.Cleanup(func() { SetDefaultNtpServers(old) })
	provider := NewNTPTimeProvider(nil)
	if len(provider.ntpServers) != 1 || provider.ntpServers[0] != "default.example" {
		t.Fatal(provider.ntpServers)
	}
	if got := GetDefaultNtpServers(); len(got) != 1 {
		t.Fatal(got)
	}

	oldQuery := queryTime
	t.Cleanup(func() { queryTime = oldQuery })
	event := events.NewEvent("ntp")
	provider = NewNTPTimeProvider(nil, "bad.example", "zero.example", "good.example").SetEvent(event)
	if provider.e != event {
		t.Fatal("event")
	}
	calls := 0
	queryTime = func(server string) (time.Time, error) {
		calls++
		switch server {
		case "bad.example":
			return time.Time{}, errors.New("query failed")
		case "zero.example":
			return time.Time{}, nil
		default:
			return time.Unix(100, 0), nil
		}
	}
	got, err := provider.GetTime(context.Background())
	if err != nil || got.Unix() != 100 || calls != 3 {
		t.Fatalf("time=%v err=%v calls=%d", got, err, calls)
	}

	queryTime = func(string) (time.Time, error) { return time.Time{}, errors.New("down") }
	if _, err := provider.GetTime(context.Background()); err != ErrTimeProviderUnavailable {
		t.Fatal(err)
	}
	canceled, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := provider.GetTime(canceled); err != context.Canceled {
		t.Fatal(err)
	}
	queryTime = func(string) (time.Time, error) {
		time.Sleep(200 * time.Millisecond)
		return time.Unix(1, 0), nil
	}
	short, stop := context.WithTimeout(context.Background(), 20*time.Millisecond)
	defer stop()
	if _, err := provider.GetTime(short); err == nil {
		t.Fatal("query ignored deadline")
	}
}
