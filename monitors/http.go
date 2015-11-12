package monitors

import (
	"bytes"
	"crypto/tls"
	"github.com/tideland/golib/logger"
	"net/http"
	"strconv"
	"strings"
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
		Status_code     int
		Contain_keyword string
	}
}

// Run loop
func (HTTPChecker *HTTPChecker) Check() (bool, int64, string) {
	reqStart := getMs()
	isUp, reason := HTTPChecker.doRequest()
	lag := getMs() - reqStart

	return isUp, lag, reason
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
	logger.Debugf("%d", resp.StatusCode)

	if HTTPChecker.Expect.Contain_keyword != "" {
		bodybuf := new(bytes.Buffer)
		_, err := bodybuf.ReadFrom(resp.Body)
		body_s := bodybuf.String()
		if err != nil {
			logger.Warningf("HTTPChecker: " + err.Error())
			return false, "HTTPChecker: " + err.Error()
		}
		if !strings.Contains(body_s, HTTPChecker.Expect.Contain_keyword) {
			logger.Infof("Response does not contain keyword: " + HTTPChecker.Expect.Contain_keyword)
			return false, "Response does not contain keyword: " + HTTPChecker.Expect.Contain_keyword
		}
	}

	return true, ""
}

func getMs() int64 {
	return time.Now().UnixNano() / int64(time.Millisecond)
}
