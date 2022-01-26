package model

import "net/http"

type (
	Test struct {
		Name           string                 `json:"name"`
		Subgroup       string                 `json:"subgroup"`
		Endpoint       string                 `json:"endpoint"`
		Method         string                 `json:"method"`
		Body           map[string]interface{} `json:"body"`
		ExpectedStatus int                    `json:"expectedStatus"`
		ExpectedBody   map[string]interface{} `json:"expectedBody,omitempty"`
		Header         http.Header            `json:"header,omitempty"`
	}

	Group struct {
		Name  string `json:"name"`
		Tests []Test `json:"tests"`
	}

	Regression struct {
		Name    string      `json:"name"`
		Groups  []Group     `json:"groups"`
		BaseURL string      `json:"baseURL"`
		Header  http.Header `json:"header,omitempty"`
		Sync    bool        `json:"async,omitempty"`
	}

	TestResult struct {
		Name     string      `json:"name"`
		Group    string      `json:"-"`
		Subgroup string      `json:"-"`
		Path     string      `json:"path"`
		Pass     bool        `json:"pass"`
		Cause    Cause       `json:"cause,omitempty"`
		Expected interface{} `json:"expected,omitempty"`
		Actual   interface{} `json:"actual,omitempty"`
		Error    string      `json:"error,omitempty"`
	}

	SubgroupResult struct {
		Name   string `json:"name"`
		Total  int    `json:"total"`
		Passed int    `json:"passed"`
		Failed int    `json:"failed"`
	}

	GroupResult struct {
		Name            string           `json:"name"`
		Results         []TestResult     `json:"results"`
		SubgroupResults []SubgroupResult `json:"subgroup_results"`
		Total           int              `json:"total"`
		Passed          int              `json:"passed"`
		Failed          int              `json:"failed"`
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
