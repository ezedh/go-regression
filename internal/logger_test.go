package internal_test

import (
	"os"
	"testing"

	"github.com/ezedh/go-regression/internal"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestNewLogger(t *testing.T) {
	os.Setenv("REGRE_MOD", "")
	l := internal.NewLogger()

	assert.NotNil(t, l)
	ce := l.Check(zap.DebugLevel, "debug mode")
	assert.Nil(t, ce)
}

func TestNewLoggerWithDebugMode(t *testing.T) {
	os.Setenv("REGRE_MOD", "debug")
	l := internal.NewLogger()

	assert.NotNil(t, l)
	ce := l.Check(zap.DebugLevel, "debug mode")
	assert.NotNil(t, ce)
}
