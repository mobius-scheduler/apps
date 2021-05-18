package roof

import (
	"github.com/mobius-scheduler/apps/utils"
	"github.com/mobius-scheduler/mobius/app"
	"github.com/mobius-scheduler/mobius/common"
)

type AppRoof struct {
	app          app.Application
	interest_map common.InterestMap
	app_id       int
	tasks        []common.TaskData
	cfg          RoofConfig
}

type RoofConfig struct {
	TaskPath string `json:"task_path"`
	StartSec int    `json:"start_sec"`
}

// init app interest map
func (a *AppRoof) Init(cfg app.AppConfig) {
	a.app_id = cfg.AppID
	a.interest_map = make(common.InterestMap)

	utils.JSONToStruct(cfg.Config, &a.cfg)
	a.tasks = utils.LoadTasksFromFile(a.cfg.TaskPath)
}

// get app id
func (a *AppRoof) GetID() int {
	return a.app_id
}

// get current interest map
func (a *AppRoof) GetInterestMap() common.InterestMap {
	return a.interest_map
}

// remove completed tasks from interest map
func (a *AppRoof) Update(completed_tasks []common.TaskData, time int) {
	if time == a.cfg.StartSec {
		a.populate_im()
	}
	for _, t := range completed_tasks {
		delete(a.interest_map, t.GetTask())
	}
}

func (a *AppRoof) populate_im() {
	for _, t := range a.tasks {
		task := common.Task{
			AppID:       t.AppID,
			Location:    t.Location,
			RequestTime: t.RequestTime,
		}
		a.interest_map[task] = t
	}
}
