package api

import (
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/labstack/echo"
	mid "github.com/labstack/echo/middleware"
	"github.com/sphiecoh/apimonitor/conf"
	"github.com/sphiecoh/apimonitor/db"
	"github.com/sphiecoh/apimonitor/middleware"
	"github.com/sphiecoh/apimonitor/schedule"
)

type Server struct {
	Config   *conf.Config
	DB       *db.Store
	Schedule *schedule.Scheduler
}

func (s *Server) Start() {
	server := echo.New()
	server.Server.ReadTimeout = time.Second * 5
	server.Server.WriteTimeout = time.Second * 10
	server.Use(mid.Logger())
	server.Use(mid.Recover())
	server.Use(middleware.WithDataStore(s.DB))
	server.Use(middleware.WithScheduler(s.Schedule))
	server.POST("/", CreateTest)
	server.GET("/", GetAll)
	logrus.Fatal(server.Start(s.Config.Port))
}
