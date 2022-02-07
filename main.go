package main

import (
	"flag"
	"os"
	"strings"

	"github.com/ezedh/go-regression/internal"
	"github.com/ezedh/go-regression/service"
)

func main() {
	path := flag.String("path", "./", "Path to the regression folder")
	base := flag.String("base", "http://127.0.0.1:8080", "Base URL for the regression")

	flag.Parse()

	l := internal.NewLogger()

	s := service.NewService(l)
	err := s.GetRegression(*path)
	if err != nil {
		panic(err)
	}

	s.SetBaseURL(*base)
	s.RunRegression()
	s.GenerateReport(getMetadata())
}

func getMetadata() map[string]string {
	args := os.Args[1:]
	metadata := make(map[string]string)
	end := false

	for _, arg := range args {
		if arg == "--" {
			end = true
			continue
		}

		if strings.Contains(arg, "=") && end {
			split := strings.Split(arg, "=")
			metadata[split[0]] = split[1]
		}
	}

	return metadata
}
