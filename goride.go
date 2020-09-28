// You can edit this code!
// Click here and start typing.
package main

import "encoding/json"
import "fmt"
import "io/ioutil"
import "net/http"

func err_handler(err error) {
	if err != nil {
		fmt.Println("Error :(")
		fmt.Println(err)
		panic(err)
	}
}

func make_request(url string) (dat map[string]interface{}) {
	access_token := "7a69d16769e7791d56aabff25268f8800ce923d7"
	fmt.Println("fetching: " + url)
	client := &http.Client{}
	req, err := http.NewRequest("GET", url, nil)
	err_handler(err)
	req.Header.Add("Authorization", ("Bearer " + access_token))
	req.Header.Add("activity", "read")
	resp, err := client.Do(req)
	err_handler(err)

	defer resp.Body.Close()
	body_bytes, err := ioutil.ReadAll(resp.Body)
	err_handler(err)

	err = json.Unmarshal(body_bytes, &dat)
	err_handler(err)
	return
}

func main() {
	
	base_url := "https://www.strava.com/api/v3/"

	// Get Athlete data
	fmt.Println("Getting athlete data...")
	url := base_url + "athlete"
	dat := make_request(url)
	fmt.Println(dat)

	// Unpack Athlete data
	athlete_id := fmt.Sprintf("%.0f", dat["id"].(float64))
	username := dat["username"].(string)
	fmt.Println("Username  : " + username)
	fmt.Println("Athlete ID: " + athlete_id)

	// Get Activity data
	fmt.Println("Getting activity data for " + username + " (" + athlete_id + ")")
	url = base_url + "athlete/activities"
	dat = make_request(url)
	fmt.Println(dat)
	
}
