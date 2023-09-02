package main

import (
	"rewriting-history/configs"
	"rewriting-history/internal/app/stat"
	"rewriting-history/internal/pkg/logger"

	"github.com/sirupsen/logrus"
)

var config configs.Config

func main() {
	logger.InitLogger()
	err := configs.LoadConfig("", &config)
	if err != nil {
		logrus.Fatalf("Load config failed: %v", err)
	}

	re := stat.NewStat(&config)
	err = re.Run()
	if err != nil {
		logrus.Fatalf("Rewriter failed to run: %v", err)
	}
}
