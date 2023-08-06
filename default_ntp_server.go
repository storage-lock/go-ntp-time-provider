package ntp_time_provider

// DefaultNtpServers 默认的NTP服务器，使用者可以在系统初始化的覆盖掉这个变量来设置默认的NTP服务器
var DefaultNtpServers = []string{
	"time.windows.com",
	"time.nist.gov",
	"ntp.ntsc.ac.cn",
	"ntp.aliyun.com",
	"time1.cloud.tencent.com",
	"time2.cloud.tencent.com",
	"time3.cloud.tencent.com",
	"time4.cloud.tencent.com",
	"time5.cloud.tencent.com",
}

// SetDefaultNtpServers 或者调用这个方法来设置默认的NTP服务器
func SetDefaultNtpServers(defaultNtpServers []string) {
	DefaultNtpServers = defaultNtpServers
}

func GetDefaultNtpServers() []string {
	return DefaultNtpServers
}
