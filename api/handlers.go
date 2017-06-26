package api

import (
	"net/http"

	"encoding/json"

	"github.com/labstack/echo"
	"github.com/robfig/cron"
	"github.com/sphiecoh/servicestatus/db"
	"github.com/sphiecoh/servicestatus/monitor"
	"github.com/sphiecoh/servicestatus/schedule"
)

func CreateTest(c echo.Context) error {
	store := c.Get("store").(*db.Store)
	sc := c.Get("schedule").(*schedule.Scheduler)
	newtest := &monitor.ApiTest{}
	c.Bind(newtest)
	data, err := json.Marshal(newtest)
	if err != nil {
		return err
	}
	if e := store.Put("tests", newtest.Name, data); e != nil {
		return e
	}
	job := schedule.ToJob(newtest, store)

	schedule, err := cron.ParseStandard(newtest.Cron)
	sc.Cron.Schedule(schedule, job)
	return c.JSON(http.StatusCreated, nil)
}
func WithDataStore(store *db.Store) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) (err error) {
			c.Set("store", store)
			return next(c)
		}
	}
}
func WithScheduler(schedule *schedule.Scheduler) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) (err error) {
			c.Set("schedule", schedule)
			return next(c)
		}
	}
}
