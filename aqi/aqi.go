package aqi

import (
	"bytes"
	"encoding/json"
	"github.com/mobius-scheduler/apps/utils"
	"github.com/mobius-scheduler/mobius/app"
	"github.com/mobius-scheduler/mobius/common"
	log "github.com/sirupsen/logrus"
	"os"
	"os/exec"
)

type AppAQI struct {
	app          app.Application
	tasks        []common.TaskData
	interest_map common.InterestMap
	app_id       int
	ground_truth map[utils.MeasurementKey]float64
	measurements []AQIMeasurement
	locations    []common.Location
	cfg          AQIConfig
}

type AQIConfig struct {
	TaskPath            string  `json:"task_path"`
	GroundTruthPath     string  `json:"ground_truth_path"`
	ConfidenceThreshold float64 `json:"confidence_threshold"`
}

const (
	AQI_GROUND_TRUTH         = "GroundTruth/aqi_database.json"
	AQI_CONFIDENCE_THRESHOLD = 1e-6
)

type AQIMeasurement struct {
	Location common.Location `json:"location"`
	AQI      float64         `json:"aqi"`
}

// scheme for input, output to Gaussian Process fit
type GPInput struct {
	Measurements []AQIMeasurement  `json:"measurements"`
	Queries      []common.Location `json:"queries"`
	Threshold    float64           `json:"threshold"`
	Time         int               `json:"time"`
}

// init app interest map
func (a *AppAQI) Init(cfg app.AppConfig) {
	a.app_id = cfg.AppID

	// parse config
	js, _ := json.Marshal(cfg.Config)
	json.Unmarshal(js, &a.cfg)

	a.interest_map = make(common.InterestMap)

	a.tasks = utils.LoadTasksFromFile(a.cfg.TaskPath)
	for t, _ := range a.tasks {
		var task = a.tasks[t]
		task.TaskTimeSeconds = 20
		a.tasks[t] = task
	}
	a.add_tasks()

	a.ground_truth = utils.LoadGroundTruth(a.cfg.GroundTruthPath)
}

func (a *AppAQI) add_tasks() {
	a.locations = make([]common.Location, len(a.tasks))
	i := 0
	for _, t := range a.tasks {
		task := common.Task{
			AppID:       t.AppID,
			Location:    t.Location,
			RequestTime: t.RequestTime,
		}
		a.interest_map[task] = t
		a.app_id = t.AppID
		a.locations[i] = t.Location
		i++
	}
}

// get app id
func (a *AppAQI) GetID() int {
	return a.app_id
}

// get current interest map
func (a *AppAQI) GetInterestMap() common.InterestMap {
	return a.interest_map
}

// run GP fit to extract cells with uncertainty above threshold
func (a *AppAQI) Update(completed_tasks []common.TaskData, time int) {
	// record measurements
	for _, t := range completed_tasks {
		m := utils.MeasurementKey{Location: t.Location}
		if _, ok := a.ground_truth[m]; !ok {
			log.Fatalf("[app] could not index aqi measurement at %v", m)
		}
		log.Debugf(
			"[app] measured aqi %v at %v, time %v",
			a.ground_truth[m],
			t.Location,
			t.FulfillTime,
		)
		a.add_measurement(t.Location, t.FulfillTime, a.ground_truth[m])
	}

	// run GP fit, update interestmap
	if len(a.measurements) > 0 {
		uncertain := a.run_gp_fit(time)
		log.Debugf("uncertainty: %+v", uncertain)
		a.interest_map = make(common.InterestMap)
		for _, t := range a.tasks {
			if _, exists := uncertain[t.Location]; exists {
				a.interest_map[t.GetTask()] = t
			}
		}
	}
}

// add measurement
func (a *AppAQI) add_measurement(location common.Location, time int, aqi float64) {
	a.measurements = append(
		a.measurements,
		AQIMeasurement{
			Location: location,
			AQI:      aqi,
		},
	)
}

// run GP fit
func (a *AppAQI) run_gp_fit(time int) map[common.Location]bool {
	inp := GPInput{
		Measurements: a.measurements,
		Queries:      a.locations,
		Threshold:    a.cfg.ConfidenceThreshold,
		Time:         time,
	}
	inpj := common.ToJSON(inp)

	// run GP fit
	cmd := exec.Command("python3", "gp.py")
	cmd.Dir = common.GetDir()
	var inpbuf, outbuf bytes.Buffer
	inpbuf.Write(inpj)
	cmd.Stdin = &inpbuf
	cmd.Stdout = &outbuf
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		log.Fatalf("[app] error running gp fit: %v", err)
	}

	var out []common.Location
	if err := json.Unmarshal(outbuf.Bytes(), &out); err != nil {
		log.Fatalf("[app] error unmarshaling json to GP output struct: %v", err)
	}

	uncertain := make(map[common.Location]bool)
	for _, loc := range out {
		uncertain[loc] = true
	}

	return uncertain
}
