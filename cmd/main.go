package main

import (
	"bytes"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
	"golang.org/x/net/html"
)

func init() {
	log.SetFormatter(&log.JSONFormatter{})
	log.SetLevel(log.InfoLevel)
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

	// Extract the __VIEWSTATES from the original response
	resp.Body = ioutil.NopCloser(bytes.NewBuffer(bodyBytes)) // reset response body
	viewStateInfo, err := getViewStates(resp)
	if err != nil {
		log.Errorf("%v", err)
		os.Exit(1)
	}

	// Finally, press the Go button to get the team's games
	resp.Body = ioutil.NopCloser(bytes.NewBuffer(bodyBytes)) // reset response body
	games, err := getGames(teamID, seasonID, viewStateInfo, resp.Cookies())
	if err != nil {
		log.Errorf("%v", err)
		os.Exit(1)
	}

	for _, g := range games {
		log.Infof("game: %+v", g)
	}

}

type game struct {
	StartTime time.Time
	Opponent  string
	Location  string
}

func getGames(teamID string, seasonID string, viewStateInfo ViewStateInfo, cookies []*http.Cookie) ([]game, error) {

	// Create the form
	form := url.Values{}
	form.Add("ctl00$ScriptManager1", "ctl00$mainContent$ctl01$UpdatePanel4|ctl00$mainContent$ctl01$btnGoF")
	form.Add("ctl00$mainContent$ctl01$ddlSeason", seasonID)
	form.Add("ctl00$mainContent$ctl01$ddlSeason_f", seasonID)
	form.Add("ctl00$mainContent$ctl01$ddlTeams", "0")
	form.Add("ctl00$mainContent$ctl01$ddlTeams_f", teamID)
	form.Add("ctl00$mainContent$ctl01$ddlDivisions", "0")
	form.Add("ctl00$mainContent$ctl01$ddlDivisions_f", "0")
	form.Add("__EVENTTARGET", "ctl00$mainContent$ctl01$btnGoF")
	form.Add("__VIEWSTATEGENERATOR", viewStateInfo.ViewStateGenerator)
	form.Add("__ASYNCPOST", "true")

	// Add the view states to the form
	form.Add("__VIEWSTATEFIELDCOUNT", strconv.Itoa(len(viewStateInfo.ViewStates)))
	for i := 0; i < len(viewStateInfo.ViewStates); i++ {
		var viewStateKey string
		if i == 0 {
			viewStateKey = "__VIEWSTATE"
		} else {
			viewStateKey = "__VIEWSTATE" + strconv.Itoa(i)
		}
		form.Add(viewStateKey, viewStateInfo.ViewStates[i])
	}
	form.Add("__EVENTVALIDATION", viewStateInfo.EventValidation)

	// Create the POST request
	req, err := http.NewRequest("POST",
		"https://canlanaisl.icesports.com/BURNABY8RINKS/soccer-schedule.aspx",
		strings.NewReader(form.Encode()))
	if err != nil {
		return nil, err
	}

	// Set the Header with standard fields
	req.Header.Set("Accept", "*/*")
	req.Header.Set("Accept-Encoding", "gzip, deflate")
	req.Header.Set("Accept-Language", "en-US,en;q=0.9")
	req.Header.Set("Connection", "keep-alive")
	req.Header.Set("Content-Length", "42586")
	req.Header.Set("Pragma", "no-cache")
	req.Header.Set("Server", "Microsoft-IIS/8.5")
	req.Header.Set("X-AspNet-Version", "4.0.30319")
	req.Header.Set("X-Powered-By", "ASP.NET")
	req.Header.Set("Date", time.Now().String())
	req.Header.Set("Expires", "-1")
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded; charset=UTF-8")
	req.Header.Set("Host", "canlanaisl.icesports.com")
	req.Header.Set("Origin", "https://canlanaisl.icesports.com")
	req.Header.Set("Referer", "https://canlanaisl.icesports.com/BURNABY8RINKS/soccer-schedule.aspx")
	req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_14_0) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/75.0.3770.142 Safari/537.36")
	req.Header.Set("X-MicrosoftAjax", "Delta=true")
	req.Header.Set("X-Requested-With", "XMLHttpRequest")

	for _, cookie := range cookies {
		req.AddCookie(cookie)
	}

	// Send the POST request
	client := http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	/*
		bodyBytes, _ := ioutil.ReadAll(resp.Body)

		log.Debugf("Response to post: %s", string(bodyBytes))

		s := string(bodyBytes)

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
		var games []game
		/*
			timeLayout := "Monday, January 2, 2006, 03:04 PM MST"
			for i, dateStr := range dateStrs {
				dateTimeStr := dateStr + ", " + timeStrs[i] + " PST"
				t, err := time.Parse(timeLayout, dateTimeStr)
				if err != nil {
					return nil, fmt.Errorf("error parsing times, %d, %v", i, err)
				}
				games = append(games, game{StartTime: t})
			}
	*/
	games := getAllGames(resp)

	return games, nil
}

func getAllGames(resp *http.Response) []game {

	var games []game
	z := html.NewTokenizer(resp.Body)
	for {
		tt := z.Next()
		switch tt {
		case html.ErrorToken:
			log.Info("error token")
			return nil
		case html.StartTagToken:
			name, hasAttr := z.TagName()
			if string(name) == "td" && hasAttr {
				key, _, _ := z.TagAttr()
				if string(key) == "colspan" {
					log.Infof("Found colspan")
					for {
						tk := z.Next()
						if tk == html.TextToken {
							text := z.Text()
							timeLayout := "  Monday, January 2, 2006, 03:04 PM MST"
							dateTimeStr := dateStr + ", " + timeStrs[i] + " PST"
								t, err := time.Parse(timeLayout, dateTimeStr)
								if err != nil {
									fmt.Printf("ERROR parsing times, %d, %v \n", i, err)
									return
								}
								fmt.Printf("time: %v \n", t)

							for i, dateStr := range dateStrs {
								dateTimeStr := dateStr + ", " + timeStrs[i] + " PST"
								t, err := time.Parse(timeLayout, dateTimeStr)
								if err != nil {
									fmt.Printf("ERROR parsing times, %d, %v \n", i, err)
									return
								}
								fmt.Printf("time: %v \n", t)
							}
							log.Infof("game date: %s", text)
							break
						}
					}
				}
			}
		}
	}
	return games
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

type ViewStateInfo struct {
	ViewStates         []string
	EventValidation    string
	ViewStateGenerator string
}

/*
	getViewStates looks for values in the section below, specifically the tags with name `input`
	that are of type `hidden`, and have name containing `__VIEWSTATE`

	<div class="aspNetHidden">
	<input type="hidden" name="__EVENTTARGET" id="__EVENTTARGET" value="" />
	<input type="hidden" name="__EVENTARGUMENT" id="__EVENTARGUMENT" value="" />
	<input type="hidden" name="__LASTFOCUS" id="__LASTFOCUS" value="" />
	<input type="hidden" name="__VIEWSTATEFIELDCOUNT" id="__VIEWSTATEFIELDCOUNT" value="28" />
	<input type="hidden" name="__VIEWSTATE" id="__VIEWSTATE" value="TVzIffnWuIz4onunMmzM
	<input type="hidden" name="__VIEWSTATE1" id="__VIEWSTATE1" value="wsUfB
	<input type="hidden" name="__VIEWSTATE2" id="__VIEWSTATE2" value="aVf
	<input type="hidden" name="__VIEWSTATE3" id="__VIEWSTATE3" value="Vn357AOxqrID0
*/
func getViewStates(resp *http.Response) (viewStateInfo ViewStateInfo, err error) {

	log.Debug("getViewStates: trying to find view states")
	z := html.NewTokenizer(resp.Body)
	for {
		tt := z.Next()
		switch tt {
		case html.ErrorToken:
			return
		case html.SelfClosingTagToken:
			// Find the input tags
			name, hasAttr := z.TagName()
			if string(name) == "input" && hasAttr {
				log.Debug("getViewStates: Found input self closing tag")
				isViewState := false
				isValidation := false
				isGenerator := false
			Loop:
				for {
					key, val, moreAttr := z.TagAttr()
					switch string(key) {
					case "type":
						if string(val) != "hidden" {
							break Loop
						}
					case "name":
						if string(val) == "__VIEWSTATEGENERATOR" {
							log.Debug("getViewStates: found __VIEWSTATEGENERATOR")
							isGenerator = true
						} else if string(val) == "__EVENTVALIDATION" {
							log.Debug("getViewStates: found __EVENTVALIDATION")
							isValidation = true
						} else if strings.Contains(string(val), "VIEWSTATE") &&
							string(val) != "__VIEWSTATEFIELDCOUNT" {
							log.Debug("getViewStates: found a VIEWSTATE")
							isViewState = true
						}
					case "value":
						if isViewState {
							log.Debugf("getViewStates: Adding to VIEWSTATES %d", len(viewStateInfo.ViewStates)+1)
							viewStateInfo.ViewStates = append(viewStateInfo.ViewStates, string(val))
						} else if isValidation {
							log.Debugf("getViewStates: setting EventValidation")
							viewStateInfo.EventValidation = string(val)
						} else if isGenerator {
							log.Debugf("getViewStates: Setting ViewStateGenerator")
							viewStateInfo.ViewStateGenerator = string(val)
						}
					}

					if !moreAttr {
						break
					}
				}
			}
			// Find input tags with type attr hiden, name attr containing VIEWSTATE
			// Find value of corresponding attr
		}
	}
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
	timeLayout := "  Monday, January 2, 2006, 03:04 PM MST"
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
