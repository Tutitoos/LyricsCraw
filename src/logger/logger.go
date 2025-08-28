package logger

import (
	"os"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var (
	Logger *zap.Logger
	Sugar  *zap.SugaredLogger
)

func InitLogger() {
	var err error

	appEnv := os.Getenv("APP_ENV")
	if appEnv == "production" {
		zapConfig := zap.NewProductionConfig()
		zapConfig.EncoderConfig.TimeKey = "timestamp"
		zapConfig.EncoderConfig.EncodeTime = func(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
			enc.AppendString(t.Format("2006-01-02 15:04:05"))
		}
		Logger, err = zapConfig.Build()

	} else {

		Logger, err = zap.NewDevelopment()
	}

	if err != nil {
		panic(err)
	}

	Sugar = Logger.Sugar()
	Sugar.Info("✅ Logger initialized")
}
