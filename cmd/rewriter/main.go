package main

import (
	"rewriting-history/configs"
	"rewriting-history/internal/app/rewriter"
	"rewriting-history/internal/pkg/logger"

	"github.com/sirupsen/logrus"
)

var config configs.Configuration

func main() {
	logger.InitLogger()
	err := configs.LoadConfig("", &config)
	if err != nil {
		logrus.Fatalf("Load config failed: %v", err)
	}

	re := rewriter.NewRewriter(config)
	err = re.Run()
	if err != nil {
		logrus.Fatalf("Rewriter failed to run: %v", err)
	}
}
