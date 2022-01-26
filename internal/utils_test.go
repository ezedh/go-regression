package internal_test

import (
	"os"
	"testing"

	"github.com/ezegrosfeld/go-regression/internal"
	"github.com/stretchr/testify/assert"
)

func TestMain(m *testing.M) {
	setup()
	c := m.Run()
	clean()
	os.Exit(c)
}

func setup() {
	// create json file in tmp folder
	body := `{"a": 1, "b": 2}`

	f, err := os.Create("/tmp/test.json")
	if err != nil {
		panic(err)
	}

	_, err = f.WriteString(body)
	if err != nil {
		panic(err)
	}

	f.Close()
}

func clean() {
	// remove json file in tmp folder
	os.Remove("/tmp/test.json")
}

func TestParseFromJSONFile(t *testing.T) {
	type testStruct struct {
		A int `json:"a"`
		B int `json:"b"`
	}

	var s testStruct

	err := internal.ParseFromJSONFile("/tmp/test.json", &s)

	assert.NoError(t, err)
	assert.Equal(t, 1, s.A)
	assert.Equal(t, 2, s.B)
}

func TestFailToParseJSON(t *testing.T) {
	type testStruct struct {
		A string `json:"a"`
		B string `json:"b"`
	}

	var s testStruct

	err := internal.ParseFromJSONFile("/tmp/test.json", &s)

	assert.Error(t, err)
}

func TestFailToFindJSON(t *testing.T) {
	type testStruct struct {
		A string `json:"a"`
		B string `json:"b"`
	}

	var s testStruct

	err := internal.ParseFromJSONFile("/tmp/test_nonexist.json", &s)

	assert.Error(t, err)
}
