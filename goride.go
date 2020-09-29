// You can edit this code!
// Click here and start typing.
package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strings"
)

const baseUrl = "https://www.strava.com/api/v3/"
const authUrl = "https://www.strava.com/oauth/authorize"
const clientId = "53956"
const accessToken = "355aabb46aa2840403a73472e01f4421f946659f"

func readClientSecret() (clientSecret string) {
	contents, err := ioutil.ReadFile("strava_client_secret.txt")
	errHandler(err)
	clientSecret = strings.TrimSpace(string(contents))
	return
}

func authHandler(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8") 
	url := authUrl + "?client_id=53956&response_type=code&scope=activity:read_all&redirect_uri=https://localhost:9000/welcome"
	fmt.Fprint(w, "<html><body>Click <a href=\"" + url + "\"><b>here</b></a> to Authenticate.</body></html>")
}

func welcomeHandler(w http.ResponseWriter, req *http.Request) {
	fmt.Println("Begin: welcomeHandler")
	fmt.Println("RawQuery = " + req.URL.RawQuery)
	dat, err := url.ParseQuery(req.URL.RawQuery)
	errHandler(err)
	fmt.Println("dat:")
	fmt.Println(dat)
	authCode := dat["code"][0]
	fmt.Println("Your authorization code is ", authCode)

	clientSecret := readClientSecret()
	fmt.Println("clientSecret: " + clientSecret)

	fmt.Println("End: welcomeHandler")
}

func errHandler(err error) {
	if err != nil {
		fmt.Println("Error :(")
		fmt.Println(err)
		panic(err)
	}
}

func makeRequest(url string) (dat map[string]interface{}) {
	fmt.Println("fetching: " + url)
	client := &http.Client{}
	req, err := http.NewRequest("GET", url, nil)
	errHandler(err)
	req.Header.Add("Authorization", ("Bearer " + accessToken))
	req.Header.Add("activity", "read")
	resp, err := client.Do(req)
	errHandler(err)

	defer resp.Body.Close()
	body_bytes, err := ioutil.ReadAll(resp.Body)
	errHandler(err)

	err = json.Unmarshal(body_bytes, &dat)
	errHandler(err)
	return
}

func getAthleteData() (username string, athlete_id string) {
	// Get Athlete data
	fmt.Println("Getting athlete data...")
	url := baseUrl + "athlete"
	dat := makeRequest(url)
	fmt.Println(dat)

	// Unpack Athlete data
	athlete_id = fmt.Sprintf("%.0f", dat["id"].(float64))
	username = dat["username"].(string)
	fmt.Println("Username  : " + username)
	fmt.Println("Athlete ID: " + athlete_id)
	return
}

func getActivityData() {
	username, athlete_id := getAthleteData()

	// Get Activity data
	fmt.Println("Getting activity data for " + username + " (" + athlete_id + ")")
	url := baseUrl + "athlete/activities"
	dat := makeRequest(url)
	fmt.Println(dat)
}

func main() {

	// handle '/' route
	http.HandleFunc("/", func(res http.ResponseWriter, req *http.Request) {
		// username, athlete_id := getAthleteData()
		// fmt.Fprint(res, "Go Ride, "+username+" ("+athlete_id+")!")
	})
	http.HandleFunc("/auth", authHandler)
	http.HandleFunc("/welcome", welcomeHandler)

	// run the server on port 9000
	log.Fatal(http.ListenAndServeTLS(":9000", "goride.crt", "goride.key", nil))
}
