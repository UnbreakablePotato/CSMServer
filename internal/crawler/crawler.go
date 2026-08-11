package crawler

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	api "github.com/UnbreakablePotato/CSMServer/internal/API"
	"github.com/joho/godotenv"
)

type Leaderboard struct {
	Tier    string `json:"tier"`
	Queue   string `json:"queue"`
	Entries []struct {
		Puuid        string `json:"puuid"`
		LeaguePoints int    `json:"leaguePoints"`
		Rank         string `json:"rank"`
		Wins         int    `json:"wins"`
		Losses       int    `json:"losses"`
		Veteran      bool   `json:"veteran"`
		Inactive     bool   `json:"inactive"`
		FreshBlood   bool   `json:"freshBlood"`
		HotStreak    bool   `json:"hotStreak"`
	} `json:"entries"`
}

var _ = godotenv.Load()

var apiKey, _ = os.LookupEnv("leagueAPI")

var challengers Leaderboard

type Queue struct {
	VisitedMatches map[string]bool
	PendingMatches chan string
	PendingPuuids  chan string
	mu             *sync.RWMutex
}

var queue Queue

/*
Gets every challenger players puuid in specified region
*/
func InitialRequest() error {
	fullUrl := "https://euw1.api.riotgames.com/lol/league/v4/challengerleagues/by-queue/RANKED_SOLO_5x5?api_key=" + apiKey

	req, err := http.NewRequest("GET", fullUrl, nil)
	if err != nil {
		fmt.Printf("Error: %s\n", err)
		return err
	}

	client := http.Client{}

	res, err := client.Do(req)
	if err != nil {
		fmt.Printf("Error: %s\n", err)
		return err
	}

	data, err := io.ReadAll(res.Body)
	if err != nil {
		fmt.Printf("Error: %s\n", err)
	}

	if err := json.Unmarshal(data, &challengers); err != nil {
		fmt.Printf("Error: %s\n", err)
		return err
	}

	//queue.mu.Lock()
	for i := range challengers.Entries {
		queue.PendingPuuids <- challengers.Entries[i].Puuid
	}
	//queue.mu.Unlock()

	return nil
}

func ExtractMatchIds(q *Queue) {
	for {
		puuid := <-queue.PendingPuuids

		fullUrl := "https://europe.api.riotgames.com/lol/match/v5/matches/by-puuid/" + puuid + "/ids?start=0&count=5?api_key=" + apiKey

		req, err := http.NewRequest("GET", fullUrl, nil)
		if err != nil {
			fmt.Printf("Error: %s\n", err)

		}

		client := http.Client{}

		var res *http.Response

		var httperr error
		for {
			res, httperr = client.Do(req)
			if httperr != nil {
				//An error here with code 429 means the rate limit has been exceeded
				//We should retry in 2 min
				if res.StatusCode == 429 {
					time.Sleep(120000)
				}
				fmt.Printf("Error: %s\n", httperr)

			} else {
				break
			}
		}

		data, err := io.ReadAll(res.Body)
		if err != nil {
			fmt.Printf("Error: %s\n", err)

		}

		interResult := string(data)

		result := strings.Split(interResult, ",")

		for i := range result {
			result[i] = strings.ReplaceAll(result[i], "\"", "")
			result[i] = strings.ReplaceAll(result[i], "[", "")
			result[i] = strings.ReplaceAll(result[i], "]", "")
			//fmt.Printf("debug: %s\n", result[i])

			/*
				If a match has already been added to the queue do not add it again...
			*/
			q.mu.Lock()
			if !q.VisitedMatches[result[i]] {
				q.PendingMatches <- result[i]
			} else {
				q.VisitedMatches[result[i]] = true
			}
			q.mu.Unlock()
		}
	}
}

var intermediateGame api.Game

func ExtractMatchData() {

	for {
		matchID := <-queue.PendingMatches
		fullUrl := "https://europe.api.riotgames.com/lol/match/v5/matches/" + matchID + "?api_key=" + apiKey
		req, err := http.NewRequest("GET", fullUrl, nil)
		if err != nil {
			fmt.Printf("Error: %s\n", err)
			//return err
		}

		client := http.Client{}

		var res *http.Response

		var httperr error
		for {
			res, httperr = client.Do(req)
			if httperr != nil {
				//An error here with code 429 means the rate limit has been exceeded
				//We should retry in 2 min
				if res.StatusCode == 429 {
					time.Sleep(120000)
				}
				fmt.Printf("Error: %s\n", httperr)

			} else {
				break
			}
		}

		data, err := io.ReadAll(res.Body)
		if err != nil {
			fmt.Printf("Error: %s\n", err)
			//return err
		}

		if err := json.Unmarshal(data, &intermediateGame); err != nil {
			fmt.Printf("Error: %s\n", err)
			//return err
		}

		queue.PendingMatches <- intermediateGame.Metadata.MatchID
	}

	//queue.PendingMatches = append(queue.PendingMatches, intermediateGame.Metadata.MatchID)
}

func AddGameToDB(db *sql.DB) {

}

/*
	log start time

	if rate limit has not been hit do goroutines

	if rate limit hit wait unti
*/

func Crawl(url string, db *sql.DB) {

	go ExtractMatchIds(&queue)
	go ExtractMatchData()
	go AddGameToDB(db)

}
