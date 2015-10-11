package main

import (
	"github.com/pellaeon/cachet-monitor/cachet"
	_ "time"
)

func main() {
	config := cachet.Config
	log := cachet.Logger

	log.Printf("System: %s, API: %s\n", config.SystemName, config.APIUrl)
	log.Printf("Starting %d monitors:\n", len(config.MonitorConfigs))

	// initialize monitors
	var allMonitors []*Monitor
	for _, monconf := range config.MonitorConfigs {
		log.Println(monconf["type"]) //debug
		allMonitors = append(allMonitors, NewMonitor(&monconf))
	}

	/*
		ticker := time.NewTicker(time.Second * time.Duration(config.CheckInterval))
		for range ticker.C {
			for _, m := range allMonitors {
				go m.Check()
			}
		}*/
}
