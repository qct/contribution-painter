package logger

import "github.com/sirupsen/logrus"

// InitLogger initializes Logrus and configure options
func InitLogger() {
	logrus.SetFormatter(&logrus.TextFormatter{
		FullTimestamp:   true,
		TimestampFormat: "2006-01-02 15:04:05",
		DisableQuote:    true,
	})
	logrus.SetLevel(logrus.DebugLevel)
}
