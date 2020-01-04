## Requirements:

### Minimum

1- Get an API token for your account in order to programmatically access JIRA data: `https://id.atlassian.com/manage/api-tokens`

2- Create `.env` file using `.env.default` with your username (email) and API token from step #1.

3- Install Go dependencies:

``` sh
go get github.com/calmh/mole/ansi
go get github.com/joho/godotenv
```

### Optional: if you want to optain to html and maintain text formatting

4- Install aha (converts terminal ansi output to html) if you want to maintain the bold formatting: `brew install aha` (Linux: `sudo apt install aha`)

## Usage

### Basic usage (ouput results to terminal)

`go run main.go`

### Slack-formatted usage (create formatted output.htm and open with web browser)

Mac OSX: `go run main.go | aha -n | sed 's/$/<br>/' > output.htm && open output.htm`

Linux (Ubuntu): `go run main.go | aha -n | sed 's/$/<br>/' > output.htm && xdg-open output.htm`

You can then copy/pasta the formatted text into slack for the rest of the team :)
