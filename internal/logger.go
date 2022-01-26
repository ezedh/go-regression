package internal

import (
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func NewLogger() *zap.Logger {
	// crate new custom logger without timestamp and caller
	encConfig := zap.NewProductionEncoderConfig()
	encConfig.CallerKey = ""
	encConfig.TimeKey = ""

	atom := zap.NewAtomicLevel()
	atom.SetLevel(zap.InfoLevel)

	return zap.New(zapcore.NewCore(
		zapcore.NewJSONEncoder(encConfig),
		zapcore.Lock(os.Stdout),
		atom,
	))
}
