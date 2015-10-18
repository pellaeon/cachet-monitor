package main

import (
	"encoding/json"
	"errors"
	"github.com/pellaeon/cachet-monitor/cachet"
	"github.com/pellaeon/cachet-monitor/monitors"
	"github.com/pellaeon/cachet-monitor/system"
	"strconv"
)

type Monitor struct {
	History        []bool
	LastFailReason string
	Incident       *cachet.Incident
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
			json.Unmarshal(m.Parameters, &checker.Parameters)
			json.Unmarshal(m.Expect, &checker.Expect)
			m.Checker = &checker
		default:
			return errors.New("Unknown type: " + string(m.Type)), nil
		}
	}

	return nil, &m
}

func (m *Monitor) Check() {
	var success bool
	var responseTime uint
	success, responseTime, m.LastFailReason = m.Checker.Check()
	_ = responseTime //TODO remove
	m.historyAppend(success)
	m.AnalyseData()
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

	if t > m.Threshold && m.Incident == nil {
		// is down, create an incident
		cachet.Logger.Println("Creating incident...")

		component_id := json.Number(strconv.Itoa(int(m.ComponentID)))
		m.Incident = &cachet.Incident{
			Name:        m.Name + " - " + system.GetHostname(), // XXX
			Message:     m.Name + " check failed",
			ComponentID: &component_id,
		}

		if m.LastFailReason != "" {
			m.Incident.Message += "\n\n - " + m.LastFailReason
		}

		// set investigating status
		m.Incident.SetInvestigating()

		// create/update incident
		m.Incident.Send()
		m.Incident.UpdateComponent()
	} else if t < m.Threshold && m.Incident != nil {
		// was down, created an incident, its now ok, make it resolved.
		cachet.Logger.Println("Updating incident to resolved...")

		component_id := json.Number(strconv.Itoa(int(m.ComponentID)))
		m.Incident = &cachet.Incident{
			Name:        m.Incident.Name,
			Message:     m.Name + " check succeeded",
			ComponentID: &component_id,
		}

		m.Incident.SetFixed()
		m.Incident.Send()
		m.Incident.UpdateComponent()

		m.Incident = nil
	}
}

func (m *Monitor) historyAppend(current bool) {
	if len(m.History) >= 10 { //TODO: make configurable
		m.History = m.History[len(m.History)-9:]
	}
	m.History = append(m.History, current)
}
