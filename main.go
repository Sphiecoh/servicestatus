package main

import (
	"flag"
	"os"
	"os/signal"
	"path"
	"syscall"

	"github.com/Sirupsen/logrus"
	"github.com/sphiecoh/apimonitor/api"
	"github.com/sphiecoh/apimonitor/conf"
	"github.com/sphiecoh/apimonitor/db"
	"github.com/sphiecoh/apimonitor/monitor"
	"github.com/sphiecoh/apimonitor/schedule"
)

func main() {
	config := &conf.Config{}
	flag.StringVar(&config.Port, "Port", ":8090", "HTTP port to listen on")
	flag.StringVar(&config.DbPath, "DataPath", "", "db dir")
	flag.Parse()

	store, err := db.Open(path.Join(config.DbPath, "apimonitor.db"))
	if err != nil {
		logrus.Fatal(err)
	}
	store.NewBucket([]byte("tests"))
	store.NewBucket([]byte("results"))
	monitor := &monitor.Monitor{
		Store:            store,
		ResultBucketName: "results",
		TestBucketName:   "tests",
	}
	tests, err := monitor.GetAllTests()
	if err != nil {
		logrus.Fatal(err)

	}
	schedule := schedule.New(tests, store)
	schedule.Start()

	srv := &api.Server{
		DB:       store,
		Config:   config,
		Schedule: schedule,
	}
	go srv.Start()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	sig := <-sigChan

	logrus.Infof("Shutting down %v signal received", sig)

}
