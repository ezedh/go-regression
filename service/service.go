package service

import (
	"encoding/json"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/ezegrosfeld/go-regression/internal"
	"github.com/ezegrosfeld/go-regression/model"
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
			group := strings.Replace(file.Name(), ".json", "", -1)

			// Loop through all the tests
			for i := 0; i < len(tests); i++ {
				// Add the group to the test
				tests[i].Group = group
			}

			s.log.Infow("Parsed test file", "file", file.Name())
			regre.Tests = append(regre.Tests, tests...)
		}
	}

	s.regre = regre

	return nil
}

// Run the regression tests using goroutines
func (s *service) RunRegression() {
	s.log.Infow("Running regression")

	result := new(model.RegressionResult)
	result.Name = s.regre.Name

	// Create channel for results
	ch := make(chan model.TestResult, len(s.regre.Tests))
	// Create channel for errors

	// Loop through all the tests
	for _, test := range s.regre.Tests {
		t := test
		go s.runTest(ch, s.regre.BaseURL, s.regre.Header, t)
	}

	groupResults := make(map[string][]model.TestResult)
	// Wait for all the goroutines to finish
	//wg.Wait()
	// listen the channel for results
	for i := 0; i < len(s.regre.Tests); i++ {
		c := <-ch
		s.log.Infow("Received result", "result", c)
		groupResults[c.Group] = append(groupResults[c.Group], c)
	}

	close(ch)

	var groups []model.GroupResult

	// Loop through all the groups
	for group, results := range groupResults {
		g := new(model.GroupResult)
		g.Name = group
		g.Total = len(results)

		for _, r := range results {
			result.Total++
			r.Group = ""
			g.Results = append(g.Results, r)
			if r.Pass {
				g.Passed++
				result.Passed++
			} else {
				g.Failed++
				result.Failed++
			}
		}

		groups = append(groups, *g)
	}

	result.Results = groups
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

// Run test and return the result throw the channel
func (s *service) runTest(ch chan model.TestResult, url string, header http.Header, test model.Test) {
	s.log.Infow("Running test", "test", test.Name)

	result := new(model.TestResult)
	result.Name = test.Name
	result.Path = test.Endpoint
	result.Group = test.Group

	// Create a new client
	client := http.Client{}

	// test.Body to io.Reader for request
	b, _ := json.Marshal(test.Body)

	// Create a new request
	req, err := http.NewRequest(test.Method, url+test.Endpoint, strings.NewReader(string(b)))
	if err != nil {
		result.Pass = false
		result.Error = err.Error()
		ch <- *result
		return
	}

	// merge test headers with regression headers
	for k, v := range header {
		req.Header[k] = v
	}

	// Add test headers
	for k, v := range test.Header {
		req.Header[k] = v
	}

	// Send the request
	resp, err := client.Do(req)
	if err != nil {
		result.Pass = false
		result.Error = err.Error()
		ch <- *result
		return
	}
	defer resp.Body.Close()

	pass := true
	var cause model.Cause

	if resp.StatusCode != test.ExpectedStatus {
		cause = model.StatusMissmatch
		pass = false
	}

	// Check if body was the expected (compare strings)
	var res map[string]interface{}

	// resp.Body to res
	d, err := io.ReadAll(resp.Body)
	if err != nil {
		s.log.Errorw("Error reading response body", "error", err)
		result.Pass = false
		result.Error = err.Error()
		ch <- *result
		return
	}
	err = json.Unmarshal(d, &res)
	if err != nil {
		s.log.Errorw("Error unmarshalling body", "error", err)
		result.Pass = false
		result.Error = err.Error()
		ch <- *result
		return
	}

	if test.ExpectedBody != nil && !cmp.Equal(res, test.ExpectedBody) {
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
			result.Actual = res
		}
	}

	ch <- *result
}
