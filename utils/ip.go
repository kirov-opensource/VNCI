package utils

import (
	"net"

	"github.com/google/logger"
	"github.com/miekg/dns"
)

// 获取实际IP
func GetActualIP(address string) string {
	config, _ := dns.ClientConfigFromFile("/etc/resolv.conf")
	c := new(dns.Client)
	m := new(dns.Msg)
	m.SetQuestion(dns.Fqdn(address), dns.TypeA)
	m.RecursionDesired = true
	// client 发起 DNS 请求，其中 c 为上文创建的 client，m 为构造的 DNS 报文
	// config 为从 /etc/resolv.conf 构造出来的配置
	r, _, err := c.Exchange(m, net.JoinHostPort(config.Servers[0], config.Port))
	if r == nil {
		logger.Errorln("*** dns解析失败: %s\n", err.Error())
		return "1.1.1.1"
	}

	if r.Rcode != dns.RcodeSuccess {
		return "1.1.1.1"
	}

	// 如果 DNS 查询成功
	for _, a := range r.Answer {
		if dnsA, ok := a.(*dns.A); ok {
			return dnsA.A.String()
		}
		// fmt.Printf("%v\n", a)
	}
	return "1.1.1.1"
}

// 检测是否是IP
func CheckIPAddress(ip string) bool {
	if net.ParseIP(ip) == nil {
		return false
	} else {
		return true
	}
}
