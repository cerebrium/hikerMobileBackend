package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"

	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
	_ "github.com/joho/godotenv/autoload"
)

// ---------------------------------------------------------- structs
// allows for syncronicity
var wg sync.WaitGroup

// declare the structure to be a blank object, so we can attach methods to it
type server struct{}

type locationString struct {
	City string
}

// Response struct to map entire response
type responceStructure struct {
	Query    []string      `json:"query"`
	Features []siteDetails `json:"features"`
}

type siteDetails struct {
	Center []float32 `json:"center"`
}

// ----------------------------------------------------------
// turn string into lat long
func getLocation(locationBody string) [2]float32 {
	// get the api ket from env
	mapBoxAPIKey := os.Getenv("MAPBOX_API_KEY")

	// construct the url
	reqString := fmt.Sprintf("https://api.mapbox.com/geocoding/v5/mapbox.places/%s.json", locationBody)
	req, err := http.NewRequest("GET", reqString, nil)
	q := req.URL.Query()

	// add the api key to the url as a parameter
	q.Add("access_token", mapBoxAPIKey)

	// add the params into the request string
	req.URL.RawQuery = q.Encode()

	// check if there is an error
	if err != nil {
		log.Fatal(err)
	}

	// make the request to the api to get coords back
	res, err := http.Get(req.URL.String())

	// check for error
	if err != nil {
		log.Fatal(err)
		os.Exit(1)
	}

	// create the response as data
	data, errReading := ioutil.ReadAll(res.Body)

	// check for errors
	if errReading != nil {
		log.Fatal(errReading)
	}

	// dont need to have eveything in the json described
	var dataMapped responceStructure
	errMarsh := json.Unmarshal([]byte(data), &dataMapped)

	// check for errors
	if errMarsh != nil {
		log.Fatal(errMarsh)
	}

	// arrays you have to predefime their size
	var latLong [2]float32

	// grab the first one... the one i want, and push values into the array
	for i := 0; i < 1; i++ {
		latLong[0] = dataMapped.Features[i].Center[0]
		latLong[1] = dataMapped.Features[i].Center[1]
	}

	// return the array
	return latLong
}

// ----------------------------------------------------------
// submit lat long into trail fetching function
// function for getting the trail data
func getTrails(latLong [2]float32) []byte {
	// load api keys
	apiKeyHiker := os.Getenv("HIKING_API_KEY")
	lat := fmt.Sprintf("%.6f", latLong[1])
	lon := fmt.Sprintf("%.6f", latLong[0])
	fmt.Println("lat: ", lat, "\nLon: ", lon)

	// make a request to the hiker url
	req, err := http.NewRequest("GET", "https://www.hikingproject.com/data/get-trails", nil)
	q := req.URL.Query()

	// add query
	q.Add("key", apiKeyHiker)
	q.Add("lat", lat)
	q.Add("lon", lon)
	q.Add("maxDistance", "200")

	req.URL.RawQuery = q.Encode()

	// handle any error that shows up
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(req.URL.String())
	// set the request to a get request
	res, err := http.Get(req.URL.String())

	// handle any error that shows up
	if err != nil {
		log.Fatal(err)
	}

	// create the response as data
	data, _ := ioutil.ReadAll(res.Body)

	// close the response
	res.Body.Close()

	var dataMapped map[string]interface{}
	errMarsh := json.Unmarshal([]byte(data), &dataMapped)
	if err != nil {
		log.Fatal(errMarsh)
	}

	// return data
	return data
}

// get weather function
func getWeather(w http.ResponseWriter, r *http.Request) {
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

// get the hiker data
func getHikerData(w http.ResponseWriter, r *http.Request) {
	// set the headers of the writting as application/json
	w.Header().Set("Content-Type", "application/json")

	// getting the body for use
	// ----------------------------------------------------------------------------

	// check if the size is too large
	r.Body = http.MaxBytesReader(w, r.Body, 1048576)

	// get the body from the request
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()

	// set the stype of the response we are looking for
	var loc locationString
	err := dec.Decode(&loc)
	if err != nil {
		var syntaxError *json.SyntaxError
		var unmarshalTypeError *json.UnmarshalTypeError

		switch {
		// syntax error case
		case errors.As(err, &syntaxError):
			msg := fmt.Sprintf("Request body contains badly formed JSON (at position %d)", syntaxError.Offset)
			http.Error(w, msg, http.StatusBadRequest)

			// case of decode returning an EOF because of bad json syntax
		case errors.Is(err, io.ErrUnexpectedEOF):
			msg := fmt.Sprintf("Request body contains badly-formed JSON")
			http.Error(w, msg, http.StatusBadRequest)

			// catch errors where types are being messed up
		case errors.As(err, &unmarshalTypeError):
			msg := fmt.Sprintf("Request body contains an invalid value for the %q field (at position %d)", unmarshalTypeError.Field, unmarshalTypeError.Offset)
			http.Error(w, msg, http.StatusBadRequest)

			// if there are extra unexpected fields in the body it throws an error
		case strings.HasPrefix(err.Error(), "json: unknown field "):
			fieldName := strings.TrimPrefix(err.Error(), "json: unknown field ")
			msg := fmt.Sprintf("Request body contains and unknown field %s", fieldName)
			http.Error(w, msg, http.StatusBadRequest)

			// if the body is empty it returns an EOF
		case errors.Is(err, io.EOF):
			msg := "Request body must not be empty"
			http.Error(w, msg, http.StatusBadRequest)

			// if the body is too long, handle that
		case err.Error() == "http: request body too large":
			msg := "Request body must not be larger than 1MB"
			http.Error(w, msg, http.StatusRequestEntityTooLarge)

			// default to sending the error and a 500
		default:
			log.Println(err.Error())
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		}
		return
	}

	// If the request body only contained a single JSON object this will return an io.EOF error. So if we get anything else,
	// we know that there is additional data in the request body.
	err = dec.Decode(&struct{}{})
	if err != io.EOF {
		msg := "Requset body must conatin a single JSON object"
		http.Error(w, msg, http.StatusBadRequest)
		return
	}
	// send the data back
	fmt.Println(getTrails(getLocation(loc.City)))

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(getTrails(getLocation(loc.City))))

}

// example post function
func post(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	w.Write([]byte(`{"message": "POST"}`))
}

// example put function
func put(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted)
	w.Write([]byte(`{"message": "PUT"}`))
}

// example delete function
func delete(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"message": "DELETE"}`))
}

// example 404 function
func notFound(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusNotFound)
	w.Write([]byte(`{"message": "404"}`))
}

// example function for dealing with parameters
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
	// allow for env variables
	err := godotenv.Load()

	// handle error of loading
	if err != nil {
		log.Fatal(err)
	}

	// declares resonses as a variabale to be handeled
	r := mux.NewRouter()

	// set each metehod of response to be dealt with by the correct function
	r.HandleFunc("/trails", getHikerData).Methods("POST")
	r.HandleFunc("/weather", getWeather).Methods("GET")
	r.HandleFunc("/", post).Methods(http.MethodPost)
	r.HandleFunc("/", put).Methods(http.MethodPut)
	r.HandleFunc("/", delete).Methods(http.MethodDelete)
	r.HandleFunc("/", notFound)

	// this sets the server to 8080
	log.Fatal(http.ListenAndServe(":8080", r))
}
