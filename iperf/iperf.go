package iperf

import (
	"github.com/mobius-scheduler/apps/utils"
	"github.com/mobius-scheduler/mobius/app"
	"github.com/mobius-scheduler/mobius/common"
)

type AppIperf struct {
	app          app.Application
	interest_map common.InterestMap
	app_id       int
	tasks        []common.TaskData
	cfg          IperfConfig
}

type IperfConfig struct {
	TaskPath string `json:"task_path"`
}

// init app interest map
func (a *AppIperf) Init(cfg app.AppConfig) {
	a.app_id = cfg.AppID
	a.interest_map = make(common.InterestMap)

	utils.JSONToStruct(cfg.Config, &a.cfg)

	a.tasks = utils.LoadTasksFromFile(a.cfg.TaskPath)
	a.populate_im()
}

// get app id
func (a *AppIperf) GetID() int {
	return a.app_id
}

// get current interest map
func (a *AppIperf) GetInterestMap() common.InterestMap {
	return a.interest_map
}

// keep completed tasks in interest map
func (a *AppIperf) Update(completed_tasks []common.TaskData, time int) {
	for _, t := range completed_tasks {
		delete(a.interest_map, t.GetTask())
	}

	if len(a.interest_map) == 0 {
		a.populate_im()
	}
}

func (a *AppIperf) populate_im() {
	for _, t := range a.tasks {
		task := common.Task{
			AppID:    t.AppID,
			Location: t.Location,
		}
		a.interest_map[task] = t
		a.app_id = t.AppID
	}
}
