package cachet

import (
	"encoding/json"
	"github.com/tideland/golib/logger"
	"strconv"
)

// SendMetric sends lag metric point
func SendMetric(metricID int, delay int64) {
	if metricID <= 0 {
		return
	}

	jsonBytes, _ := json.Marshal(&map[string]interface{}{
		"value": delay,
	})

	resp, _, err := MakeRequest("POST", "/metrics/"+strconv.Itoa(metricID)+"/points", jsonBytes)
	if err != nil || resp.StatusCode != 200 {
		logger.Errorf("Could not log data point: %v", err)
		return
	}
}
