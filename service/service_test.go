package service

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/ezedh/go-regression/internal"
	"github.com/ezedh/go-regression/model"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func setup() {
	// create test_go_regression folder inside tmp
	os.Mkdir("/tmp/test_go_regression", 0777)

	config_body := `{"name": "Example test","baseURL": "http://localhost:8080","header": {"Accept": ["application/json"]}}`
	f, err := os.Create("/tmp/test_go_regression/config.json")
	if err != nil {
		panic(err)
	}

	_, err = f.WriteString(config_body)
	if err != nil {
		panic(err)
	}

	f.Close()

	setupTestFile()
}

func setupTestFile() {
	test_body := `[{
		"name": "Test Buyer Creation",
		"endpoint": "/buyer",
		"subgroup": "buyer creation",
		"method": "POST",
		"body": {},
		"expectedStatus": 201,
		"expectedBody": {
			"id": "1",
			"name": "Test Buyer",
			"address": "Test Address",
			"phone": "1234567890",
			"email": "1"
		}
	}]`
	f, err := os.Create("/tmp/test_go_regression/buyer.json")
	if err != nil {
		panic(err)
	}

	_, err = f.WriteString(test_body)
	if err != nil {
		panic(err)
	}

	f.Close()
}

func setupService() *service {
	return &service{
		log: internal.NewLogger().Sugar(),
	}
}

func setupMockServer(endpoint string) http.Handler {
	// create gin server with mock handler
	gin.SetMode(gin.TestMode)
	r := gin.New()

	r.GET(endpoint, func(c *gin.Context) {
		c.JSON(200, gin.H{
			"test": "test",
		})
	})

	return r
}

func TestMain(m *testing.M) {
	setup()
	os.Exit(m.Run())
}

func TestGetRegression(t *testing.T) {
	s := setupService()

	err := s.GetRegression("/tmp/test_go_regression")

	a := assert.New(t)

	a.NoError(err)
	a.NotNil(s.regre)
	a.Len(s.regre.Groups, 1)
	a.Len(s.regre.Groups[0].Tests, 1)
	a.Equal("Test Buyer Creation", s.regre.Groups[0].Tests[0].Name)
	a.Equal("/buyer", s.regre.Groups[0].Tests[0].Endpoint)
	a.Equal("buyer creation", s.regre.Groups[0].Tests[0].Subgroup)
}

func TestFailToGetRegression(t *testing.T) {
	s := setupService()

	err := s.GetRegression("/tmp/test_go_regression_noexists/")

	a := assert.New(t)

	a.Error(err)
	a.Nil(s.regre)
}

func TestRunRegression(t *testing.T) {
	s := setupService()

	err := s.GetRegression("/tmp/test_go_regression")
	a := assert.New(t)

	a.NoError(err)

	s.RunRegression()

	a.Equal(1, s.result.Total)
}

func TestGenerateReport(t *testing.T) {
	s := setupService()

	s.regre = &model.Regression{
		Name: "Test Regre",
	}

	s.GenerateReport()

	a := assert.New(t)

	// check if ./report folder exists
	_, err := os.Stat("./report")
	a.NoError(err)

	// check if ./report/test_regre.json exists
	_, err = os.Stat("./report/test_regre_report.json")
	a.NoError(err)

	// delete report folder
	err = os.RemoveAll("./report")
	a.NoError(err)
}

func TestRunSingleTest(t *testing.T) {
	a := assert.New(t)
	test := model.Test{
		Name:     "Test",
		Subgroup: "subgroup",
		Endpoint: "/test",
		Method:   "GET",
		ExpectedBody: map[string]interface{}{
			"test": "test",
		},
		ExpectedStatus: 200,
	}

	s := setupService()
	err := s.GetRegression("/tmp/test_go_regression")
	a.NoError(err)

	h := setupMockServer("/test")
	srv := httptest.NewServer(h)
	s.regre.BaseURL = srv.URL
	defer srv.Close()

	res := s.runSingleTest(test)

	a.True(res.Pass)
}

func TestSingleTestNotPassMissmatchedStatus(t *testing.T) {
	a := assert.New(t)
	test := model.Test{
		Name:           "Test",
		Subgroup:       "subgroup",
		Endpoint:       "/test",
		Method:         "GET",
		ExpectedStatus: 201,
	}

	s := setupService()
	err := s.GetRegression("/tmp/test_go_regression")
	a.NoError(err)

	h := setupMockServer("/test")
	srv := httptest.NewServer(h)
	s.regre.BaseURL = srv.URL
	defer srv.Close()

	res := s.runSingleTest(test)

	a.False(res.Pass)
	a.Equal(model.StatusMissmatch, res.Cause)
	a.Equal(201, res.Expected)
	a.Equal(200, res.Actual)
}

func TestSingleTestNotPassMissmatchedBody(t *testing.T) {
	a := assert.New(t)
	test := model.Test{
		Name:     "Test",
		Subgroup: "subgroup",
		Endpoint: "/test",
		Method:   "GET",
		ExpectedBody: map[string]interface{}{
			"test": "test1",
		},
		ExpectedStatus: 200,
	}

	s := setupService()
	err := s.GetRegression("/tmp/test_go_regression")
	a.NoError(err)

	h := setupMockServer("/test")
	srv := httptest.NewServer(h)
	s.regre.BaseURL = srv.URL
	defer srv.Close()

	res := s.runSingleTest(test)

	a.False(res.Pass)
	a.Equal(model.BodyMissmatch, res.Cause)
	a.Equal(test.ExpectedBody, res.Expected)
	a.Equal(map[string]interface{}{"test": "test"}, res.Actual)
}

func TestExecuteFails404(t *testing.T) {
	a := assert.New(t)
	test := model.Test{
		Name:           "Test",
		Subgroup:       "subgroup",
		Endpoint:       "/test",
		Method:         "GET",
		ExpectedStatus: 201,
	}

	s := setupService()
	err := s.GetRegression("/tmp/test_go_regression")
	a.NoError(err)

	h := setupMockServer("/whatever")
	srv := httptest.NewServer(h)
	s.regre.BaseURL = srv.URL
	defer srv.Close()

	res := s.runSingleTest(test)

	a.False(res.Pass)
	a.Equal(model.StatusMissmatch, res.Cause)
	a.Equal(201, res.Expected)
	a.Equal(404, res.Actual)
}
