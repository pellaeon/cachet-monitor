package main

import (
	"github.com/pellaeon/cachet-monitor/cachet"
	"time"
)

func main() {
	config := cachet.Config
	log := cachet.Logger

	log.Printf("System: %s, API: %s\n", config.SystemName, config.APIUrl)
	log.Printf("Starting %d monitors:\n", len(config.Monitors))
	for _, mon := range config.Monitors {
		log.Printf(" %s: GET %s & Expect HTTP %d\n", mon.Name, mon.URL, mon.ExpectedStatusCode)
		if mon.MetricID > 0 {
			log.Printf(" - Logs lag to metric id: %d\n", mon.MetricID)
		}
	}

	log.Println()

	ticker := time.NewTicker(time.Second * time.Duration(config.CheckInterval))
	for range ticker.C {
		for _, mon := range config.Monitors {
			go mon.Run()
		}
	}
}
