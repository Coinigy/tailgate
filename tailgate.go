// tailgate.go
// a log scanner -> slack

package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"hash/crc32"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
)

var path = flag.String("path", "", "path of the file to monitor")
var match = flag.String("match", "PHP message", "send alert if file matches that string")
var httpserver = flag.Bool("httpserver", false, "Http server which serves the error log on http port 8080")
var apikey = flag.String("apikey", "Slack API Key", "https://hooks.slack.com/services/YOUR-SLACK-TOKEN-HERE")
var channel = flag.String("channel", "#development", "Default channel where to post log messages")
var lastPos = 0
var hash uint32

func reader(path string) []string {
	file, err := os.Open(path)
	if err != nil {
		log.Fatalln(err)
	}
	defer file.Close()

	if _, err = file.Seek(0, os.SEEK_CUR); err != nil {
		log.Fatalln(err)
	}

	s := bufio.NewScanner(file)
	var lines []string
	for s.Scan() {
		lines = append(lines, s.Text())
	}
	return lines
}

func handler(w http.ResponseWriter, r *http.Request) {
	lines := reader(*path)
	nl := len(lines) - 1
	ul := 0
	if nl > 100 {
		ul = nl - 100
	}
	for i := nl; i >= ul; i-- {
		fmt.Fprintln(w, lines[i])
	}
}

func main() {
	flag.Parse()
	if _, err := os.Stat(*path); os.IsNotExist(err) {
		log.Fatalln("No such file or directory:", *path)
	}
	if *httpserver == true {
		log.Println("Starting httpserver on 8080")
		http.HandleFunc("/", handler)
		http.ListenAndServe(":8080", nil)
	}
	for {
		errorCheck()
		time.Sleep(time.Second)
	}
}

func errorCheck() {
	lines := reader(*path)
	nl := len(lines) - 1
	if nl == -1 {
		return
	}
	if lastPos == nl {
		return
	}
	for i := nl; i >= lastPos; i-- {
		if strings.Contains(lines[i], *match) {
			/* Compute a hash of last error and compare it to previous one */
			h := crc32.NewIEEE()
			h.Write([]byte(lines[i]))
			newHash := h.Sum32()
			if newHash != hash {
				slackLine(lines[i])
			}
			hash = newHash
			break
		}
	}
	lastPos = nl
	return
}

func slackLine(line string) error {
	type Message struct {
		Text     string `json:"text"`
		Username string `json:"username"`
		Channel  string `json:"channel"`
	}
	m := Message{line, "Error Log Scanner", *channel}
	url := "https://hooks.slack.com/services/" + *apikey

	j, _ := json.Marshal(m)
	b := bytes.NewReader(j)
	_, err := http.Post(url, "application/json", b)
	return err
}
