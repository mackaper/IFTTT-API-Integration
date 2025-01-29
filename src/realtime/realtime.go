package realtime

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"
	"pvk/API/db"

	"github.com/gobwas/ws"
	"github.com/gobwas/ws/wsutil"
	"github.com/google/uuid"
)


// Send webhook post request to IFTTT
func postIFTTT(hd *StateManager, series SocketSeries) {
	header := http.Header{}
	IFTTT_SERVICE_KEY := os.Getenv("IFTTT_SERVICE_KEY")
	if IFTTT_SERVICE_KEY == "" {
		log.Fatal("IFTTT_SERVICE_KEY is not set")
	}

	triggerIds := getTriggerIdentity(hd, series)
	if len(triggerIds) == 0 {
		return
	}

	id := uuid.New()
	header.Add("IFTTT-Service-Key", IFTTT_SERVICE_KEY)
	header.Add("Content-Type", "application/json")
	header.Add("X-Request-ID", id.String())

	var data IFTTTNotification
	for _, t := range triggerIds {
		data.Data = append(data.Data, IFTTTNotificationItem{TriggerIdentity: t})
	}

	jsonData, err := json.Marshal(data)
	if err != nil {
		log.Println(err)
	}
	body := bytes.NewReader(jsonData)

	url := "https://realtime.ifttt.com/v1/notifications"

	r, _ := http.NewRequest("POST", url, body)
	r.Header = header

	client := http.Client{}
	resp, err := client.Do(r)

	if err != nil {
		log.Println(err)
	}
	fmt.Println("IFTTT response: " + resp.Status)

	defer resp.Body.Close()
}

func parseTime(timr_str string) string {
	layout := "2006-01-02T15:04:05Z"
	t, _ := time.Parse(layout, timr_str)
	return t.Format("15:04")
}

func getValue1(game string, series string, tournament string) string {
	var seriesVal string
	var tournamentVal string
	if series != "not_found" {
		seriesVal = series
	} else {
		seriesVal = ""
	}
	if tournament != "not_found" {
		tournamentVal = tournament
	} else {
		tournamentVal = ""
	}
	return game + ": " + tournamentVal + " - " + seriesVal
}

func checkSeriesInState(hd *StateManager, series SocketSeries) bool {
	fmt.Println(hd.CurrentSeries)
	for _, stateSeries := range hd.CurrentSeries {
		if stateSeries == series.Payload.State.ID {
			return true
		}
	}
	return false
}

func shouldNotify(hd *StateManager, series SocketSeries) bool {
	inSeries := checkSeriesInState(hd, series)
	if inSeries {
		return false
	}

	if series.Payload.State.Lifecycle != "upcoming" {
		return false
	}

	layout := "2006-01-02T15:04:05Z"
	cutoff := time.Now().Add(time.Minute * 20)
	seriesStart, _ := time.Parse(layout, series.Payload.State.Start)
	if cutoff.Sub(seriesStart) < 0 {
		return false
	}

	hd.CurrentSeries = append(hd.CurrentSeries, series.Payload.State.ID)

	return true
}

func getTriggerIdentity(hd *StateManager, serires SocketSeries) []string {
  var teamIds []int
	for _, p := range serires.Payload.State.Participants {
		teamIds = append(teamIds, p.Roster.Team.ID)
	}

	var foundTriggers []string
	foundTriggers = append(foundTriggers, "e6e2c0007d65745c8e35f0f6d8b3229de841a7c9")
	for _, t := range teamIds {
		res := hd.Client.GetData(db.TeamPayload{
			Team: t,
		})
		if res != "" {
			foundTriggers = append(foundTriggers, res)
		}
	}
	return foundTriggers
}

func SocketListner(hd *StateManager) {
	// Connecting the websocket
	address := "wss://hermes.abiosgaming.com/subscribe?"
	//address := "ws://localhost:8000"
	header := http.Header{}

	header.Add("Abios-Secret", "0747fbc6230a456cb818262ce2855428")
	header.Add("Abios-Channel", "series_updates")
	header2 := ws.HandshakeHeaderHTTP(header)
	fmt.Println("Connecting to " + address)

	ctx := context.Background()

	conn, _, _, err := ws.Dialer{Header: header2}.Dial(ctx, address)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Connected to " + address)
	defer conn.Close()

	for {
		// Read message from the server
		msg, _, err := wsutil.ReadServerData(conn)
		if err != nil {
			log.Println(err)
			break
		}
		// parse msg to json
		var series SocketSeries
		json.Unmarshal(msg, &series)
		hd.Cache.PopulateRosters(&series.Payload.State)

		if shouldNotify(hd, series) {
			postIFTTT(hd, series)
		}
	}
}
