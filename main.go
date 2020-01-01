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

	updated := 72

	fmt.Println("Issues updated in the past", updated, "hours:")

	query1 := `
		project = INF
		AND updated >= -` + strconv.Itoa(updated) + `h
		AND status changed
		ORDER BY status DESC, updated DESC
	`

	jiraQuery(username, password, host, query1)

	fmt.Println("Issues stuck for the past", updated, "hours:")

	query2 := `
		project = INF
		AND ( status = "In Progress" OR status = "In Review" )
		AND not status changed AFTER ` + strconv.Itoa(updated) + `h
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

	for _, issue := range issues {
		var fields, assignee, status rawJSON
		var summary, assigneeName, key, statusName string

		err = json.Unmarshal(*issue["fields"], &fields)
		err = json.Unmarshal(*fields["assignee"], &assignee)
		err = json.Unmarshal(*fields["status"], &status)

		err = json.Unmarshal(*issue["key"], &key)
		err = json.Unmarshal(*fields["summary"], &summary)
		err = json.Unmarshal(*status["name"], &statusName)
		err = json.Unmarshal(*assignee["displayName"], &assigneeName)

		fmt.Println(key, summary, statusName, assigneeName)
	}

	if err != nil {
		log.Fatal(err)
	}
}
