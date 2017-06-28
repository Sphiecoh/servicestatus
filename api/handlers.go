package api

import (
	"net/http"

	"encoding/json"

	"github.com/Sirupsen/logrus"
	"github.com/labstack/echo"
	"github.com/pkg/errors"
	"github.com/sphiecoh/apimonitor/db"
	"github.com/sphiecoh/apimonitor/schedule"
	"gopkg.in/robfig/cron.v2"
)

func CreateTest(c echo.Context) error {
	store := c.Get("store").(*db.Store)
	sc := c.Get("schedule").(*schedule.Scheduler)
	newtest := &db.ApiTest{}
	newtest.ID = db.GenerateID()
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

	schedule, err := cron.Parse(newtest.Cron)
	if err != nil {
		logrus.Error(errors.Wrapf(err, "Test has invalid cron %s (%s)", newtest.Cron, newtest.Name))
	}
	sc.Cron.Schedule(schedule, job)
	logrus.Infof("New test created %s", newtest.Name)
	return c.JSON(http.StatusCreated, nil)
}

func GetAllTests(c echo.Context) error {

	store := c.Get("store").(*db.Store)
	result, err := store.GetAllTests()
	if err != nil {
		return err
	}
	c.JSON(200, result)
	return nil
}
func GetTestResult(c echo.Context) error {
	id := c.Param("id")
	store := c.Get("store").(*db.Store)
	result, err := store.GetResultsByTest(id)
	if err != nil {
		return err
	}
	c.JSON(200, result)
	return nil
}
