package main

import (
	"context"
	"fmt"
	"github.com/storage-lock/go-events"
	ntp_time_provider "github.com/storage-lock/go-ntp-time-provider"
)

func main() {

	event := events.NewEvent("test")
	timeProvider := ntp_time_provider.NewNTPTimeProvider(event)
	time, err := timeProvider.GetTime(context.Background())
	if err != nil {
		panic(err)
	}
	fmt.Println(time)
	// Output:
	// 2023-08-07 01:36:18.29416265 +0800 CST m=-1.335806449

}
