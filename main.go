package main

import (
	"fmt"

	// "github.com/go-fed/activity/streams"
	"github.com/go-fed/activity/pub"
	"github.com/gorilla/mux"
	// "errors"
	"log"
	"net/http"

	// "net/url"

	"encoding/json"
	"io/ioutil"
	"os"

	"gopkg.in/ini.v1"

	// "html"
	// "context"

	// "github.com/davecgh/go-spew/spew"
)

var baseURL = "http://example.com"

func main() {

	var err error

	fmt.Println("=========================================================================")

	// read configuration file (config.ini)

	cfg, err := ini.Load("config.ini")
    if err != nil {
        fmt.Printf("Fail to read file: %v", err)
        os.Exit(1)
    }
	// Load base url from configuration file
	baseURL = cfg.Section("general").Key("baseURL").String()

    fmt.Println("Domain Name:", baseURL)

	var outboxHandler http.HandlerFunc = func(w http.ResponseWriter, r *http.Request) {
		username := mux.Vars(r)["actor"]
		// TODO replace this with a LoadActor that loads an actor from the database with this username
		actor, err := LoadActor(username)
		if err != nil {
			fmt.Println("Can't create local actor")
			return
		}
		if pub.IsActivityPubRequest(r){
			actor.HandleOutbox(w, r)
		} else {
			// The above does nothing if it's a non-ActivityPub request so 
			// handle non-ActivityPub request here, such as serving a webpage.
		}
	}
	var inboxHandler http.HandlerFunc = func(w http.ResponseWriter, r *http.Request) {
		// Populate c with request-specific information
		username := mux.Vars(r)["actor"]
		// TODO replace this with a LoadActor that loads an actor from the database with this username
		actor, err := LoadActor(username)
		if err != nil {
			fmt.Println("Can't create local actor")
			return
		}
		if pub.IsActivityPubRequest(r){
			actor.HandleInbox(w, r)
		} else {
			// The above does nothing if it's a non-ActivityPub request so 
			// handle non-ActivityPub request here, such as serving a webpage.
		}
	}

	var actorHandler http.HandlerFunc = func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("Remote server just fetched our /actor endpoint")

		username := mux.Vars(r)["actor"]
		// TODO replace this with a LoadActor that loads an actor from the database with this username
		// error out if this actor does not exist
		actor, err := LoadActor(username)
		if err != nil {
			fmt.Println("Can't create local actor")
			return
		}
		fmt.Fprintf(w, actor.whoAmI())
	}

	// Add the handlers to a HTTP server
	gorilla := mux.NewRouter()
	gorilla.HandleFunc("/{actor}/outbox", outboxHandler)
	gorilla.HandleFunc("/{actor}/inbox", inboxHandler)
	gorilla.HandleFunc("/{actor}/inbox/", inboxHandler)
	gorilla.HandleFunc("/{actor}", actorHandler)
	gorilla.HandleFunc("/{actor}/", actorHandler)
	http.Handle("/", gorilla)

	// get the list of users to relay
	jsonFile, err := os.Open("actors.json")

	if err != nil {
		fmt.Println("something is wrong with the json file containing the actors")
		fmt.Println(err)
	}

	// Unmarshall it into a map of string arrays
	whoFollowsWho := make(map[string][]string)
	byteValue, _ := ioutil.ReadAll(jsonFile)
	json.Unmarshal(byteValue, &whoFollowsWho)

	// fmt.Println(string(byteValue))
	// create all local actors if they don't exist yet
	for follower, followees := range whoFollowsWho {
		fmt.Println("Local Actor: "+ follower)
		followerActor, err := GetActor(follower, "emptySummary", "Service", baseURL+"/"+follower)
		if err != nil {
			fmt.Println("error creating local follower")
			return
		}
		// Now follow each one of it's users
		fmt.Println("Users to relay:")
		for _, followee := range followees {
			fmt.Println(followee)
			if err != nil{
				fmt.Println("Couldn't create local actor")
				return
			}
			followerActor.Follow(followee)
		}
	}

	http.HandleFunc("/hi", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Hi")
	})

	log.Fatal(http.ListenAndServe(":8081", nil))

}
