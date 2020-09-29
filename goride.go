// You can edit this code!
// Click here and start typing.
package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
)

const base_url = "https://www.strava.com/api/v3/"
const access_token = "355aabb46aa2840403a73472e01f4421f946659f"

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
	req.Header.Add("Authorization", ("Bearer " + access_token))
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
	url := base_url + "athlete"
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
	url := base_url + "athlete/activities"
	dat := makeRequest(url)
	fmt.Println(dat)
}

func main() {

	// handle '/' route
	http.HandleFunc("/", func(res http.ResponseWriter, req *http.Request) {
		username, athlete_id := getAthleteData()
		fmt.Fprint(res, "Go Ride, " + username + " (" + athlete_id + ")!")
	})

	// run the server on port 9000
	log.Fatal(http.ListenAndServeTLS(":9000", "goride.crt", "goride.key", nil))
}
