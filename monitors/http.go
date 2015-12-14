package monitors

import (
	"bytes"
	"crypto/tls"
	"errors"
	"fmt"
	"github.com/tideland/golib/logger"
	"net"
	"net/http"
	"strconv"
	"strings"
	"time"
)

const timeout = time.Duration(time.Second)

// HTTPChecker data model
type HTTPChecker struct {
	Parameters struct {
		URL        string // will override IP, ServerName, Scheme, Port, Path if non empty
		Strict_tls bool
		IP         string // IP to connect to
		ServerName string // used in HTTP Host header
		Scheme     string // http or https
		Port       uint
		Path       string // starts with slash
	}
	Expect struct {
		Status_code     int
		Contain_keyword string
	}
}

// Run loop
func (HTTPChecker *HTTPChecker) Check() (bool, int64, string) {
	reqStart := getMs()
	checkerr := HTTPChecker.doRequest()
	lag := getMs() - reqStart
	if checkerr == nil {
		return true, lag, ""
	} else {
		return false, lag, checkerr.Error()
	}

}
func (HTTPChecker *HTTPChecker) Test() string {
	return HTTPChecker.Parameters.URL
}

func (HTTPChecker *HTTPChecker) doRequest() error {
	client := &http.Client{
		Timeout: timeout,
	}
	if HTTPChecker.Parameters.Strict_tls == false {
		client.Transport = &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		}
	}

	var err error
	var resp *http.Response
	if HTTPChecker.Parameters.URL != "" {
		resp, err = client.Get(HTTPChecker.Parameters.URL)
	} else {
		ip := net.ParseIP(HTTPChecker.Parameters.IP)
		if ip == nil {
			return fmt.Errorf("Cannot parse IP: %s", HTTPChecker.Parameters.IP)
		}
		if !(HTTPChecker.Parameters.Scheme == "http" || HTTPChecker.Parameters.Scheme == "https") {
			return fmt.Errorf("Scheme must be http or https")
		}
		if HTTPChecker.Parameters.Port > 65535 || HTTPChecker.Parameters.Port < 1 {
			return fmt.Errorf("Port must be 1-65535")
		}
		req, err := http.NewRequest("GET", HTTPChecker.Parameters.Scheme+"://"+HTTPChecker.Parameters.IP+":"+
			strconv.Itoa(int(HTTPChecker.Parameters.Port))+HTTPChecker.Parameters.Path, nil)
		if err != nil {
			return err
		}
		req.Host = HTTPChecker.Parameters.ServerName
		resp, err = client.Do(req)
	}

	if err != nil {
		return err
	}

	defer resp.Body.Close()

	if resp.StatusCode != HTTPChecker.Expect.Status_code {
		reason := "Unexpected response code: " + strconv.Itoa(resp.StatusCode) + ". Expected " + strconv.Itoa(HTTPChecker.Expect.Status_code)
		return errors.New(reason)
	}

	if HTTPChecker.Expect.Contain_keyword != "" {
		bodybuf := new(bytes.Buffer)
		_, err := bodybuf.ReadFrom(resp.Body)
		body_s := bodybuf.String()
		if err != nil {
			logger.Warningf("HTTPChecker: " + err.Error())
			return errors.New("HTTPChecker: " + err.Error())
		}
		if !strings.Contains(body_s, HTTPChecker.Expect.Contain_keyword) {
			logger.Infof("Response does not contain keyword: " + HTTPChecker.Expect.Contain_keyword)
			return errors.New("Response does not contain keyword: " + HTTPChecker.Expect.Contain_keyword)
		}
	}

	return nil
}

func getMs() int64 {
	return time.Now().UnixNano() / int64(time.Millisecond)
}
