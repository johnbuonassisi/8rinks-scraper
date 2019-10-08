package main

import (
	"bytes"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
	"golang.org/x/net/html"
)

func init() {
	log.SetFormatter(&log.JSONFormatter{})
	log.SetLevel(log.DebugLevel)
}

func main() {

	var teamName = flag.String("tn", "Megpies FC", "Team name for which the schedule will be retrieved")
	flag.Parse()

	// First, navigate to the soccer schedule page
	resp, err := getSoccerSchedule()
	if err != nil {
		log.Errorf("%v", err)
		os.Exit(1)
	}

	// Read all of the response body and store in memory. Each processing function
	// should make a new io.ReadCloser of the response body to start at the beginning
	// of the response.
	bodyBytes, _ := ioutil.ReadAll(resp.Body)
	resp.Body.Close()

	// Then, find the current season the league is in
	resp.Body = ioutil.NopCloser(bytes.NewBuffer(bodyBytes)) // reset response body
	seasonID, err := getSeasonID(resp)
	if err != nil {
		log.Errorf("%v", err)
		os.Exit(1)
	}
	resp.Body.Close()
	log.Infof("Season ID: %s", seasonID)

	// Next, find the provided team in the dropdown
	resp.Body = ioutil.NopCloser(bytes.NewBuffer(bodyBytes)) // reset response body
	teamID, err := getTeamID(*teamName, resp)
	if err != nil {
		log.Errorf("%v", err)
		os.Exit(1)
	}
	resp.Body.Close()
	log.Infof("Team Name: %s", *teamName)
	log.Infof("Team ID: %s", teamID)

	// Finally, press the Go button to get the team's games
	resp.Body = ioutil.NopCloser(bytes.NewBuffer(bodyBytes)) // reset response body
	games, err := getGames(teamID, seasonID, resp)
	if err != nil {
		log.Errorf("%v", err)
		os.Exit(1)
	}

	log.Infof("Games: %v", games)
}

type game struct {
	startTime time.Time
	opponent  string
	location  string
}

func getGames(teamID string, seasonID string, response *http.Response) ([]game, error) {
	return nil, nil
}

func getSoccerSchedule() (*http.Response, error) {

	req, err := http.NewRequest("GET", "https://canlanaisl.icesports.com/BURNABY8RINKS/soccer-schedule.aspx", nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Host", "canlanaisl.icesports.com")
	req.Header.Set("Connection", "keep-alive")
	req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_14_0) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/75.0.3770.142 Safari/537.36")
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3")
	req.Header.Set("Accept-Encoding", "gzip, deflate")
	req.Header.Set("Accept-Language", "en-US,en;q=0.9")

	client := http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return resp, err
	}

	return resp, nil
}

func getSeasonID(scheduleResp *http.Response) (string, error) {

	log.Debug("getSeasonID: trying to find current season")

	z := html.NewTokenizer(scheduleResp.Body)
	for {
		tt := z.Next()
		switch tt {
		case html.ErrorToken:
			log.Fatal(z.Err())
		case html.StartTagToken:
			// Find select tags
			name, hasAttr := z.TagName()
			if string(name) == "select" && hasAttr {
				// Find the attribute that identifies a season selector
				key, val, _ := z.TagAttr()
				if string(key) == "name" && string(val) == "ctl00$mainContent$ctl01$ddlSeason" {
					log.Debug("getSeasonID: FOUND SEASONS!")
					// Find the season we are looking for in the next tokens
					for {
						tt = z.Next()
						name, _ = z.TagName()
						// Finish if all teams have been iterated through
						if tt == html.EndTagToken && string(name) == "select" {
							log.Debug("getSeasonID: FINISHED SEASONS!")
							return "", fmt.Errorf("No current season found")
						}
						// Save the seasonId specified in the tag
						if tt == html.StartTagToken && string(name) == "option" {
							_, selected, moreAttr := z.TagAttr()
							// Check for the selected season, which is assumed to be the current one
							if string(selected) == "selected" && moreAttr {
								log.Debug("getSeasonID: Found selected season")
								// Get the seasonId, assuming it is the next attribute
								_, seasonID, _ := z.TagAttr()
								log.Debugf("getSeasonID: Found seasonID, %s", seasonID)
								return string(seasonID), nil
							}
						}
					}
				}
			}
		}
	}
}

func getTeamID(teamName string, scheduleResp *http.Response) (string, error) {

	log.Debugf("getTeamID: trying to find %s", teamName)

	z := html.NewTokenizer(scheduleResp.Body)
	for {
		tt := z.Next()
		switch tt {
		case html.ErrorToken:
			log.Fatal(z.Err())
		case html.StartTagToken:
			// Find select tags
			name, hasAttr := z.TagName()
			if string(name) == "select" && hasAttr {
				// Find the attribute that identifies a team selector
				key, val, _ := z.TagAttr()
				if string(key) == "name" && string(val) == "ctl00$mainContent$ctl01$ddlTeams" {
					log.Debug("getTeamId: FOUND TEAMS!")
					// Find the team we are looking for in the next tokens
					var teamID []byte
					for {
						tt = z.Next()
						name, _ = z.TagName()
						// Finish if all teams have been iterated through
						if tt == html.EndTagToken && string(name) == "select" {
							log.Debug("getTeamId: FINISHED TEAMS!")
							return "", fmt.Errorf("No team found matching %s", teamName)
						}
						// Save the teamId specified in the tag
						if tt == html.StartTagToken && string(name) == "option" {
							_, teamID, _ = z.TagAttr()
						}
						// Check if the text token matches your team
						if tt == html.TextToken {
							text := z.Text()
							if string(text) == teamName {
								log.Debugf("getTeamId: Found your team: %s", string(text))
								if len(teamID) == 0 {
									log.Debug("getTeamId: Could not find Team Id")
								}
								log.Debugf("getTeamId: Team Id: %s", string(teamID))
								return string(teamID), nil
							}
						}
					}
				}
			}
		}
	}
}

/*
	file, err := ioutil.ReadFile("example.xml") // For read access.
	if err != nil {
		log.Fatal(err)
	}

	s := string(file)

	// Find the dates in order
	dateIdxs := findDaysOfWeek(s)
	var dateStrs []string
	for _, dateIdx := range dateIdxs {
		dateEndIdx := strings.Index(s[dateIdx:], "<")
		dateStr := s[dateIdx : dateIdx+dateEndIdx]
		dateStrs = append(dateStrs, dateStr)
	}

	// Find the times in order
	timeIdxs := findTime(s, dateIdxs)
	var timeStrs []string
	for _, timeIdx := range timeIdxs {
		timeStr := s[timeIdx-6 : timeIdx+2]
		timeStrs = append(timeStrs, timeStr)
	}

	// Parse dates and times to time.Time
	timeLayout := "Monday, January 2, 2006, 03:04 PM MST"
	for i, dateStr := range dateStrs {
		dateTimeStr := dateStr + ", " + timeStrs[i] + " PST"
		t, err := time.Parse(timeLayout, dateTimeStr)
		if err != nil {
			fmt.Printf("ERROR parsing times, %d, %v \n", i, err)
			return
		}
		fmt.Printf("time: %v \n", t)
	}
*/

func findTime(s string, daysIdxs []int) []int {
	var idxs []int
	for _, val := range daysIdxs {
		idx := strings.Index(s[val:], "PM")
		idxs = append(idxs, idx+val)
	}
	return idxs
}

func findDaysOfWeek(s string) []int {
	daysOfWeek := []string{"Monday", "Tuesday", "Wednesday", "Thursday", "Friday", "Saturday", "Sunday"}
	var idxs []int
	for _, day := range daysOfWeek {
		t := s
		var dayIdxs []int
		for occ := 0; ; occ++ {
			idx := strings.Index(t, day)
			if idx == -1 {
				break
			}
			var prev int
			if occ > 0 {
				prev = dayIdxs[occ-1] + 1
			}
			dayIdxs = append(dayIdxs, idx+prev)
			t = t[idx+1:]
		}
		idxs = append(idxs, dayIdxs...)
	}
	return idxs
}
