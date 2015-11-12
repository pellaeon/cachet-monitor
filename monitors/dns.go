package monitors

import (
	"fmt"
	"github.com/miekg/dns"
	"math"
)

type DNSChecker struct {
	Parameters struct {
		Rr_type string
		Query   string
		Server  string
	}
	Expect struct {
		Contain string
	}
}

func (dc *DNSChecker) Check() (bool, int64, string) {
	m := new(dns.Msg)
	c := new(dns.Client)
	switch dc.Parameters.Rr_type {
	case "A":
		m.SetQuestion(dc.Parameters.Query, dns.TypeA)
	case "MX":
		m.SetQuestion(dc.Parameters.Query, dns.TypeMX)
	default:
		fmt.Println("DNSChecker: unknown query type")
		return false, -1, "DNSChecker: unknown query type"
	}
	res, rtt, err := c.Exchange(m, dc.Parameters.Server+":53")
	if err != nil {
		fmt.Printf("DNSChecker: %v\n", err)
		return false, -1, err.Error()
	}
	var isUp bool
	if res.Rcode == dns.RcodeSuccess {
		isUp = true
	} else {
		isUp = false
	}
	reason := dns.RcodeToString[res.Rcode]
	// rounding
	//fmt.Printf("%v -> %v\n", rtt.Nanoseconds(), math.Floor((float64(rtt.Nanoseconds())/1000000.0)+0.5))
	return isUp, int64(math.Floor((float64(rtt.Nanoseconds()) / 1000000.0) + 0.5)), reason
}
