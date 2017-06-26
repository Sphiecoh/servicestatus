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
	flag.StringVar(&config.SlackUrl, "SlackUrl", "https://hooks.slack.com/services/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", "Slack WebHook Url")
	flag.StringVar(&config.SlackChannel, "SlackChannel", "#general", "Slack channel")
	flag.StringVar(&config.SlackUser, "slackuser", "", "Slack username")
	flag.Parse()

	store, err := db.NewStore(path.Join(config.DbPath, "apimonitor.db"))
	if err != nil {
		logrus.Fatal(err)
	}
	defer func() {
		if err := store.Close(); err != nil {
			logrus.Error(err)
		}
	}()
	createerr := store.CreateBuckets()
	if createerr != nil {
		logrus.Fatal(createerr)
	}
	monitor := &monitor.Monitor{
		Store: store,
	}
	tests, err := monitor.GetAllTests()
	if err != nil {
		logrus.Fatal(err)

	}
	schedule := schedule.New(tests, store, config)
	schedulererror := schedule.Start()
	if schedulererror != nil {
		logrus.Fatalf("Failed to start scheduler %v", schedulererror)
	}

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
