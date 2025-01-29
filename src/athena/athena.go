package athena

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"pvk/API/db"
	"reflect"
	"time"
)

func CacheSet[T Cacheable](dbClient *db.DBClient, key int, data T, lifetime time.Duration) {
	prefix := ""
	switch reflect.TypeOf(data).String() {
	case "athena.Game":
		prefix = "game"
	case "athena.Roster":
		prefix = "roster"
	case "athena.Tournament":
		prefix = "tournament"
	}
	v, err := json.Marshal(data)
	prefixed_key := fmt.Sprintf("%s:%v", prefix, key)
	err = dbClient.Client.Set(db.Ctx, prefixed_key, v, lifetime).Err()
	if err != nil {
		panic(err)
	}
}

func CacheGet[T Cacheable](dbClient *db.DBClient, key int) T {
	val, err := dbClient.Client.Get(db.Ctx, fmt.Sprintf("%v", key)).Result()
	if err != nil {
		panic(err)
	}
	valByteArr := []byte(val)
	var data T
	err = json.Unmarshal(valByteArr, &data)
	if err != nil {
		panic(err)
	}
	fmt.Println(data)
	return data
}

func GetSeries(limit int) []Series {
	layout := "2006-01-02T15:04:05Z"
	now := time.Now().Add(time.Minute * 20)
	now_str := now.Format(layout)
	url := "https://atlas.abiosgaming.com/v3/series?order=start-desc&filter=start!=null,start<=" + now_str + "&take=" + fmt.Sprintf("%v", limit)
	header := http.Header{}
	header.Add("Abios-Secret", "<notPublishedInGit")

	//Create a request object
	req, err := http.NewRequest("GET", url, nil)

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
	}
	defer resp.Body.Close()
	var series []Series
	err = json.NewDecoder(resp.Body).Decode(&series)
	if err != nil {
		log.Fatal(err)
	}

	return series
}

func fetchMultiple(endpoint string, ids []int) *http.Response {
	filter := ""
	for i, id := range ids {
		if i == 0 {
			filter += fmt.Sprintf("%v", id)
		} else {
			filter += "," + fmt.Sprintf("%v", id)
		}
	}
	url := "https://atlas.abiosgaming.com/v3/" + endpoint + "?filter=id<={" + filter + "}"
	header := http.Header{}
	header.Add("Abios-Secret", "0747fbc6230a456cb818262ce2855428")

	//Create a request object
	req, err := http.NewRequest("GET", url, nil)

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
	}

	return resp
}

func (cache *Athena) PopulateRosters(serie *Series) {
	unCachedRosters := make([]int, 0)
	for i, p := range serie.Participants {
		cached := cache.CachedRosters.Get(p.Roster.ID)
		if cached.value.ID == 0 {
			unCachedRosters = append(unCachedRosters, p.Roster.ID)
		} else {
			serie.Participants[i].Roster = cached.value
		}
	}

	if len(unCachedRosters) > 0 {
		var rosters []Roster
		log.Print("Fetching rosters")
		resp := fetchMultiple("rosters", unCachedRosters)
		defer resp.Body.Close()
		err := json.NewDecoder(resp.Body).Decode(&rosters)
		if err != nil {
			log.Fatal(err)
		}
		for _, r := range rosters {
			cache.CachedRosters.Set(r.ID, r)
		}
	}
}

func (cache *Athena) PopulateSeries(series []Series) {
	unCachedTournaments := make([]int, 0)
	unCachedGames := make([]int, 0)
	unCachedRosters := make([]int, 0)

	for i, s := range series {
		if s.Tournament.ID != 0 {
			cached := CacheGet[Tournament](cache.DBClient, s.Tournament.ID)
			if cached.ID == 0 {
				unCachedTournaments = append(unCachedTournaments, s.Tournament.ID)
			} else {
				series[i].Tournament = cached
			}
		}
		if s.Game.ID != 0 {
			cached := CacheGet[Game](cache.DBClient, s.Game.ID)
			if cached.ID == 0 {
				unCachedGames = append(unCachedGames, s.Game.ID)
			} else {
				series[i].Game = cached
			}
		}
		for j, p := range s.Participants {
			cached := CacheGet[Roster](cache.DBClient, p.Roster.ID)
			if cached.ID == 0 {
				unCachedRosters = append(unCachedRosters, p.Roster.ID)
			} else {
				series[i].Participants[j].Roster = cached
			}
		}
	}

	cache.populateUncached(series, unCachedTournaments, unCachedGames, unCachedRosters)
}

func (cache *Athena) populateUncached(series []Series, tIDs []int, gIDs []int, rIDs []int) {
	var tournaments []Tournament
	var games []Game
	var rosters []Roster

	if len(tIDs) > 0 {
		log.Print("Fetching tournaments")
		resp := fetchMultiple("tournaments", tIDs)
		defer resp.Body.Close()
		err := json.NewDecoder(resp.Body).Decode(&tournaments)
		if err != nil {
			log.Fatal(err)
		}
		for _, t := range tournaments {
			CacheSet(cache.DBClient, t.ID, t, time.Hour*48)
		}
	}

	if len(gIDs) > 0 {
		log.Print("Fetching games")
		resp := fetchMultiple("games", gIDs)
		defer resp.Body.Close()
		err := json.NewDecoder(resp.Body).Decode(&games)
		if err != nil {
			log.Fatal(err)
		}
		for _, g := range games {
			CacheSet(cache.DBClient, g.ID, g, time.Hour*48)
		}
	}

	if len(rIDs) > 0 {
		log.Print("Fetching rosters")
		resp := fetchMultiple("rosters", rIDs)
		defer resp.Body.Close()
		err := json.NewDecoder(resp.Body).Decode(&rosters)
		if err != nil {
			log.Fatal(err)
		}

		teamsToFetch := make([]int, 0)
		for _, r := range rosters {
			teamsToFetch = append(teamsToFetch, r.Team.ID)
		}

		teamsResp := fetchMultiple("teams", teamsToFetch)
		var teamData []Team
		err = json.NewDecoder(teamsResp.Body).Decode(&teamData)
		if err != nil {
			log.Fatal(err)
		}
		teamsResp.Body.Close()

		for i, r := range rosters {
			for _, t := range teamData {
				if r.Team.ID == t.ID {
					rosters[i].Team = t
				}
			}
			CacheSet(cache.DBClient, r.ID, rosters[i], time.Hour*48)
		}

		for i, s := range series {
			for j, p := range s.Participants {
				for _, r := range rosters {
					if p.Roster.ID == r.ID {
						series[i].Participants[j].Roster = r
					}
				}
			}

			for _, g := range games {
				if s.Game.ID == g.ID {
					series[i].Game = g
				}
			}

			for _, t := range tournaments {
				if s.Tournament.ID == t.ID {
					series[i].Tournament = t
				}
			}
		}
	}
}
