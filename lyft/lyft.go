package lyft

import (
	"github.com/mobius-scheduler/apps/utils"
	"github.com/mobius-scheduler/mobius/app"
	"github.com/mobius-scheduler/mobius/common"
	log "github.com/sirupsen/logrus"
)

type AppLyft struct {
	app          app.Application
	interest_map common.InterestMap
	app_id       int
	tasks        []common.TaskData
	cfg          LyftConfig
}

type LyftConfig struct {
	TaskPath       string `json:"task_path"`
	UpdateInterval int    `json:"update_interval"`
}

// init app interest map
func (a *AppLyft) Init(cfg app.AppConfig) {
	a.app_id = cfg.AppID
	a.interest_map = make(common.InterestMap)

	utils.JSONToStruct(cfg.Config, &a.cfg)
	a.tasks = utils.LoadTasksFromFile(a.cfg.TaskPath)

	// add first batch of tasks
	a.add_tasks(0, a.cfg.UpdateInterval)
}

// get app id
func (a *AppLyft) GetID() int {
	return a.app_id
}

// get current interest map
func (a *AppLyft) GetInterestMap() common.InterestMap {
	return a.interest_map
}

// update app interest map
func (a *AppLyft) Update(completed_tasks []common.TaskData, time int) {
	var completed, expired int
	// remove completed tasks
	for _, t := range completed_tasks {
		task := t.GetTask()
		if _, exists := a.interest_map[task]; exists {
			completed += 1
			var x = a.interest_map[task]
			x.Interest -= t.Interest
			x.TaskTimeSeconds -= t.TaskTimeSeconds
			a.interest_map[task] = x
			if a.interest_map[task].Interest < 0 {
				log.Fatalf("[app] interest cannot be < 0: %v %v", t, a.interest_map[task])
			}
			if a.interest_map[task].Interest == 0 {
				delete(a.interest_map, task)
			}
		} else {
			log.Warnf("[app] task %+v not in interest map (cannot delete)", t.GetTask())
		}
	}

	// remove expired tasks
	for t, _ := range a.interest_map {
		if t.RequestTime < time {
			delete(a.interest_map, t)
			expired += 1
		}
	}

	log.Debugf(
		"[app] customer %d, %v tasks completed, %v tasks expired",
		a.app_id,
		completed,
		expired,
	)

	// add new tasks
	a.add_tasks(time-a.cfg.UpdateInterval, time)
}

// add tasks after start and before (including) end
func (a *AppLyft) add_tasks(start, end int) {
	var added int
	for _, t := range a.tasks {
		// ignore task if out of time range
		if !(t.RequestTime > start && t.RequestTime <= end) {
			continue
		}

		// add task
		task := t.GetTask()
		if _, exists := a.interest_map[task]; exists {
			var x = a.interest_map[task]
			x.TaskTimeSeconds += t.TaskTimeSeconds
			x.Interest += t.Interest
			a.interest_map[task] = x
		} else {
			a.interest_map[task] = t
		}
		added++
	}

	log.Debugf("[app] customer %d, %v new tasks", a.app_id, added)
}
