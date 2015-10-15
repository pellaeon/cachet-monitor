package monitors

import (
	"crypto/tls"
	"fmt"
	"net/http"
	"strconv"
	"time"
)

const timeout = time.Duration(time.Second)

// HTTPChecker data model
type HTTPChecker struct {
	URL                string `json:"url"`
	ExpectedStatusCode int    `json:"expected_status_code"`
	StrictTLS          bool   `json:"strict_tls"`
	Parameters         struct {
		URL        string
		Strict_tls bool
	}
	Expect struct {
		Status_code uint
	}
}

//func (HTTPChecker *HTTPChecker) New(parameter, expect interface{}) *Checker {

// Run loop
func (HTTPChecker *HTTPChecker) Check() (bool, uint, string) {
	fmt.Println("HTTP.Parameters.URL= " + HTTPChecker.Parameters.URL)
	reqStart := getMs()
	isUp, reason := HTTPChecker.doRequest()
	lag := getMs() - reqStart

	/* TODO
	if isUp == true && HTTPChecker.MetricID > 0 {
		SendMetric(HTTPChecker.MetricID, lag)
	}*/

	return isUp, uint(lag), reason
}
func (HTTPChecker *HTTPChecker) Test() string {
	return HTTPChecker.Parameters.URL
}

func (HTTPChecker *HTTPChecker) doRequest() (bool, string) {
	client := &http.Client{
		Timeout: timeout,
	}
	if HTTPChecker.StrictTLS == false {
		client.Transport = &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		}
	}

	resp, err := client.Get(HTTPChecker.URL)
	if err != nil {
		return false, err.Error()
	}

	defer resp.Body.Close()

	if resp.StatusCode != HTTPChecker.ExpectedStatusCode {
		reason := "Unexpected response code: " + strconv.Itoa(resp.StatusCode) + ". Expected " + strconv.Itoa(HTTPChecker.ExpectedStatusCode)
		return false, reason
	}

	return true, ""
}

func getMs() int64 {
	return time.Now().UnixNano() / int64(time.Millisecond)
}
