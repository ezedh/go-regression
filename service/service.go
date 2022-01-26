package service

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/ezedh/go-regression/internal"
	"github.com/ezedh/go-regression/model"
	"github.com/google/go-cmp/cmp"
	"go.uber.org/zap"
)

type Service interface {
	GetRegression(path string) error
	RunRegression()
	GenerateReport()
}

type service struct {
	log    *zap.SugaredLogger
	regre  *model.Regression
	result *model.RegressionResult
}

func NewService(log *zap.Logger) Service {
	return &service{
		log: log.Sugar(),
	}
}

func (s *service) GetRegression(path string) error {
	s.log.Infow("Getting regression", "path", path)

	// Get config.json from path
	// Parse config.json to Regression struct

	regre := new(model.Regression)

	err := internal.ParseFromJSONFile(path+"/config.json", regre)
	if err != nil {
		s.log.Errorw("Error parsing config.json", "error", err)
		return err
	}

	// Get all the files in the regression folder
	files, err := os.ReadDir(path)
	if err != nil {
		s.log.Errorw("Error reading regression folder", "error", err)
		return err
	}

	var groups []model.Group

	// Loop through all the files
	for _, file := range files {
		if file.Name() == "config.json" {
			continue
		}

		if strings.HasSuffix(file.Name(), ".json") {
			// Parse the file to a Test struct
			var tests []model.Test

			err := internal.ParseFromJSONFile(path+"/"+file.Name(), &tests)
			if err != nil {
				s.log.Errorw("Error parsing test file", "error", err)
				return err
			}

			// set group as the file name without the .json
			g := strings.Replace(file.Name(), ".json", "", -1)

			// Add the group to the groups
			groups = append(groups, model.Group{
				Name:  g,
				Tests: tests,
			})

			s.log.Infow("Parsed test file", "file", file.Name())
		}
	}

	regre.Groups = groups

	s.regre = regre

	return nil
}

// Run the regression tests using goroutines
func (s *service) RunRegression() {
	s.log.Infow("Running regression")

	result := new(model.RegressionResult)
	result.Name = s.regre.Name

	// Create channel for results
	ch := make(chan model.GroupResult, len(s.regre.Groups))
	// Create channel for errors

	// Loop through all the tests
	for _, test := range s.regre.Groups {
		t := test
		go s.runGroup(ch, t)
	}

	var groupResults []model.GroupResult
	// Wait for all the goroutines to finish

	// listen the channel for results
	for i := 0; i < len(s.regre.Groups); i++ {
		c := <-ch
		groupResults = append(groupResults, c)
	}

	close(ch)

	result.Results = groupResults

	t, p, f := s.getTotalsFromGroups(groupResults)

	result.Total = t
	result.Passed = p
	result.Failed = f

	s.result = result

}

func (s *service) GenerateReport() {
	s.log.Infow("Generating report")

	// Create report folder
	err := os.MkdirAll("./report", 0755)
	if err != nil {
		s.log.Errorw("Error creating report folder", "error", err)
		return
	}

	// create name_report.json file
	reportFile, err := os.Create("./report/" + strings.Replace(strings.ToLower(s.regre.Name), " ", "_", -1) + "_report.json")
	if err != nil {
		s.log.Errorw("Error creating report file", "error", err)
		return
	}

	// Marshal the result to json
	json.NewEncoder(reportFile).Encode(s.result)

	// Close the file
	reportFile.Close()

	s.log.Infow("Finished generating report")
}

func (s *service) runSingleTest(test model.Test) *model.TestResult {
	s.log.Infow("Running test", "test", test.Name, "endpoint", test.Endpoint, "method", test.Method)
	result := new(model.TestResult)
	result.Name = test.Name
	result.Path = test.Endpoint
	result.Subgroup = test.Subgroup

	// test.Body to io.Reader for request
	b, _ := json.Marshal(test.Body)

	resp, r, err := executeRequest(s.regre.BaseURL+test.Endpoint, string(b), test.Method, []http.Header{s.regre.Header, test.Header})
	if err != nil {
		fmt.Println(err.Error())
		result.Pass = false
		result.Error = err.Error()
		return result
	}

	pass := true
	var cause model.Cause

	if resp.StatusCode != test.ExpectedStatus {
		cause = model.StatusMissmatch
		pass = false
	}

	if test.ExpectedBody != nil && !cmp.Equal(r, test.ExpectedBody) {
		cause = model.BodyMissmatch
		pass = false
	}

	result.Pass = pass

	if !pass {
		result.Cause = cause
		switch cause {
		case model.StatusMissmatch:
			result.Expected = test.ExpectedStatus
			result.Actual = resp.StatusCode
		case model.BodyMissmatch:
			result.Expected = test.ExpectedBody
			result.Actual = r
		}
	}

	return result
}

func (s *service) runGroup(ch chan model.GroupResult, group model.Group) {
	s.log.Infow("Running group", "group", group.Name)
	result := new(model.GroupResult)
	result.Name = group.Name

	for _, test := range group.Tests {
		result.Total++
		t := *s.runSingleTest(test)

		if t.Pass {
			result.Passed++
		} else {
			result.Failed++
		}

		result.Results = append(result.Results, t)
	}

	s.createSubgroupResults(result)

	ch <- *result
}

func (s *service) createSubgroupResults(gr *model.GroupResult) {
	s.log.Infow("Creating subgroup results", "group", gr.Name)
	var sgr []model.SubgroupResult

	tm := make(map[string]int)
	pm := make(map[string]int)
	fm := make(map[string]int)

	for _, r := range gr.Results {
		group := r.Subgroup

		tm[group] += 1
		if r.Pass {
			pm[group] += 1
		} else {
			fm[group] += 1
		}
	}

	for k, v := range tm {
		sgr = append(sgr, model.SubgroupResult{
			Name:   k,
			Total:  v,
			Passed: pm[k],
			Failed: fm[k],
		})
	}

	gr.SubgroupResults = sgr
}

func (s *service) getTotalsFromGroups(groups []model.GroupResult) (int, int, int) {
	s.log.Infow("Getting totals from groups")
	var t, p, f int

	for _, g := range groups {
		t += g.Total
		p += g.Passed
		f += g.Failed
	}

	return t, p, f
}

func executeRequest(url, body, method string, headers []http.Header) (*http.Response, map[string]interface{}, error) {
	// Create a new client
	client := http.Client{}

	// Create a new request
	req, err := http.NewRequest(method, url, strings.NewReader(body))
	if err != nil {
		return nil, nil, err
	}

	// merge test headers with regression headers
	for _, v := range headers {
		for kk, vv := range v {
			req.Header[kk] = vv
		}
	}

	// Send the request
	resp, err := client.Do(req)
	if err != nil {
		return nil, nil, err
	}
	defer resp.Body.Close()

	// Check if body was the expected (compare strings)
	var res map[string]interface{}

	// resp.Body to res
	d, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, nil, err
	}
	err = json.Unmarshal(d, &res)
	if err != nil {
		return nil, nil, err
	}

	return resp, res, nil
}
