package storage

import (
	"context"
	"fmt"
	"github.com/storage-lock/go-events"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestNTPTimeProvider_GetTime(t *testing.T) {

	e := events.NewEvent("xxx").AddListeners(events.NewListenerWrapper("test-listener", func(ctx context.Context, e *events.Event) {
		fmt.Println(e.ToJsonString())
	}))

	for _, server := range DefaultNtpServers {
		provider := NewNTPTimeProvider(e.Fork(), server)
		time, err := provider.GetTime(context.Background())
		assert.Nil(t, err)
		assert.False(t, time.IsZero())
		if err != nil {
			fmt.Println(server)
		}
	}
}
