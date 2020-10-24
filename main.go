package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"strconv"
	"time"

	"github.com/patrickmn/go-cache"
)

type Lane struct {
	Headers map[string]string `json:"headers"`
	Content string            `json:"content"`
	LaneKey string            `json:"-"`
}

type Config struct {
	Default string          `json:"default"`
	Lanes   map[string]Lane `json:"lanes"`
}

type ConfigWithPort struct {
	*Config
	Port int `json:"port"`
}

type LaneChange struct {
	LaneKey  string        `json:"lane"`
	Duration time.Duration `json:"duration"`
}

type LaneChangeResp struct {
	LaneKey string    `json:"lane"`
	Expires time.Time `json:"expires"`
	IP      string    `json:"ip"`
}

// https://golangcode.com/get-the-request-ip-addr/
// https://stackoverflow.com/a/33301173
func GetIP(r *http.Request) string {
	forwarded := r.Header.Get("X-FORWARDED-FOR")
	if forwarded != "" {
		return forwarded
	}
	ip, _, _ := net.SplitHostPort(r.RemoteAddr)
	return ip
}

func main() {
	log.Println("Welcome to LaneChange, see the readme for more information")

	// try to open config file
	jsonData, err := ioutil.ReadFile("config.json")

	if err != nil {
		fmt.Println("Error:", err)
		fmt.Println("Ensure that config.json exists in the current directory and is well-formatted")
		return
	}

	// load our preferences from the config file
	// and some error handling
	var config ConfigWithPort
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

	for laneKey, lane := range config.Lanes {
		// hookup the key inside the lane for reference later
		lane.LaneKey = laneKey
		config.Lanes[laneKey] = lane
	}

	// setup the cache (drop expired every 10 min)
	// TODO: load from disk in case we crashed
	users := cache.New(cache.NoExpiration, 10*time.Minute)

	// our main endpoint, lookup our lane for the incoming IP
	// and return the expected response for that lane here
	http.HandleFunc("/", func(res http.ResponseWriter, req *http.Request) {
		ip := GetIP(req)

		// lookup current lane for our user
		var lane *Lane
		lane = &defaultLane
		userLane, found := users.Get(ip)
		if found {
			lane = userLane.(*Lane)
		}

		// apply lane headers
		headers := res.Header()
		headers.Set("Cache-Control", "max-age=0, no-cache, no-store")
		for key, value := range lane.Headers {
			headers.Set(key, value)
		}

		// write lane content
		fmt.Fprintf(res, lane.Content)
	})

	http.HandleFunc("/change", func(res http.ResponseWriter, req *http.Request) {
		ip := GetIP(req)

		if req.Method == http.MethodDelete {
			// remove entry for IP (does nothing if doesn't exist)
			users.Delete(ip)
			return
		}

		// process incoming lane change preference
		if req.Method == http.MethodPost {
			var change LaneChange

			err := json.NewDecoder(req.Body).Decode(&change)
			if err != nil {
				http.Error(res, err.Error(), http.StatusBadRequest)
				return
			}

			userLane, ok := config.Lanes[change.LaneKey]
			if ok {
				// set the lane pointer for this IP, expire with their duration
				// if duration is omitted, will be 0, which will go to the default expiry
				users.Set(ip, &userLane, change.Duration*time.Second)
				return
			}

			http.Error(res, "Error with request (Lane key invalid?)", http.StatusBadRequest)
			return
		}

		// try to lookup user (GET)
		userLane, expiration, found := users.GetWithExpiration(ip)
		if found {
			var change LaneChangeResp
			change.LaneKey = userLane.(*Lane).LaneKey
			change.Expires = expiration
			change.IP = ip
			output, _ := json.MarshalIndent(change, "", "\t")
			res.Write(output)
			return
		}
		res.WriteHeader(http.StatusNotFound)
	})

	http.HandleFunc("/config", func(res http.ResponseWriter, req *http.Request) {
		// respond with our config, but don't expose port
		var configNoPort Config
		configNoPort.Lanes = config.Lanes
		configNoPort.Default = config.Default
		output, _ := json.MarshalIndent(configNoPort, "", "\t")
		res.Write(output)
	})

	log.Printf("Listening on localhost:%d\n", config.Port)

	log.Fatal(http.ListenAndServe(":"+strconv.Itoa(config.Port), nil))
}
