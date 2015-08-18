package monitors

import (
	"crypto/tls"
	"encoding/json"
	"github.com/pellaeon/cachet-monitor/cachet"
	"github.com/pellaeon/cachet-monitor/system"
	"net/http"
	"strconv"
	"time"
)

const timeout = time.Duration(time.Second)

// HTTPMonitor data model
type HTTPMonitor struct {
	Name               string  `json:"name"`
	URL                string  `json:"url"`
	MetricID           int     `json:"metric_id"`
	Threshold          float32 `json:"threshold"`
	ComponentID        *int    `json:"component_id"`
	ExpectedStatusCode int     `json:"expected_status_code"`
	StrictTLS          *bool   `json:"strict_tls"`

	History        []bool           `json:"-"`
	LastFailReason *string          `json:"-"`
	Incident       *cachet.Incident `json:"-"`
}

// Run loop
func (HTTPMonitor *HTTPMonitor) Run() (bool, uint) {
	reqStart := getMs()
	isUp := HTTPMonitor.doRequest()
	lag := getMs() - reqStart

	if len(HTTPMonitor.History) >= 10 {
		HTTPMonitor.History = HTTPMonitor.History[len(HTTPMonitor.History)-9:]
	}
	HTTPMonitor.History = append(HTTPMonitor.History, isUp)
	HTTPMonitor.AnalyseData()

	/* TODO
	if isUp == true && HTTPMonitor.MetricID > 0 {
		SendMetric(HTTPMonitor.MetricID, lag)
	}*/

	return isUp, uint(lag)
}

func (HTTPMonitor *HTTPMonitor) doRequest() bool {
	client := &http.Client{
		Timeout: timeout,
	}
	if HTTPMonitor.StrictTLS != nil && *HTTPMonitor.StrictTLS == false {
		client.Transport = &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		}
	}

	resp, err := client.Get(HTTPMonitor.URL)
	if err != nil {
		errString := err.Error()
		HTTPMonitor.LastFailReason = &errString
		return false
	}

	defer resp.Body.Close()

	if resp.StatusCode != HTTPMonitor.ExpectedStatusCode {
		failReason := "Unexpected response code: " + strconv.Itoa(resp.StatusCode) + ". Expected " + strconv.Itoa(HTTPMonitor.ExpectedStatusCode)
		HTTPMonitor.LastFailReason = &failReason
		return false
	}

	return true
}

// AnalyseData decides if the HTTPMonitor is statistically up or down and creates / resolves an incident
func (HTTPMonitor *HTTPMonitor) AnalyseData() {
	// look at the past few incidents
	numDown := 0
	for _, wasUp := range HTTPMonitor.History {
		if wasUp == false {
			numDown++
		}
	}

	t := (float32(numDown) / float32(len(HTTPMonitor.History))) * 100
	cachet.Logger.Printf("%s %.2f%% Down at %v. Threshold: %.2f%%\n", HTTPMonitor.URL, t, time.Now().UnixNano()/int64(time.Second), HTTPMonitor.Threshold)

	if len(HTTPMonitor.History) != 10 {
		// not enough data
		return
	}

	if t > HTTPMonitor.Threshold && HTTPMonitor.Incident == nil {
		// is down, create an incident
		cachet.Logger.Println("Creating incident...")

		component_id := json.Number(strconv.Itoa(*HTTPMonitor.ComponentID))
		HTTPMonitor.Incident = &cachet.Incident{
			Name:        HTTPMonitor.Name + " - " + system.GetHostname(), // XXX
			Message:     HTTPMonitor.Name + " check failed",
			ComponentID: &component_id,
		}

		if HTTPMonitor.LastFailReason != nil {
			HTTPMonitor.Incident.Message += "\n\n - " + *HTTPMonitor.LastFailReason
		}

		// set investigating status
		HTTPMonitor.Incident.SetInvestigating()

		// create/update incident
		HTTPMonitor.Incident.Send()
		HTTPMonitor.Incident.UpdateComponent()
	} else if t < HTTPMonitor.Threshold && HTTPMonitor.Incident != nil {
		// was down, created an incident, its now ok, make it resolved.
		cachet.Logger.Println("Updating incident to resolved...")

		component_id := json.Number(strconv.Itoa(*HTTPMonitor.ComponentID))
		HTTPMonitor.Incident = &cachet.Incident{
			Name:        HTTPMonitor.Incident.Name,
			Message:     HTTPMonitor.Name + " check succeeded",
			ComponentID: &component_id,
		}

		HTTPMonitor.Incident.SetFixed()
		HTTPMonitor.Incident.Send()
		HTTPMonitor.Incident.UpdateComponent()

		HTTPMonitor.Incident = nil
	}
}

func getMs() int64 {
	return time.Now().UnixNano() / int64(time.Millisecond)
}
