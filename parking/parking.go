package parking

import (
	"github.com/mobius-scheduler/apps/utils"
	"github.com/mobius-scheduler/mobius/app"
	"github.com/mobius-scheduler/mobius/common"
)

type AppParking struct {
	app          app.Application
	interest_map common.InterestMap
	app_id       int
	tasks        []common.TaskData
	cfg          ParkingConfig
}

type ParkingConfig struct {
	TaskPath    string `json:"task_path"`
	IntervalSec int    `json:"interval_sec"`
}

// init app interest map
func (a *AppParking) Init(cfg app.AppConfig) {
	a.app_id = cfg.AppID

	utils.JSONToStruct(cfg.Config, &a.cfg)

	a.interest_map = make(common.InterestMap)

	a.tasks = utils.LoadTasksFromFile(a.cfg.TaskPath)
	a.populate_im()
}

// get app id
func (a *AppParking) GetID() int {
	return a.app_id
}

// get current interest map
func (a *AppParking) GetInterestMap() common.InterestMap {
	return a.interest_map
}

// remove completed tasks from interest map
func (a *AppParking) Update(completed_tasks []common.TaskData, time int) {
	for _, t := range completed_tasks {
		delete(a.interest_map, t.GetTask())
	}

	if time%a.cfg.IntervalSec == 0 {
		a.populate_im()
	}
}

func (a *AppParking) populate_im() {
	for _, t := range a.tasks {
		task := common.Task{
			AppID:    t.AppID,
			Location: t.Location,
		}
		a.interest_map[task] = t
		a.app_id = t.AppID
	}
}
