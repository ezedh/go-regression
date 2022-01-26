package internal_test

import (
	"testing"

	"github.com/ezedh/go-regression/internal"
	"github.com/stretchr/testify/assert"
)

func TestNewLogger(t *testing.T) {
	l := internal.NewLogger()

	assert.NotNil(t, l)
}
