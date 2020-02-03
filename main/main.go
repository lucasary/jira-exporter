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

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/calmh/mole/ansi"
	"github.com/joho/godotenv"
)

func main() {
	lambdaFunc := os.Getenv("AWS_LAMBDA_FUNCTION_NAME")
	if len(lambdaFunc) > 0 {
		lambda.Start(start)
	} else {
		start()
	}
}

func start() {
	lambdaFunc := os.Getenv("AWS_LAMBDA_FUNCTION_NAME")
	if len(lambdaFunc) == 0 {
		err := godotenv.Load()
		if err != nil {
			log.Fatal("Error loading .env file")
		}
	}

	username := os.Getenv("JIRA_USER")
	password := os.Getenv("JIRA_PASS")
	host := os.Getenv("JIRA_HOST")

	weekday := time.Now().Weekday()
	var updated int

	var message string

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

	message += "\n\n"
	message += "Morning team,"
	message += "\n\n"

	message += getQuote()

	message += "\n\n"
	message += strings.Join(
		[]string{"Issues updated in the past", strconv.Itoa(updated), "hours:"},
		" ")
	message += "\n\n"

	query1 := `
		project = INF
		AND status changed AFTER -` + strconv.Itoa(updated) + `h
		AND Sprint is not EMPTY
		ORDER BY status DESC, updated DESC
	`

	jiraQuery(username, password, host, query1, &message)

	fmt.Println()
	fmt.Println("Issues stuck for the past", updated, "hours+:")
	fmt.Println()

	message += "\n\n"
	message += strings.Join(
		[]string{"Issues stuck for the past", strconv.Itoa(updated), "hours+:"},
		" ")
	message += "\n\n"

	query2 := `
		project = INF
		AND ( status = "In Progress" OR status = "Review" )
		AND not status changed AFTER -` + strconv.Itoa(updated) + `h
		AND Sprint is not EMPTY
		ORDER BY status DESC, updated DESC
	`

	jiraQuery(username, password, host, query2, &message)

	slack(message)

}

func slack(message string) {
	url2 := os.Getenv("SLACK_URL")
	httpClient2 := &http.Client{}
	payload := strings.NewReader("{\"text\": \"" + message + "\"}")
	req2, err := http.NewRequest("POST", url2, payload)
	res2, err := httpClient2.Do(req2)

	_, _ = err, res2
}

func jiraQuery(username, password, host, query string, messagePtr *string) {
	url := host + "?jql=" + url.QueryEscape(query)

	httpClient := &http.Client{}
	req, err := http.NewRequest("GET", url, nil)
	req.SetBasicAuth(username, password)
	res, err := httpClient.Do(req)

	display(res, messagePtr)

	if err != nil {
		log.Fatal(err)
	}
}

func display(res *http.Response, messagePtr *string) {
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

		*messagePtr += strings.Join(
			[]string{strconv.Itoa(index + 1), "*[" + strings.Replace(key, "-", " ", 1) + "]*", summary, "/", "*" + assigneeName + "*", "->", "*" + statusName + "*"},
			" ",
		)
		*messagePtr += "\n"
	}

	if err != nil {
		log.Fatal(err)
	}
}

func getQuote() string {
	quote := strings.Join([]string{"'", quoteQuery(), "'"}, "")
	return quote
}

func quoteQuery() string {
	url := os.Getenv("QUOTE_URL")

	httpClient := &http.Client{}
	req, err := http.NewRequest("GET", url, nil)
	res, err := httpClient.Do(req)

	value := parseQuote(res)

	if err != nil {
		log.Fatal(err)
	}

	return value
}

func parseQuote(res *http.Response) string {
	bodyText, err := ioutil.ReadAll(res.Body)

	type rawJSON = map[string]*json.RawMessage

	var body, contents rawJSON
	var quotes []rawJSON
	var quote string

	err = json.Unmarshal(bodyText, &body)
	err = json.Unmarshal(*body["contents"], &contents)
	err = json.Unmarshal(*contents["quotes"], &quotes)
	_quote := quotes[0]
	err = json.Unmarshal(*_quote["quote"], &quote)

	if err != nil {
		log.Fatal(err)
	}

	return quote
}
