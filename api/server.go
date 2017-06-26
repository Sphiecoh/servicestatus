package api

import (
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
	"github.com/sphiecoh/servicestatus/conf"
	"github.com/sphiecoh/servicestatus/db"
	"github.com/sphiecoh/servicestatus/schedule"
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
	server.Use(middleware.Logger())
	server.Use(middleware.Recover())
	server.Use(WithDataStore(s.DB))
	server.Use(WithScheduler(s.Schedule))
	server.POST("/", CreateTest)
	logrus.Fatal(server.Start(s.Config.Port))
}
