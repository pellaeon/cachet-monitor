package monitors

import (
	"github.com/beevik/ntp"
	"github.com/pellaeon/cachet-monitor/cachet"
)

type NTPChecker struct {
	Parameters struct {
		Server string
	}
	Expect struct {
	}
}

func (nc *NTPChecker) Check() (bool, int64, string) {
	reqStart := getMs()
	_, err := ntp.Time(nc.Parameters.Server)
	lag := getMs() - reqStart
	if err != nil {
		cachet.Logger.Println(err.Error())
		return false, lag, err.Error()
	}
	return true, lag, ""
}
