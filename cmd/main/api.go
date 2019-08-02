package main

import (
	"github.com/netology-group/ulms-meta/pkg/app"
	"github.com/sirupsen/logrus"
	"github.com/spf13/pflag"
)

var (
	config    string
	port      string
	logFormat string
)

func init() {
	pflag.StringVarP(&config, "config", "c", "configs/default.yml", "path to config file")
	pflag.StringVarP(&port, "port", "p", ":8000", "interface:port to listen on")
	pflag.StringVar(&logFormat, "log-format", "text", "log format (text|json)")
}

func main() {
	pflag.Parse()
	setLogFormat()
	if err := app.New(config).Run(port); err != nil {
		logrus.WithError(err).Fatal("can't start or gracefully stop application")
	} else {
		logrus.Info("graceful shutdown completed")
	}
}

func setLogFormat() {
	var formatter logrus.Formatter
	switch logFormat {
	case "text":
		formatter = &logrus.TextFormatter{}
	case "json":
		formatter = &logrus.JSONFormatter{}
	default:
		logrus.Fatalf("Unknown log format: %v", logFormat)
	}
	logrus.SetFormatter(formatter)
}
