package main

import (
	"github.com/pellaeon/cachet-monitor/cachet"
	"github.com/tideland/golib/logger"
	"os"
	"time"
)

func main() {
	config := cachet.Config
	// TODO support log path
	logger.SetLogger(logger.NewTimeformatLogger(os.Stderr, "2006-01-02 15:04:05"))
	logger.SetLevel(logger.LevelDebug)

	logger.Infof("System: %s, API: %s", config.SystemName, config.APIUrl)
	logger.Infof("Starting %d monitors", len(config.MonitorConfigs))

	// initialize monitors
	var allMonitors []*cachet.Monitor
	for _, monconf := range config.MonitorConfigs {
		err, mon := cachet.NewMonitor(&monconf)
		if err == nil {
			err = cachet.SyncMonitor(mon)
			if err != nil {
				logger.Errorf("%v", err)
			}
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
