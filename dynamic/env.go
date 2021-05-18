package dynamic

import (
	"github.com/mobius-scheduler/mobius/common"
	"math"
	"math/rand"
)

func CoordInBox(n int, corner [2]float64, deltax float64, app int) []common.TaskData {
	var tasks []common.TaskData
	for i := 0; i < n; i++ {
		samples := [2]float64{rand.Float64(), rand.Float64()}
		x := corner[0] + samples[0]*deltax
		y := corner[1] + samples[1]*deltax
		tasks = append(
			tasks,
			common.TaskData{
				AppID: app,
				Location: common.Location{
					Latitude:  math.Round(x*1e6) / 1e6,
					Longitude: math.Round(y*1e6) / 1e6,
				},
				Interest:        1.0,
				TaskTimeSeconds: 10.0,
			},
		)
	}
	return tasks
}
