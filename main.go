package main

import (
	"fmt"
	// "os"
	"io/ioutil"
	"strconv"
	// "html"
	"encoding/json"
	"log"
	"net/http"
)

type Lane struct {
	Headers map[string]string `json:"headers"`
	Content string `json:"content"`
}

type Config struct {
	Port int `json:"port"`
	Default string `json:"default"`
	Lanes map[string]Lane `json:"lanes"`
}

func main() {
	log.Println("Welcome to LaneChange™")

	// try to open config file
	jsonData, err := ioutil.ReadFile("config.json")

	if err != nil {
		fmt.Println("Error:", err)
		fmt.Println("Ensure that config.json exists in the current directory and is well-formatted")
		return
	}

	// load our preferences from the config file
	// and some error handling
	var config Config
	err = json.Unmarshal(jsonData, &config)

	if err != nil {
		fmt.Println("Error:", err)
		fmt.Println("An issue was encountered trying to parse config.json")
		return
	}

	if config.Port == 0 || config.Default == "" || len(config.Lanes) == 0 {
		fmt.Println("Error: Check that \"port\", \"default\", and \"lanes\" keys are all provided in config.json!")
		return
	}

	defaultLane, ok := config.Lanes[config.Default]
	if !ok {
		fmt.Printf("Error: Provided default key \"%s\" does not exist in lanes!\n", config.Default)
		return
	}
	// fmt.Printf("%+v\n", defaultLane)

	// our main endpoint, lookup our lane for the incoming IP
	// and return the expected response for that lane here
	http.HandleFunc("/", func(res http.ResponseWriter, req *http.Request) {

		// lookup current lane for our user
		lane := defaultLane // TODO

		// apply lane headers
		headers := res.Header()
		headers.Set("Cache-Control", "max-age=0, no-cache, no-store")
		for key, value := range lane.Headers {
			headers.Set(key, value)
		}

		// write lane content
		fmt.Fprintf(res, lane.Content)
	})

	log.Printf("Listening on localhost:%d\n", config.Port)

	log.Fatal(http.ListenAndServe(":" + strconv.Itoa(config.Port), nil))
}