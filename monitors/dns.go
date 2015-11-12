package monitors

import (
	"github.com/miekg/dns"
	"github.com/tideland/golib/logger"
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
		logger.Warningf("DNSChecker: unsupported query type")
		return false, -1, "DNSChecker: unsupported query type"
	}
	res, rtt, err := c.Exchange(m, dc.Parameters.Server+":53")
	if err != nil {
		logger.Warningf("DNSChecker: %v", err)
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
