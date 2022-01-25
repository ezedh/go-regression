package model

import "net/http"

type (
	Test struct {
		Name           string                 `json:"name"`
		Group          string                 `json:"group"`
		Endpoint       string                 `json:"endpoint"`
		Method         string                 `json:"method"`
		Body           map[string]interface{} `json:"body"`
		ExpectedStatus int                    `json:"expectedStatus"`
		ExpectedBody   map[string]interface{} `json:"expectedBody,omitempty"`
		Header         http.Header            `json:"header,omitempty"`
	}

	Regression struct {
		Name    string      `json:"name"`
		Tests   []Test      `json:"tests"`
		BaseURL string      `json:"baseURL"`
		Header  http.Header `json:"header,omitempty"`
		Sync    bool        `json:"async,omitempty"`
	}

	TestResult struct {
		Name     string      `json:"name"`
		Group    string      `json:"group,omitempty"`
		Path     string      `json:"path"`
		Pass     bool        `json:"pass"`
		Cause    Cause       `json:"cause,omitempty"`
		Expected interface{} `json:"expected,omitempty"`
		Actual   interface{} `json:"actual,omitempty"`
		Error    string      `json:"error,omitempty"`
	}

	GroupResult struct {
		Name    string       `json:"name"`
		Results []TestResult `json:"results"`
		Total   int          `json:"total"`
		Passed  int          `json:"passed"`
		Failed  int          `json:"failed"`
	}

	RegressionResult struct {
		Name    string        `json:"name"`
		Results []GroupResult `json:"results"`
		Total   int           `json:"total"`
		Passed  int           `json:"passed"`
		Failed  int           `json:"failed"`
	}

	Cause string
)

const (
	StatusMissmatch Cause = "Status code does not match"
	BodyMissmatch   Cause = "Body does not match"
)
