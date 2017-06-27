package api

import (
	"net/http"

	"encoding/json"

	"github.com/Sirupsen/logrus"
	"github.com/labstack/echo"
	"github.com/pkg/errors"
	"github.com/robfig/cron"
	"github.com/sphiecoh/apimonitor/db"
	"github.com/sphiecoh/apimonitor/schedule"
)

func CreateTest(c echo.Context) error {
	store := c.Get("store").(*db.Store)
	sc := c.Get("schedule").(*schedule.Scheduler)
	newtest := &db.ApiTest{}
	c.Bind(newtest)
	data, err := json.Marshal(newtest)
	if err != nil {
		logrus.Error(err)
		return err
	}
	if e := store.Put(newtest.Name, sc.Store.TestBucket, data); e != nil {
		logrus.Error(e)
		return e
	}
	job := schedule.ToJob(newtest, store, sc.Config)

	schedule, err := cron.ParseStandard(newtest.Cron)
	if err != nil {
		logrus.Error(errors.Wrapf(err, "Test has invalid cron %s (%s)", newtest.Cron, newtest.Name))
	}
	sc.Cron.Schedule(schedule, job)
	logrus.Infof("New test created %s", newtest.Name)
	return c.JSON(http.StatusCreated, nil)
}

func GetAll(c echo.Context) error {

	store := c.Get("store").(*db.Store)
	result, err := store.GetAllTests()
	if err != nil {
		return err
	}
	c.JSON(200, result)
	return nil
}
