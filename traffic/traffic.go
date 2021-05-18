package traffic

import (
	"github.com/gonum/stat"
	"github.com/mobius-scheduler/apps/utils"
	"github.com/mobius-scheduler/mobius/app"
	"github.com/mobius-scheduler/mobius/common"
	log "github.com/sirupsen/logrus"
	"sort"
)

type AppTraffic struct {
	app          app.Application
	tasks        []common.TaskData
	interest_map common.InterestMap
	app_id       int
	ground_truth map[utils.MeasurementKey]float64
	measurements map[common.Location][]SpeedMeasurement
	init_done    []bool
	cfg          TrafficConfig
}

type TrafficConfig struct {
	TaskPath            string `json:"task_path"`
	GroundTruthPath     string `json:"ground_truth_path"`
	NumSamplesInit      int    `json:"num_samples_init"`
	SampleValidSec      int    `json:"sample_valid_sec"`
	NumSamplesUncertain int    `json:"num_samples_uncertain"`
}

type SpeedMeasurement struct {
	Location common.Location
	Time     int
	Speed    float64
}

// init app interest map
func (a *AppTraffic) Init(cfg app.AppConfig) {
	a.app_id = cfg.AppID
	a.interest_map = make(common.InterestMap)
	a.measurements = make(map[common.Location][]SpeedMeasurement)

	utils.JSONToStruct(cfg.Config, &a.cfg)

	// load tasks
	a.tasks = utils.LoadTasksFromFile(a.cfg.TaskPath)
	for i, t := range a.tasks {
		task := common.Task{
			AppID:       t.AppID,
			Location:    t.Location,
			RequestTime: t.RequestTime,
		}
		t.TaskTimeSeconds = 25
		a.interest_map[task] = t
		a.tasks[i] = t
		a.measurements[t.Location] = []SpeedMeasurement{}
	}

	// set path to ground truth
	a.ground_truth = utils.LoadGroundTruth(a.cfg.GroundTruthPath)
	a.init_done = make([]bool, a.cfg.NumSamplesInit)
}

// get app id
func (a *AppTraffic) GetID() int {
	return a.app_id
}

// get current interest map
func (a *AppTraffic) GetInterestMap() common.InterestMap {
	return a.interest_map
}

// remove completed tasks from interest map
func (a *AppTraffic) Update(completed_tasks []common.TaskData, time int) {
	// record measurements
	for _, t := range completed_tasks {
		m := utils.MeasurementKey{Location: t.Location, Time: t.FulfillTime}
		if _, ok := a.ground_truth[m]; !ok {
			log.Debugf("[app] could not index speed measurement at %v", m)
		}
		log.Debugf(
			"[app] measured speed %v at %v, time %v",
			a.ground_truth[m],
			t.Location,
			t.FulfillTime,
		)
		a.add_measurement(t.Location, t.FulfillTime, a.ground_truth[m])
	}

	// update interestmap
	// priority: (1) init, (2) expired, (3) revisit
	a.interest_map = make(common.InterestMap)
	if stage := a.get_init_stage(); stage != -1 {
		log.Debugf("[app] traffic app, init phase %v", stage)
		for _, t := range a.tasks {
			if len(a.measurements[t.Location]) < stage+1 {
				a.interest_map[t.GetTask()] = t
			}
		}
	} else if expired := a.get_expired_locations(time); len(expired) > 0 {
		log.Debugf("[app] renewing %v expired locations", len(expired))
		for _, t := range a.tasks {
			if _, exists := expired[t.Location]; exists {
				a.interest_map[t.GetTask()] = t
			}
		}
	} else {
		uncertain := a.get_uncertain_locations()
		log.Debugf("[app] submitting %v uncertain locations", len(uncertain))
		for _, t := range a.tasks {
			if _, exists := uncertain[t.Location]; exists {
				a.interest_map[t.GetTask()] = t
			}
		}
	}
	if len(a.interest_map) == 0 {
		log.Fatalf("[app] interestmap length should be > 0")
	}
}

func (a *AppTraffic) add_measurement(location common.Location, time int, speed float64) {
	// add measurement
	a.measurements[location] = append(
		a.measurements[location],
		SpeedMeasurement{
			Location: location,
			Time:     time,
			Speed:    speed,
		},
	)

	// update initialization progress
out:
	for stage := 0; stage < a.cfg.NumSamplesInit; stage++ {
		for _, m := range a.measurements {
			if len(m) < stage+1 {
				break out
			}
		}
		a.init_done[stage] = true
	}
}

// check if done collecting initial measurements
// return stage to focus on; if done with init, return -1
func (a *AppTraffic) get_init_stage() int {
	for stage, d := range a.init_done {
		if !d {
			return stage
		}
	}
	return -1
}

// revisit "expired" locations
// measurements are valid for SAMPLE_VALID_SEC
func (a *AppTraffic) get_expired_locations(time int) map[common.Location]bool {
	expired := make(map[common.Location]bool)
	for _, m := range a.measurements {
		end := m[len(m)-1]
		if time-end.Time >= a.cfg.SampleValidSec {
			expired[end.Location] = true
		}
	}
	return expired
}

// get uncertain locations (based on stdev)
func (a *AppTraffic) get_uncertain_locations() map[common.Location]bool {
	// compute summary of measurements (grouped by location)
	stdev := make([]SpeedMeasurement, len(a.measurements))
	var idx int
	for loc, m := range a.measurements {
		speeds := make([]float64, len(m))
		for _, x := range m {
			speeds = append(speeds, x.Speed)
		}
		stdev[idx] = SpeedMeasurement{
			Location: loc,
			Speed:    stat.StdDev(speeds, nil),
		}
		idx++
	}

	sort.Slice(
		stdev,
		func(i, j int) bool {
			return stdev[i].Speed > stdev[j].Speed
		},
	)

	uncertain := make(map[common.Location]bool)
	for i := 0; i < a.cfg.NumSamplesUncertain; i++ {
		loc := stdev[i].Location
		uncertain[loc] = true
	}
	return uncertain
}
