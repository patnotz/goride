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
	"os/exec"
	"runtime"
	"strconv"
	"strings"
	"time"
)

const clientID = 53956
const activitiesPerPage = 200
const metersPerMile = 1609.344
const feetPerMeter = 3.2808399
const secondsPerHour = 3200.0
const baseURL = "https://www.strava.com/api/v3"
const authURL = "https://www.strava.com/oauth/authorize"
const componentsLog = "components.json"
const hostDomain = "penguin.linux.test"
const hostPort = "9000"
const authPage = "https://" + hostDomain + ":" + hostPort + "/auth"

// Time layouts, expressed as examples for: Mon Jan 2 15:04:05 MST 2006
const stLayout = "January 2, 2006"
const stravaLayout = "2006-01-02T15:04:05Z"

// AthleteData contains the basic identificatin data for the logged in Athlete
type AthleteData struct {
	FirstName string
	LastName  string
	ID        float64
	Username  string
}

// AuthContext holds the tokens and expiry info for the logged in Athlete
type AuthContext struct {
	AccessToken  string  `json:"access_token"`
	RefreshToken string  `json:"refresh_token"`
	ExpiresAt    float64 `json:"expires_at"`
	ExpiresIn    float64 `json:"expires_in"`
	Athlete      AthleteData
}

// GearData contains the data for a Strava Gear data object
type GearData struct {
	ID          string
	Name        string
	BrandName   string `json:"brand_name"`
	ModelName   string `json:"model_name"`
	Description string
}

// ActivityData contains the data for a Strava Activity
// All units are metric, per the native Strava API.
type ActivityData struct {
	Name               string
	Distance           float64
	MovingTime         float64 `json:"moving_time"`
	ElapsedTime        float64 `json:"elapsed_time"`
	TotalElevationGain float64 `json:"total_elevation_gain"`
	Type               string
	ID                 float64
	StartDate          string `json:"start_time"`
	StartDateLocal     string `json:"start_date_local"`
	Timezone           string
	UtcOffset          float64 `json:"utc_offset"`
	GearID             string  `json:"gear_id"`
	GearName           string
	Kilojoules         float64
	SufferScore        float64 `json:"suffer_score"`
	AverageWatts       float64 `json:"average_watts"`
	MaxWatts           float64 `json:"max_watts"`
	AverageHeartrate   float64 `json:"average_heartrate"`
	MaxHeartrate       float64 `json:"max_heartrate"`
}

// ComponentData holds detailed information about a bike component.
type ComponentData struct {
	Bike     string
	Type     string
	Brand    string
	Model    string
	Added    SimpleTime
	Removed  SimpleTime
	Distance float64
	Time     float64
	Notes    string
}

// HistoryData holds an ActivityData object and additional data about this activity history of activities up to this point.
type HistoryData struct {
	Activity             ActivityData
	CumulativeDistance   float64
	CumulativeElevation  float64
	CumulativeMovingTime float64
	CumulativeKilojoules float64
	GearDistance         map[string]float64
	GearTime             map[string]float64
}

// SimpleTime is a type for holding a Time.time value with simple formatting of JSON data.
type SimpleTime struct {
	time.Time
}

// UnmarshalJSON decodes a SimpleTime based on our prefered format.
func (st *SimpleTime) UnmarshalJSON(b []byte) (err error) {
	s := strings.Trim(string(b), "\"")
	if s == "null" {
		st.Time = time.Time{}
		return
	}
	st.Time, err = time.Parse(stLayout, s)
	return
}

// MarshalJSON encodes a SimpleTime based on our prefered format.
func (st *SimpleTime) MarshalJSON() ([]byte, error) {
	if st.Time.UnixNano() == nilTime {
		return []byte("null"), nil
	}
	return []byte(fmt.Sprintf("\"%s\"", st.Time.Format(stLayout))), nil
}

var nilTime = (time.Time{}).UnixNano()

// IsSet implements the IsSet method for a SimpleTime object.
func (st *SimpleTime) IsSet() bool {
	return st.UnixNano() != nilTime
}

// MetersToMiles converts a distance in meters to a distance in miles.
func MetersToMiles(meters float64) float64 {
	return meters / metersPerMile
}

// MetersToFeet converts a distance in meters to a distance in feet.
func MetersToFeet(meters float64) float64 {
	return meters * feetPerMeter
}

// SecondsToHours converts a time in seconds to a time in hours.
func SecondsToHours(seconds float64) float64 {
	return seconds / secondsPerHour
}

func printBytesAsStringMap(b []byte) {
	var m map[string]interface{}
	err := json.Unmarshal(b, &m)
	errHandler(err)
	fmt.Println(m)
}

func open(url string) error {
	var cmd string
	var args []string

	switch runtime.GOOS {
	case "windows":
		cmd = "cmd"
		args = []string{"/c", "start"}
	case "darwin":
		cmd = "open"
	default: // "linux", "freebsd", "openbsd", "netbsd"
		cmd = "xdg-open"
	}
	args = append(args, url)
	return exec.Command(cmd, args...).Start()
}

func readClientSecret() (clientSecret string) {
	contents, err := ioutil.ReadFile("strava_client_secret.txt")
	errHandler(err)
	clientSecret = strings.TrimSpace(string(contents))
	return
}

func readComponentsData() (components []ComponentData) {
	fmt.Println("Reading", componentsLog)
	contents, err := ioutil.ReadFile(componentsLog)
	errHandler(err)
	err = json.Unmarshal(contents, &components)
	errHandler(err)
	for _, component := range components {
		fmt.Println(component)
	}
	return
}

func authHandler(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	url := fmt.Sprintf("%s?client_id=%d&response_type=code&scope=activity:read_all&redirect_uri=https://%s:%s/welcome", authURL, clientID, hostDomain, hostPort)
	fmt.Fprint(w, "<html><body><a href=\""+url+"\"><img src=\"btn_strava_connectwith_orange.png\" alt=\"Connect with Stava\"/></a></body></html>")
}

func welcomeHandler(w http.ResponseWriter, req *http.Request) {
	funcMap := template.FuncMap{
		"m_to_mi": MetersToMiles,
		"m_to_ft": MetersToFeet,
		"s_to_h":  SecondsToHours,
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
	resp, err := client.PostForm(baseURL+"/oauth/token",
		url.Values{
			"client_id":     {fmt.Sprintf("%d", clientID)},
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

	activities := getActivityData(authContext)
	fmt.Printf("Found %d Rides.\n", len(activities))

	componentsData := readComponentsData()
	compTypes := make(map[string]bool)
	for _, c := range componentsData {
		compTypes[c.Type] = true
	}

	var History []HistoryData
	var cumulativeDistance, cumulativeElevation,
		cumulativeMovingTime, cumulativeKilojoules float64
	for i := len(activities) - 1; i >= 0; i-- {
		activity := activities[i]
		cumulativeDistance += activity.Distance
		cumulativeElevation += activity.TotalElevationGain
		cumulativeMovingTime += activity.MovingTime
		cumulativeKilojoules += activity.Kilojoules
		var hd HistoryData
		hd.Activity = activity
		hd.CumulativeDistance = cumulativeDistance
		hd.CumulativeElevation = cumulativeElevation
		hd.CumulativeMovingTime = cumulativeMovingTime
		hd.CumulativeKilojoules = cumulativeKilojoules
		startTime, err := time.Parse(stravaLayout, activity.StartDateLocal)
		errHandler(err)
		hd.GearDistance = make(map[string]float64)
		hd.GearTime = make(map[string]float64)
		for i := range componentsData {
			comp := &componentsData[i]
			startedAfterAdded := (*comp).Added.Time == time.Time{} || startTime.After((*comp).Added.Time)
			startedBeforeRemoved := (*comp).Removed.Time == time.Time{} || startTime.Before((*comp).Removed.Time)
			if startedAfterAdded && startedBeforeRemoved {
				compType := (*comp).Type
				(*comp).Distance += activity.Distance
				(*comp).Time += activity.MovingTime
				hd.GearDistance[compType] = (*comp).Distance
				hd.GearTime[compType] = (*comp).Time
			}
		}
		History = append(History, hd)
	}
	m := make(map[string]interface{})
	m["historyData"] = History
	m["athleteData"] = authContext.Athlete
	m["compTypes"] = compTypes
	m["authPage"] = authPage
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	err = tmpl.Execute(w, m)
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
	url := baseURL + "/athlete/activities"

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
			gear, found := gearMap[activity.GearID]
			if !found {
				gear = getGearData(authContext, activity.GearID)
				gearMap[activity.GearID] = gear
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

func getGearData(authContext AuthContext, gearID string) (gear GearData) {
	url := baseURL + "/gear/" + gearID
	bodyBytes := makeRequest(url, make(map[string]string), authContext)
	err := json.Unmarshal(bodyBytes, &gear)
	errHandler(err)
	return
}

func main() {
	http.HandleFunc("/auth", authHandler)
	http.HandleFunc("/welcome", welcomeHandler)
	http.Handle("/", http.FileServer(http.Dir("./static")))

	fmt.Println("Serving on", authPage)
	log.Fatal(http.ListenAndServeTLS(":"+hostPort, "goride.crt", "goride.key", nil))
}
