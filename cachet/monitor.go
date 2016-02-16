package cachet

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/pellaeon/cachet-monitor/monitors"
	"github.com/tideland/golib/logger"
	"net/smtp"
	"strconv"
	"strings"
	"time"
)

type Monitor struct {
	Status         string
	History        []bool
	LastFailReason string
	Incident       *Incident
	MetricID       int `json:"metric_id"`
	Checker        Checker
	Threshold      float32 `json:"threshold"`
	Name           string  `json:"name"`
	ComponentID    uint    `json:"component_id"`
	Type           string
	Parameters     json.RawMessage
	Expect         json.RawMessage
}

func NewMonitor(monconfp *json.RawMessage) (error, *Monitor) {
	var m Monitor
	json.Unmarshal(*monconfp, &m)

	if m.Name == "" {
		return errors.New("Monitor \"name\" no set"), nil
	}
	if m.Parameters == nil {
		return errors.New("Monitor \"parameters\" no set"), nil
	}

	if m.Type == "" {
		return errors.New("Monitor \"type\" no set"), nil
	} else {
		switch m.Type {
		case "http":
			var checker monitors.HTTPChecker
			err := json.Unmarshal(m.Parameters, &checker.Parameters)
			if err != nil {
				logger.Errorf("Unmarshal: %v", err)
			}
			err = json.Unmarshal(m.Expect, &checker.Expect)
			if err != nil {
				logger.Errorf("Unmarshal: %v", err)
			}
			m.Checker = &checker
		case "dns":
			var checker monitors.DNSChecker
			err := json.Unmarshal(m.Parameters, &checker.Parameters)
			if err != nil {
				logger.Errorf("Unmarshal: %v", err)
			}
			err = json.Unmarshal(m.Expect, &checker.Expect)
			if err != nil {
				logger.Errorf("Unmarshal: %v", err)
			}
			m.Checker = &checker
		case "ntp":
			var checker monitors.NTPChecker
			err := json.Unmarshal(m.Parameters, &checker.Parameters)
			if err != nil {
				logger.Errorf("Unmarshal: %v", err)
			}
			err = json.Unmarshal(m.Expect, &checker.Expect)
			if err != nil {
				logger.Errorf("Unmarshal: %v", err)
			}
			m.Checker = &checker
		default:
			return errors.New("Unknown type: " + string(m.Type)), nil
		}
	}

	return nil, &m
}

func (m *Monitor) Check() {
	var success bool
	success, _, m.LastFailReason = m.Checker.Check()

	m.historyAppend(success)
	m.AnalyseData()
	// TODO metric
}

// AnalyseData decides if the Monitor is statistically up or down and creates / resolves an incident
func (m *Monitor) AnalyseData() {
	// look at the past few incidents
	numDown := 0
	for _, wasUp := range m.History {
		if wasUp == false {
			numDown++
		}
	}

	t := (float32(numDown) / float32(len(m.History))) * 100
	// TODO cachet.Logger.Printf("%s %.2f%% Down at %v. Threshold: %.2f%%\n", m.URL, t, time.Now().UnixNano()/int64(time.Second), m.Threshold)

	if len(m.History) != 10 {
		// not enough data
		return
	}

	if t > m.Threshold && m.Status != "D" {
		// is down, create an incident
		logger.Infof("Monitor %s is down...", m.Name)
		resp, _, err := MakeRequest("PATCH", "/api/monitors/"+strconv.Itoa(int(m.ComponentID))+"/", []byte("{\"status\":\"D\"}"))
		if err != nil || resp.StatusCode != 200 {
			logger.Errorf("Could not set monitor down: (resp code %d) %v", resp.StatusCode, err)
		}

		header := make(map[string]string)
		header["From"] = "notification@" + Config.SystemName
		header["To"] = Config.NotifEmail
		header["Subject"] = "Monitor " + m.Name + " is DOWN"
		header["MIME-Version"] = "1.0"
		header["Content-Type"] = "text/plain; charset=\"utf-8\""
		header["Content-Transfer-Encoding"] = "base64"

		body := time.Now().Format("Mon Jan 2 15:04:05 -0700 MST 2006") + "\r\n" + m.LastFailReason

		message := ""
		for k, v := range header {
			message += fmt.Sprintf("%s: %s\r\n", k, v)
		}
		message += "\r\n" + base64.StdEncoding.EncodeToString([]byte(body))
		auth := smtp.PlainAuth("", "", "", "localhost")
		err = smtp.SendMail(
			"localhost:smtp",
			auth,
			"notification@"+Config.SystemName,
			strings.Split(Config.NotifEmail, ","),
			[]byte(message),
		)
		if err != nil {
			logger.Errorf("Could not send email notification: %v", err)
		}

		m.Status = "D"

		// TODO create Incident
	} else if t < m.Threshold && m.Status != "U" {
		// was down, created an incident, its now ok, make it resolved.
		logger.Infof("Monitor %s is up...", m.Name)
		resp, _, err := MakeRequest("PATCH", "/api/monitors/"+strconv.Itoa(int(m.ComponentID))+"/", []byte("{\"status\":\"U\"}"))
		if err != nil || resp.StatusCode != 200 {
			logger.Errorf("Could not set monitor up: (resp code %d) %v", resp.StatusCode, err)
		}

		m.Status = "U"
		// TODO create Incident
	}
}

func (m *Monitor) historyAppend(current bool) {
	if len(m.History) >= 10 { //TODO: make configurable
		m.History = m.History[len(m.History)-9:]
	}
	m.History = append(m.History, current)
}
