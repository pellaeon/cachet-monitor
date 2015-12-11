package cachet

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"github.com/tideland/golib/logger"
	"io/ioutil"
	"net/http"
	"strconv"
)

func MakeRequest(requestType string, url string, reqBody []byte) (*http.Response, []byte, error) {
	req, err := http.NewRequest(requestType, Config.APIUrl+url, bytes.NewBuffer(reqBody))

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Token "+Config.APIToken)

	client := &http.Client{}
	if Config.InsecureAPI == true {
		client.Transport = &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		}
	}

	res, err := client.Do(req)
	if err != nil {
		return nil, []byte{}, err
	}

	defer res.Body.Close()
	body, _ := ioutil.ReadAll(res.Body)

	return res, body, nil
}

// Updates names of monitor on server, create monitor if not exist
func SyncMonitor(m *Monitor) error {
	monitorData := map[string]string{
		"name": m.Name,
	}
	monitorDataJson, _ := json.Marshal(monitorData)
	logger.Debugf("%s", string(monitorDataJson))
	req, err := http.NewRequest("PATCH", Config.APIUrl+"/api/monitors/"+strconv.Itoa(int(m.ComponentID))+"/", bytes.NewBuffer(monitorDataJson))

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Token "+Config.APIToken)

	client := &http.Client{}
	if Config.InsecureAPI == true {
		client.Transport = &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		}
	}

	res, err := client.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	/*
		if res.StatusCode == 404 {
			// no such monitor, create a new one
			monitorData["status"] = "U"
			// TODO need to write obtained monitor pk back to config
	*/
	if res.StatusCode != 200 {
		body, _ := ioutil.ReadAll(res.Body)
		return fmt.Errorf("Cannot sync monitor %d (status code %d): %s", m.ComponentID, res.StatusCode, string(body))
	}

	return nil
}
