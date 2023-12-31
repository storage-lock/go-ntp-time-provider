# NTP Time Provider

# 一、这是什么

基于NTP的TimeProvider实现，Storage的具体实现可以引入这个库来实现GetTime方法，让分布式系统中的各个角色使用统一的NTP时间源。

# 二、安装依赖

```bash
go get -u github.com/storage-lock/go-ntp-time-provider
```

# 三、API示例

```go
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
```

