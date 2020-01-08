package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/calmh/mole/ansi"
	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	username := os.Getenv("JIRA_USER")
	password := os.Getenv("JIRA_PASS")
	host := os.Getenv("JIRA_HOST")

	weekday := time.Now().Weekday()
	var updated int

	if len(os.Args) == 2 {
		updated, _ = strconv.Atoi(os.Args[1])
		fmt.Println("Manual override: scanning over the last", updated, "hours.")
	} else if weekday.String() == "Monday" {
		updated = 72
		fmt.Println("Today is a", weekday, "so I will scan over the last", updated, "hours.")
	} else {
		updated = 24
		fmt.Println("Default:", updated, "hours")
	}

	fmt.Println()
	fmt.Println("Issues updated in the past", updated, "hours:")
	fmt.Println()

	query1 := `
		project = INF
		AND status changed AFTER -` + strconv.Itoa(updated) + `h
		ORDER BY status DESC, updated DESC
	`

	jiraQuery(username, password, host, query1)

	fmt.Println()
	fmt.Println("Issues stuck for the past", updated, "hours+:")
	fmt.Println()

	query2 := `
		project = INF
		AND ( status = "In Progress" OR status = "Review" )
		AND not status changed AFTER -` + strconv.Itoa(updated) + `h
		ORDER BY status DESC, updated DESC
	`

	jiraQuery(username, password, host, query2)
}

func jiraQuery(username, password, host, query string) {
	url := host + "?jql=" + url.QueryEscape(query)

	httpClient := &http.Client{}
	req, err := http.NewRequest("GET", url, nil)
	req.SetBasicAuth(username, password)
	res, err := httpClient.Do(req)

	display(res)

	if err != nil {
		log.Fatal(err)
	}
}

func display(res *http.Response) {
	bodyText, err := ioutil.ReadAll(res.Body)
	// s := string(bodyText)

	type rawJSON = map[string]*json.RawMessage

	var body rawJSON
	var issues []rawJSON

	err = json.Unmarshal(bodyText, &body)
	err = json.Unmarshal(*body["issues"], &issues)

	if len(issues) == 0 {
		return
	}
	for index, issue := range issues {
		var fields, assignee, status rawJSON
		var summary, assigneeName, key, statusName string

		if issue["fields"] != nil {
			err = json.Unmarshal(*issue["fields"], &fields)
		}
		if fields["assignee"] != nil {
			err = json.Unmarshal(*fields["assignee"], &assignee)
		}
		if fields["status"] != nil {
			err = json.Unmarshal(*fields["status"], &status)
		}
		if issue["key"] != nil {
			err = json.Unmarshal(*issue["key"], &key)
		}
		if fields["summary"] != nil {
			err = json.Unmarshal(*fields["summary"], &summary)
		}
		if status["name"] != nil {
			err = json.Unmarshal(*status["name"], &statusName)
		}
		if assignee["displayName"] != nil {
			err = json.Unmarshal(*assignee["displayName"], &assigneeName)
		}

		fmt.Println(index+1, ansi.Bold("["+strings.Replace(key, "-", " ", 1)+"]"), summary, "/", ansi.Bold(assigneeName), "->", ansi.Bold(statusName))
	}

	if err != nil {
		log.Fatal(err)
	}
}
