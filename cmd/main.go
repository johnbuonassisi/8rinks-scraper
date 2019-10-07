package main

import (
	"fmt"
	"log"
	"net/http"
	"strings"

	"golang.org/x/net/html"
)

func main() {

	teamName := "Megpies FC"

	req, err := http.NewRequest("GET", "https://canlanaisl.icesports.com/BURNABY8RINKS/soccer-schedule.aspx", nil)
	if err != nil {
		log.Fatal(err)
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
		log.Fatal(err)
	}

	defer resp.Body.Close()
	z := html.NewTokenizer(resp.Body)
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
					fmt.Println("FOUND TEAMS!")
					// Find the team we are looking for in the next tokens
					var teamID []byte
					for {
						tt = z.Next()
						name, _ = z.TagName()

						// Finish if all teams have been iterated through
						if tt == html.EndTagToken && string(name) == "select" {
							fmt.Println("FINISHED TEAMS!")
							return
						}
						// Save the teamId specified in the tag
						if tt == html.StartTagToken && string(name) == "option" {
							_, teamID, _ = z.TagAttr()

						}
						// Check if the text token matches your team
						if tt == html.TextToken {
							text := z.Text()
							if string(text) == teamName {
								fmt.Printf("Found your team: %s \n", string(text))
								if len(teamID) == 0 {
									fmt.Printf("Could not find Team Id")
								}
								fmt.Printf("Team Id: %s \n", string(teamID))
								return
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

}

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
