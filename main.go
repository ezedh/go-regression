package main

import (
	"flag"

	"github.com/ezegrosfeld/go-regression/service"
	"go.uber.org/zap"
)

func main() {
	path := flag.String("path", "./", "Path to the regression folder")

	//args := os.Args[1:]

	flag.Parse()

	/* 	if len(args) == 0 {
		panic("No argument provided")
	} */

	l, err := zap.NewProduction()
	if err != nil {
		panic(err)
	}

	s := service.NewService(l)
	err = s.GetRegression(*path)
	if err != nil {
		panic(err)
	}

	s.RunRegression()
	s.GenerateReport()
}
