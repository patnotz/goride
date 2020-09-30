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

const baseUrl = "https://www.strava.com/api/v3"
const authUrl = "https://www.strava.com/oauth/authorize"
const clientId = 53956
const accessToken = "355aabb46aa2840403a73472e01f4421f946659f"

type AthleteData struct {
	FirstName string
	LastName  string
	Id        float64
	Username  string
}

type AuthContext struct {
	AccessToken  string  `json:"access_token"`
	RefreshToken string  `json:"refresh_token"`
	ExpiresAt    float64 `json:"expires_at"`
	ExpiresIn    float64 `json:"expires_in"`
	Athlete      AthleteData
}

type Activity struct {
	Name               string
	Distance           float64
	MovingTime         float64 `json:"moving_time"`
	TotalElevationGain float64 `json:"total_elevation_gain"`
	Type               string
	Id                 float64
	StartDate          string `json:"start_time"`
	StartDateLocal     string `json:"start_date_local"`
	Timezone           string
	UtcOffset          float64 `json:"utc_offset"`
	GearId             string  `json:"gear_id"`
	Kilojoules         float64
	SufferScore        float64 `json:"suffer_score"`
}

func printBytesAsStringMap(b []byte) {
	var m map[string]interface{}
	err := json.Unmarshal(b, &m)
	errHandler(err)
	fmt.Println(m)
}

func readClientSecret() (clientSecret string) {
	contents, err := ioutil.ReadFile("strava_client_secret.txt")
	errHandler(err)
	clientSecret = strings.TrimSpace(string(contents))
	return
}

func authHandler(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	url := authUrl + "?client_id=53956&response_type=code&scope=activity:read_all&redirect_uri=https://localhost:9000/welcome"
	fmt.Fprint(w, "<html><body>Click <a href=\""+url+"\"><b>here</b></a> to Authenticate.</body></html>")
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

	fmt.Println("Fetching auth tokens...")
	client := &http.Client{}
	resp, err := client.PostForm(baseUrl+"/oauth/token",
		url.Values{
			"client_id":     {fmt.Sprintf("%d", clientId)},
			"client_secret": {clientSecret},
			"code":          {authCode}})

	errHandler(err)
	bodyBytes, _ := ioutil.ReadAll(resp.Body)
	fmt.Println("response map:")
	printBytesAsStringMap(bodyBytes)
	var authContext AuthContext
	err = json.Unmarshal(bodyBytes, &authContext)
	errHandler(err)
	fmt.Println("authContext:")
	s, _ := json.MarshalIndent(authContext, "", "\t")
	fmt.Println(string(s))

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	link := "https://localhost:9000/auth"
	fmt.Fprint(w, "<html><body>Try again: <a href=\""+link+"\"/><b>"+link+"</b></a></body></html>")

	activities := getActivityData(authContext)
	fmt.Println(activities)
	s, _ = json.MarshalIndent(activities, "", "\t")
	fmt.Println(string(s))
	fmt.Println("End: welcomeHandler")
}

func errHandler(err error) {
	if err != nil {
		fmt.Println("Error :(")
		fmt.Println(err)
		panic(err)
	}
}

func makeRequest(url string, authContext AuthContext) (bodyBytes []byte) {
	fmt.Println("fetching: " + url)
	client := &http.Client{}
	req, err := http.NewRequest("GET", url, nil)
	errHandler(err)
	req.Header.Add("Authorization", ("Bearer " + authContext.AccessToken))
	resp, err := client.Do(req)
	errHandler(err)

	defer resp.Body.Close()
	bodyBytes, err = ioutil.ReadAll(resp.Body)
	errHandler(err)
	return
}

func getActivityData(authContext AuthContext) (activities []Activity) {
	// Get Activity data
	username := authContext.Athlete.Username
	athlete_id := fmt.Sprintf("%.0f", authContext.Athlete.Id)
	fmt.Println("Getting activity data for " + username + " (" + athlete_id + ")")
	url := baseUrl + "/athlete/activities"
	bodyBytes := makeRequest(url, authContext)
	err := json.Unmarshal(bodyBytes, &activities)
	errHandler(err)
	return
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
