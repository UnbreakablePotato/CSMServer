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
	PendingMatches []string
	PendingPuuids  []string
	mu             *sync.Mutex
}

var queue Queue

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

	queue.mu.Lock()
	for i := range challengers.Entries {
		queue.PendingPuuids = append(queue.PendingPuuids, challengers.Entries[i].Puuid)
	}
	queue.mu.Unlock()

	return nil
}

func ExtractMatchIds(q *Queue) ([]string, error) {

	fullUrl := "https://europe.api.riotgames.com/lol/match/v5/matches/by-puuid/" + q.PendingPuuids[0] + "/ids?start=0&count=5?api_key=" + apiKey

	//remove used puuid from pending
	q.PendingPuuids = append(q.PendingPuuids[:0], q.PendingPuuids[1:]...)

	req, err := http.NewRequest("GET", fullUrl, nil)
	if err != nil {
		fmt.Printf("Error: %s\n", err)
		return []string{}, err
	}

	client := http.Client{}

	res, err := client.Do(req)
	if err != nil {
		fmt.Printf("Error: %s\n", err)
		return []string{}, err
	}

	data, err := io.ReadAll(res.Body)
	if err != nil {
		fmt.Printf("Error: %s\n", err)
		return []string{}, err
	}

	interResult := string(data)

	result := strings.Split(interResult, ",")

	for i := range result {
		result[i] = strings.ReplaceAll(result[i], "\"", "")
		result[i] = strings.ReplaceAll(result[i], "[", "")
		result[i] = strings.ReplaceAll(result[i], "]", "")
		//fmt.Printf("debug: %s\n", result[i])
	}

	return result, nil
}

var intermediateGame api.Game

func ExtractMatchData(matches []string) error {

	for i := range matches {
		fullUrl := "https://europe.api.riotgames.com/lol/match/v5/matches/" + matches[i] + "?api_key=" + apiKey
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
			return err
		}

		if err := json.Unmarshal(data, &intermediateGame); err != nil {
			fmt.Printf("Error: %s\n", err)
			return err
		}
	}

	queue.PendingMatches = append(queue.PendingMatches, intermediateGame.Metadata.MatchID)

	return nil
}

func AddGameToDB(db *sql.DB) {

}

//goroutine which collects urls

func CollectUrls() error {

	return nil
}

func Crawl(url string, db *sql.DB) (bool, error) {

	return true, nil
}
