package main

import "github.com/pellaeon/cachet-monitor/monitors"

func Check(Type string, Parameter interface{}, Expect interface{}) {
	var success bool
	var responseTime uint
	switch Type {
	case "http":
		success, responseTime = monitors.HTTPMonitor.Run()
	default:
		log.Printf("No checker for type %s found", Type)
	}
}
