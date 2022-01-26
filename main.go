package main

import (
	"flag"

	"github.com/ezegrosfeld/go-regression/internal"
	"github.com/ezegrosfeld/go-regression/service"
)

func main() {
	path := flag.String("path", "./", "Path to the regression folder")

	flag.Parse()

	l := internal.NewLogger()

	s := service.NewService(l)
	err := s.GetRegression(*path)
	if err != nil {
		panic(err)
	}

	s.RunRegression()
	s.GenerateReport()
}
