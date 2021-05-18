package dynamic

import (
	"github.com/mobius-scheduler/apps/utils"
	"github.com/mobius-scheduler/mobius/app"
	"github.com/mobius-scheduler/mobius/common"
)

type AppDynamic struct {
	app          app.Application
	interest_map common.InterestMap
	app_id       int
	cfg          DynamicConfig
}

type DynamicConfig struct {
	Corner [2]float64 `json:"corner"`
	Spread float64    `json:"spread"`
	Num    int        `json:"num"`
}

// init app interest map
func (a *AppDynamic) Init(cfg app.AppConfig) {
	a.app_id = cfg.AppID

	utils.JSONToStruct(cfg.Config, &a.cfg)
	a.generate_tasks()
}

// get app id
func (a *AppDynamic) GetID() int {
	return a.app_id
}

// get current interest map
func (a *AppDynamic) GetInterestMap() common.InterestMap {
	return a.interest_map
}

// remove completed tasks from interest map
func (a *AppDynamic) Update(completed_tasks []common.TaskData, time int) {
	a.generate_tasks()
}

func (a *AppDynamic) generate_tasks() {
	a.interest_map = make(common.InterestMap)
	tasks := CoordInBox(a.cfg.Num, a.cfg.Corner, a.cfg.Spread, a.app_id)
	for _, t := range tasks {
		task := common.Task{
			AppID:       t.AppID,
			Location:    t.Location,
			RequestTime: t.RequestTime,
		}
		a.interest_map[task] = t
	}
}
