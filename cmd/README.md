calendar

- Goal of this app is to scrape the 8-rinks website for the schedule of a particular team and add it to
the users google calendar.

Done:

- So far it is able to get the team schedule page of the website and find if the user entered team
exists and also return the team's unique identifier.
- Did some initial parsing of the games, but it is all text based. This should be replaced with
some code that uses the html.Tockenizer

TODO:

- Use the team's unqiue identifer to search for games
- Get the games for a particular team
- Add the games for a team to a users calendar
- Run this app periodically

Enhancements:
- Get games in a specified time range
- Allow multiple teams to be specified
