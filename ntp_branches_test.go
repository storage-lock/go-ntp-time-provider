package ntp_time_provider

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/storage-lock/go-events"
)

func TestNTPProviderBranches(t *testing.T) {
	old := GetDefaultNtpServers()
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
	var calls int
	var callsMu sync.Mutex
	queryTime = func(server string) (time.Time, error) {
		callsMu.Lock()
		calls++
		callsMu.Unlock()
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
	callsMu.Lock()
	gotCalls := calls
	callsMu.Unlock()
	if err != nil || got.Unix() != 100 || gotCalls != 3 {
		t.Fatalf("time=%v err=%v calls=%d", got, err, gotCalls)
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
	queryDone := make(chan struct{})
	queryTime = func(string) (time.Time, error) {
		defer close(queryDone)
		time.Sleep(200 * time.Millisecond)
		return time.Unix(1, 0), nil
	}
	short, stop := context.WithTimeout(context.Background(), 20*time.Millisecond)
	defer stop()
	if _, err := provider.GetTime(short); err == nil {
		t.Fatal("query ignored deadline")
	}
	select {
	case <-queryDone:
	case <-time.After(time.Second):
		t.Fatal("timed-out query goroutine did not finish")
	}
}
