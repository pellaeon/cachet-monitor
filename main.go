package main

import (
	"fmt"
	"github.com/pellaeon/cachet-monitor/cachet"
	"time"
)

func main() {
	config := cachet.Config
	log := cachet.Logger

	log.Printf("System: %s, API: %s\n", config.SystemName, config.APIUrl)
	log.Printf("Starting %d monitors:\n", len(config.MonitorConfigs))

	// initialize monitors
	var allMonitors []*Monitor
	for _, monconf := range config.MonitorConfigs {
		err, mon := NewMonitor(&monconf)
		fmt.Println(mon.Checker.Test())
		if err == nil {
			allMonitors = append(allMonitors, mon)
		} else {
			log.Printf("Parsing monitor error, skipping: %v", err)
		}
	}

	ticker := time.NewTicker(time.Second * time.Duration(config.CheckInterval))
	for range ticker.C {
		for _, m := range allMonitors {
			go m.Check()
		}
	}
}
