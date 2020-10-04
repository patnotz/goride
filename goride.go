// You can edit this code!
// Click here and start typing.
package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

const baseUrl = "https://www.strava.com/api/v3"
const authUrl = "https://www.strava.com/oauth/authorize"
const clientId = 53956
const activitiesPerPage = 30
const metersPerMile = 1609.344
const feetPerMeter = 3.2808399
const secondsPerHour = 3200.0

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

type GearData struct {
	Id          string
	Name        string
	BrandName   string `json:"brand_name"`
	ModelName   string `json:"model_name"`
	Description string
}

// All units are metric, per the native Strava API.
type ActivityData struct {
	Name               string
	Distance           float64
	MovingTime         float64 `json:"moving_time"`
	ElapsedTime        float64 `json:"elapsed_time"`
	TotalElevationGain float64 `json:"total_elevation_gain"`
	Type               string
	Id                 float64
	StartDate          string `json:"start_time"`
	StartDateLocal     string `json:"start_date_local"`
	Timezone           string
	UtcOffset          float64 `json:"utc_offset"`
	GearId             string  `json:"gear_id"`
	GearName           string
	Kilojoules         float64
	SufferScore        float64 `json:"suffer_score"`
	AverageWatts       float64 `json:"average_watts"`
	MaxWatts           float64 `json:"max_watts"`
	AverageHeartrate   float64 `json:"average_heartrate"`
	MaxHeartrate       float64 `json:"max_heartrate"`
}

type HistoryData struct {
	Activity             ActivityData
	CumulativeDistance   float64
	CumulativeElevation  float64
	CumulativeMovingTime float64
	CumulativeKilojoules float64
}

type UserData struct {
	Athlete AthleteData
	History []HistoryData
}

func MetersToMiles(meters float64) float64 {
	return meters / metersPerMile
}

func MetersToFeet(meters float64) float64 {
	return meters * feetPerMeter
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
	funcMap := template.FuncMap{
		"m_to_mi": MetersToMiles,
		"m_to_ft": MetersToFeet,
	}
	tmpl := template.Must(template.New("welcome.html").
		Funcs(funcMap).
		ParseFiles("welcome.html"))

	dat, err := url.ParseQuery(req.URL.RawQuery)
	errHandler(err)
	authCode := dat["code"][0]
	fmt.Println("Your authorization code is", authCode)
	clientSecret := readClientSecret()

	fmt.Println("Fetching auth tokens...")
	client := &http.Client{}
	resp, err := client.PostForm(baseUrl+"/oauth/token",
		url.Values{
			"client_id":     {fmt.Sprintf("%d", clientId)},
			"client_secret": {clientSecret},
			"code":          {authCode}})

	errHandler(err)
	bodyBytes, _ := ioutil.ReadAll(resp.Body)
	var authContext AuthContext
	err = json.Unmarshal(bodyBytes, &authContext)
	errHandler(err)
	fmt.Println("authContext:")
	s, _ := json.MarshalIndent(authContext, "", "\t")
	fmt.Println(string(s))

	var userData UserData
	userData.Athlete = authContext.Athlete
	activities := getActivityData(authContext)
	fmt.Printf("Found %d Rides.\n", len(activities))

	cumulativeDistance := 0.0
	cumulativeElevation := 0.0
	cumulativeMovingTime := 0.0
	cumulativeKilojoules := 0.0
	for i := len(activities) - 1; i >= 0; i-- {
		activity := activities[i]	
		cumulativeDistance += activity.Distance
		cumulativeElevation += activity.TotalElevationGain
		cumulativeMovingTime += activity.MovingTime
		cumulativeKilojoules += activity.Kilojoules
		var history HistoryData
		history.Activity = activity
		history.CumulativeDistance = cumulativeDistance
		history.CumulativeElevation = cumulativeElevation
		history.CumulativeMovingTime = cumulativeMovingTime
		history.CumulativeKilojoules = cumulativeKilojoules
		userData.History = append(userData.History, history)
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	err = tmpl.Execute(w, userData)
	errHandler(err)
}

func errHandler(err error) {
	if err != nil {
		fmt.Println("Error :(")
		fmt.Println(err)
		panic(err)
	}
}

func makeRequest(url string, params map[string]string, authContext AuthContext) (bodyBytes []byte) {
	client := &http.Client{}
	req, err := http.NewRequest("GET", url, nil)
	errHandler(err)
	req.Header.Add("Authorization", ("Bearer " + authContext.AccessToken))
	q := req.URL.Query()
	for key, value := range params {
		q.Add(key, value)
	}
	req.URL.RawQuery = q.Encode()
	fmt.Println("getting:", req.URL)
	resp, err := client.Do(req)
	errHandler(err)

	defer resp.Body.Close()
	bodyBytes, err = ioutil.ReadAll(resp.Body)
	errHandler(err)
	return
}

func getActivityData(authContext AuthContext) (activities []ActivityData) {
	url := baseUrl + "/athlete/activities"

	gearMap := make(map[string]GearData)
	params := make(map[string]string)
	params["per_page"] = strconv.Itoa(activitiesPerPage)
	// Pagination
	for page := 1; true; page++ {
		params["page"] = strconv.Itoa(page)
		bodyBytes := makeRequest(url, params, authContext)
		if len(bodyBytes) == 0 {
			break
		}
		var batch []ActivityData
		err := json.Unmarshal(bodyBytes, &batch)
		errHandler(err)
		fmt.Printf("page %d: fetched %d activities.\n", page, len(batch))
		for _, activity := range batch {
			if activity.Type != "Ride" {
				continue
			}
			gear, found := gearMap[activity.GearId]
			if !found {
				gear = getGearData(authContext, activity.GearId)
				gearMap[activity.GearId] = gear
			}
			activity.GearName = gear.Name
			activities = append(activities, activity)
		}
		if len(batch) < activitiesPerPage {
			break
		}
	}
	return
}

func getGearData(authContext AuthContext, gearId string) (gear GearData) {
	url := baseUrl + "/gear/" + gearId
	bodyBytes := makeRequest(url, make(map[string]string), authContext)
	err := json.Unmarshal(bodyBytes, &gear)
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
