package main

import (
	"github.com/pellaeon/cachet-monitor/cachet"
	"github.com/tideland/golib/logger"
	"os"
	"time"
)

func main() {
	config := cachet.Config
	logger.SetLogger(logger.NewTimeformatLogger(os.Stderr, "2006-01-02 15:04:05"))
	logger.SetLevel(logger.LevelDebug)

	logger.Infof("System: %s, API: %s", config.SystemName, config.APIUrl)
	logger.Infof("Starting %d monitors", len(config.MonitorConfigs))

	// initialize monitors
	var allMonitors []*Monitor
	for _, monconf := range config.MonitorConfigs {
		err, mon := NewMonitor(&monconf)
		if err == nil {
			allMonitors = append(allMonitors, mon)
		} else {
			logger.Errorf("Parsing monitor error, skipping: %v", err)
		}
	}

	ticker := time.NewTicker(time.Second * time.Duration(config.CheckInterval))
	for range ticker.C {
		for _, m := range allMonitors {
			go m.Check()
		}
	}
}
