package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
)

// declare the structure to be a blank object, so we can attach methods to it
type server struct{}

// default structure for recieving external http requests
// var DefaultClient = &http.Client{}

// get function
func get(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// response and error from get request for data
	res, err := http.Get("https://api.darksky.net/forecast/2cd42058712708466e62c7d34e7874f5/37.8267,-122.4233")

	// check if there is an error, handle it
	if err != nil {
		log.Fatal(err)
	}

	// get all the data from the response
	data, _ := ioutil.ReadAll(res.Body)

	// close the response body
	res.Body.Close()

	// show the data
	fmt.Printf("%s\n", data)

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(data))
}

// post function
func post(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	w.Write([]byte(`{"message": "POST"}`))
}

// put function
func put(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted)
	w.Write([]byte(`{"message": "PUT"}`))
}

// delete function
func delete(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"message": "DELETE"}`))
}

// 404 function
func notFound(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusNotFound)
	w.Write([]byte(`{"message": "404"}`))
}

// function for dealing with parameters
func params(w http.ResponseWriter, r *http.Request) {
	// set the parameters passed in as a variable here so it can be dealt with
	pathParams := mux.Vars(r)
	w.Header().Set("Content-Type", "application/json")

	// set id to negative one, check to make sure an id is given that is an integer
	userID := -1
	var err error
	if val, ok := pathParams["userID"]; ok {
		userID, err = strconv.Atoi(val)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(`{"message": "int required"}`))
			return
		}
	}

	commentID := -1
	if val, ok := pathParams["commentID"]; ok {
		commentID, err = strconv.Atoi(val)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(`{"message": "int required"}`))
			return
		}
	}

	query := r.URL.Query()
	location := query.Get("location")

	w.Write([]byte(fmt.Sprintf(`{"userID": %d, "commentID": %d, "location": "%s" }`, userID, commentID, location)))
}

func main() {
	// declares resonses as a variabale to be handeled
	r := mux.NewRouter()

	// set each metehod of response to be dealt with by the correct function
	r.HandleFunc("/", get).Methods("GET")
	r.HandleFunc("/", post).Methods(http.MethodPost)
	r.HandleFunc("/", put).Methods(http.MethodPut)
	r.HandleFunc("/", delete).Methods(http.MethodDelete)
	r.HandleFunc("/", notFound)

	// this sets the server to 8080
	log.Fatal(http.ListenAndServe(":8080", r))
}
