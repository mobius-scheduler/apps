package utils

import (
	"encoding/json"
	"github.com/mobius-scheduler/mobius/common"
	log "github.com/sirupsen/logrus"
	"math"
	"strconv"
	"strings"
)

type MeasurementKey struct {
	Location common.Location
	Time     int
}

func JSONToStruct(j interface{}, s interface{}) {
	js, _ := json.Marshal(j)
	json.Unmarshal(js, s)
}

func LoadTasksFromFile(task_file string) []common.TaskData {
	var tasks []common.TaskData
	common.FromFile(task_file, &tasks)

	for i, _ := range tasks {
		var t = tasks[i]
		t.Location.Latitude = math.Round(t.Location.Latitude*1e6) / 1e6
		t.Location.Longitude = math.Round(t.Location.Longitude*1e6) / 1e6
		t.Destination.Latitude = math.Round(t.Destination.Latitude*1e6) / 1e6
		t.Destination.Longitude = math.Round(t.Destination.Longitude*1e6) / 1e6
		tasks[i] = t
	}

	return tasks
}

// load ground truth csv
func LoadGroundTruth(gt_file string) map[MeasurementKey]float64 {
	gtraw := make(map[string]float64)
	common.FromFile(gt_file, &gtraw)

	// create groundtruth database
	gt := make(map[MeasurementKey]float64)
	for key, val := range gtraw {
		m := strings.Split(key, " ")
		lat, err_lat := strconv.ParseFloat(m[1], 64)
		lon, err_lon := strconv.ParseFloat(m[2], 64)
		time, err_time := strconv.Atoi(m[0])
		if err_lat != nil || err_lon != nil || err_time != nil {
			log.Fatalf("[app] unable to parse traffic ground truth: %v", key)
		}
		sm := MeasurementKey{
			Location: common.Location{
				Latitude:  lat,
				Longitude: lon,
			},
			Time: time,
		}
		gt[sm] = val
	}
	return gt
}
