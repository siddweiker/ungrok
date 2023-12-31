//go:build ignore

// This program generates grok-patterns
// You can override the default url using the "-url" argument
package main

import (
	"flag"
	"io"
	"log"
	"net/http"
	"os"
	"text/template"
	"time"
)

var packageTemplate = template.Must(template.New("").Parse(`# Code generated by go generate; DO NOT EDIT.
# This file was generated by robots on {{ .Timestamp }}
# Downloaded from: {{ .URL }}
{{ .Data }}`))

func main() {
	url := flag.String(
		"url",
		"https://raw.githubusercontent.com/logstash-plugins/logstash-patterns-core/main/patterns/ecs-v1/grok-patterns",
		"The URL to download grok patterns from.",
	)
	flag.Parse()

	resp, err := http.Get(*url)
	if err != nil {
		log.Panic(err)
	}
	defer resp.Body.Close()

	f, err := os.Create("grok-patterns")
	if err != nil {
		log.Panic(err)
	}
	defer f.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Panic(err)
	}

	packageTemplate.Execute(f, struct {
		Timestamp string
		URL       string
		Data      string
	}{
		Timestamp: time.Now().Format(time.DateOnly),
		URL:       *url,
		Data:      string(data),
	})
}
