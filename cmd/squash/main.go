package main

import (
	"rewriting-history/configs"
	"rewriting-history/internal/app/squash"
	"rewriting-history/internal/pkg/logger"

	"github.com/sirupsen/logrus"
)

var config configs.Config

// Deprecated: This is a deprecated function.
func main() {
	logger.InitLogger()
	err := configs.LoadConfig("", &config)
	if err != nil {
		logrus.Fatalf("Load config failed: %v", err)
	}

	s := squash.NewSquasher(&config)
	err = s.Squash()
	if err != nil {
		logrus.Fatalf("Squash failed: %v", err)
	}
}
