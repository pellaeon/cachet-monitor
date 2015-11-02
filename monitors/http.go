package monitors

import (
	"crypto/tls"
	"github.com/pellaeon/cachet-monitor/cachet"
	"net/http"
	"strconv"
	"time"
)

const timeout = time.Duration(time.Second)

// HTTPChecker data model
type HTTPChecker struct {
	Parameters struct {
		URL        string
		Strict_tls bool
	}
	Expect struct {
		Status_code int
	}
}

// Run loop
func (HTTPChecker *HTTPChecker) Check() (bool, uint, string) {
	reqStart := getMs()
	isUp, reason := HTTPChecker.doRequest()
	lag := getMs() - reqStart

	return isUp, uint(lag), reason
}
func (HTTPChecker *HTTPChecker) Test() string {
	return HTTPChecker.Parameters.URL
}

func (HTTPChecker *HTTPChecker) doRequest() (bool, string) {
	client := &http.Client{
		Timeout: timeout,
	}
	if HTTPChecker.Parameters.Strict_tls == false {
		client.Transport = &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		}
	}

	resp, err := client.Get(HTTPChecker.Parameters.URL)
	if err != nil {
		return false, err.Error()
	}

	defer resp.Body.Close()

	if resp.StatusCode != HTTPChecker.Expect.Status_code {
		reason := "Unexpected response code: " + strconv.Itoa(resp.StatusCode) + ". Expected " + strconv.Itoa(HTTPChecker.Expect.Status_code)
		return false, reason
	}
	cachet.Logger.Printf("%d", resp.StatusCode)

	return true, ""
}

func getMs() int64 {
	return time.Now().UnixNano() / int64(time.Millisecond)
}
