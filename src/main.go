package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"pvk/API/athena"
	"pvk/API/realtime"

	"pvk/API/db"
	"strings"
	"time"
)

type StateWrapper struct {
	*realtime.StateManager
}

func validateIFTTTServiceKey(r *http.Request) bool {
	IFTTT_SERVICE_KEY := os.Getenv("IFTTT_SERVICE_KEY")
	if IFTTT_SERVICE_KEY == "" {
		log.Fatal("IFTTT_SERVICE_KEY is not set")
	}
	header_value := r.Header.Get("IFTTT-Service-Key")
	return header_value == IFTTT_SERVICE_KEY
}

func index(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Hello World"))
}

func setup(w http.ResponseWriter, r *http.Request) {
	if !validateIFTTTServiceKey(r) {
		w.WriteHeader(http.StatusUnauthorized)
		log.Print("Invalid IFTTT-Service-Key")
		w.Write([]byte("Unauthorized"))
		return
	}
	response := struct {
		Data struct{} `json:"data"`
	}{}
	json_body, err := json.Marshal(response)
	if err != nil {
		log.Fatal("Invalid data")
	}
	w.Write(json_body)
}

func status(w http.ResponseWriter, r *http.Request) {
	IFTTT_SERVICE_KEY := os.Getenv("IFTTT_SERVICE_KEY")
	if IFTTT_SERVICE_KEY == "" {
		w.WriteHeader(http.StatusInternalServerError)
		log.Fatal("IFTTT_SERVICE_KEY is not set")
	}
	header_value := r.Header.Get("IFTTT-Service-Key")
	if header_value != IFTTT_SERVICE_KEY {
		w.WriteHeader(http.StatusUnauthorized)
		log.Print("IFTTT-Service-Key is not valid")
		w.Write([]byte("Unauthorized"))
		return
	}
	w.Write([]byte("OK"))
}

func trigger_fields(w http.ResponseWriter, r *http.Request) {
	log.Print("trigger fields polled")
	if !validateIFTTTServiceKey(r) {
		w.WriteHeader(http.StatusUnauthorized)
		response := struct { //create a response of erros
			Errors []IFTTTErrorMessage `json:"errors"`
		}{
			Errors: []IFTTTErrorMessage{
				{Message: "Invalid service key"},
			},
		}
		jsonResp, err := json.Marshal(response) //Marshal this into a JSON-object
		if err != nil {
			log.Print(err) //Print error if Marshaling went wrong
		}
		log.Print("Invalid IFTTT-Service-Key")
		w.Write(jsonResp)
		return
	}
	var request IFTTTRequestValue
	reqErr := json.NewDecoder(r.Body).Decode(&request)
	//If the request is invalid
	if reqErr != nil {
		log.Print("Invalid Request from IFTTT")
	}
	value := request.DynamicValue
	var response IFTTTRespValue

	//Making a direct request to Atlas, instead of CACHE
	requ := "https://atlas.abiosgaming.com/v3/teams?filter=name=" + value
	header := http.Header{}
	header.Add("Abios-Secret", "0747fbc6230a456cb818262ce2855428")

	//Create a request object
	req, err := http.NewRequest("GET", requ, nil)

	//if creation of request-object fails
	if err != nil {
		log.Fatalln(err)
	}
	req.Header = header

	//Send the request
	resp, err := http.DefaultClient.Do(req)

	// if request fails
	if err != nil {
		log.Fatal(err)
		response.DataValid = false
		response.DataMessage = "Sorry bro, the team does not exist xD"
	} else {
		response.DataValid = true
	}
}

func check_time(time_str string) bool {
	layout := "2006-01-02T15:04:05Z"
	t, _ := time.Parse(layout, time_str)
	return t.Sub(time.Now()).Minutes() < 10
}

func formatCompetitors(series athena.Series) string {
	result := ""
	if len(series.Participants) > 1 {
		result = fmt.Sprintf("%s vs %s", series.Participants[0].Roster.Team.Name, series.Participants[1].Roster.Team.Name)
		return result
	}
	if len(series.Participants) == 1 {
		result = series.Participants[0].Roster.Team.Name
	}
	return result
}

func (hd *StateWrapper) basic_trigger(w http.ResponseWriter, r *http.Request) {
	log.Print("basic_trigger polled")
	if !validateIFTTTServiceKey(r) {
		//WriteHeader sends an HTTP response header with the provided status code
		//Which in this case is "Unauthorized"
		w.WriteHeader(http.StatusUnauthorized)
		response := struct { //create a response of erros
			Errors []IFTTTErrorMessage `json:"errors"`
		}{
			Errors: []IFTTTErrorMessage{
				{Message: "Invalid service key"},
			},
		}
		jsonResp, err := json.Marshal(response) //Marshal this into a JSON-object
		if err != nil {
			log.Print(err) //Print error if Marshaling went wrong
		}
		log.Print("Invalid IFTTT-Service-Key")
		w.Write(jsonResp)
		return
	}

	var request IFTTTTriggerReqBody
	reqErr := json.NewDecoder(r.Body).Decode(&request)
	//If the request is invalid
	if reqErr != nil {
		log.Print("Invalid Request from IFTTT")
	}

	var limit int //limit = how many items to send back to IFTTT
	if request.Limit == nil {
		limit = 5
	} else {
		limit = *request.Limit //default limit = 50
	}

	series := athena.GetSeries(limit)
	hd.Cache.PopulateSeries(series)

	ids := make([]int, 0)
	for _, serie := range series {
		ids = append(ids, serie.ID)
	}
	hd.CurrentSeries = ids
	//This payload is what WE send back to IFTTT
	var responseBody IFTTTPayload
	responseBody.Data = make([]IFTTTDataType, 0)
	for _, serie := range series {
		dataObject := IFTTTDataType{
			GameName:       serie.Game.Title,
			TournamentName: serie.Tournament.Title,
			Competitors:    formatCompetitors(serie),
			CreatedAt:      serie.Start,
			Meta: IFTTTMetaData{
				ID:        serie.ID,
				Timestamp: time.Now().Unix(),
			},
		}
		responseBody.Data = append(responseBody.Data, dataObject)
	}
	jsonBody, err := json.Marshal(responseBody)
	if err != nil {
		log.Fatal(err)
	}

	w.Write(jsonBody)
}

func (hd *StateWrapper) deleteEndpoint(w http.ResponseWriter, r *http.Request) {
	log.Print("Delete_endpoint polled")
	if !validateIFTTTServiceKey(r) {
		w.WriteHeader(http.StatusUnauthorized)
		response := struct {
			Errors []IFTTTErrorMessage `json:"errors"`
		}{
			Errors: []IFTTTErrorMessage{
				{Message: "Invalid service key"},
			},
		}
		jsonResp, err := json.Marshal(response)
		if err != nil {
			log.Print(err)
		}
		log.Print("Invalid IFTTT-Service-Key")
		w.Write(jsonResp)
		return
	}
	trigger_id := strings.TrimPrefix(r.URL.Path, "/ifttt/v1/triggers/new_thing_created/trigger_identity/")
	hd.Client.DeleteData(trigger_id)
	w.Write([]byte("OK"))
}

func main() {
	db := db.MakeDBClient("localhost:6379", "", 0)
	state := realtime.StateManager{
		CurrentSeries: make([]int, 0),
		Cache:         athena.Athena{DBClient: db},
		Client:        db,
	}

	wrapper := StateWrapper{
		&state,
	}

	fmt.Println("Ready to go!")

	go realtime.SocketListner(&state)
	http.HandleFunc("/", index)
	http.HandleFunc("/ifttt/v1/status", status)
	http.HandleFunc("/ifttt/v1/test/setup", setup)
	http.HandleFunc("/ifttt/v1/triggers/new_thing_created", wrapper.basic_trigger)
	http.HandleFunc("/ifttt/v1/triggers/new_thing_created/fields/team/validate", trigger_fields)
	http.HandleFunc("/ifttt/v1/triggers/new_thing_created/trigger_identity/", wrapper.deleteEndpoint)
	http.ListenAndServe(":8080", nil)
}
