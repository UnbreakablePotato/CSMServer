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
	MatchData      chan api.Game
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
				q.VisitedMatches[result[i]] = true
			} else {
				//huh
				//q.VisitedMatches[result[i]] = true
			}
			q.mu.Unlock()
		}
	}
}

func ExtractMatchData() {

	for {
		var intermediateGame api.Game
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

		//queue.PendingMatches <- intermediateGame.Metadata.MatchID

		queue.MatchData <- intermediateGame
	}

	//queue.PendingMatches = append(queue.PendingMatches, intermediateGame.Metadata.MatchID)
}

func AddGameToDB(db *sql.DB) {
	for {
		match := <-queue.MatchData

		query := `INSERT INTO games (match_id, data_version, end_of_game_result, game_creation,
			game_duration, game_end_timestamp, game_id, game_mode, game_name, game_start_timestamp,
			game_type, game_version, map_id, platform_id, queue_id, tournament_code)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?);`
		_, err := db.Exec(query, match.Metadata.MatchID, match.Metadata.DataVersion, match.Info.EndOfGameResult,
			match.Info.GameCreation, match.Info.GameDuration, match.Info.GameEndTimestamp, match.Info.GameID, match.Info.GameMode,
			match.Info.GameName, match.Info.GameStartTimestamp, match.Info.GameType, match.Info.GameVersion, match.Info.MapID,
			match.Info.PlatformID, match.Info.QueueID, match.Info.TournamentCode)
		if err != nil {
			fmt.Printf("Error adding game: %s", err)
		}

		query = `INSERT INTO participants (
			match_id,
	 		participant_id,
	  		puuid, summoner_id,
	   		summoner_name,
			summoner_level,
			riot_id_game_name,
			riot_id_tagline,
			profile_icon,
			champion_id,
			champion_name,
			champion_level,
			champion_experience,
			team_id, team_position,
			individual_position,
			lane,
			role,
			kills,
			deaths,
			assists,
			win,
			gold_earned,
			gold_spent,
			total_minions_killed,
			neutral_minions_killed,
			total_damage_dealt,
			total_damage_dealt_to_champions,
			total_damage_taken,
			damage_self_mitigated,
			total_heal,
			vision_score,
			wards_placed,
			ward_killed,
			item_0,
			item_1,
			item_2,
			item_3,
			item_4,
			item_5,
			item_6,
			summoner_spell_1_id,
			summoner_spell_2_id,
			first_blood_kill,
			first_blood_assist,
			first_tower_kill,
			first_tower_assist,
			double_kills,
			triple_kills,
			quadra_kills,
			penta_kills)
			VALUES(?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, 
			?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?,
			?, ?, ?, ?, ?, ?);`
		for i := range match.Info.Participants {
			_, err = db.Exec(query,
				match.Metadata.MatchID,
				match.Info.Participants[i].ParticipantID,
				match.Info.Participants[i].Puuid,
				match.Info.Participants[i].SummonerID,
				match.Info.Participants[i].SummonerName,
				match.Info.Participants[i].SummonerLevel,
				match.Info.Participants[i].RiotIDGameName,
				match.Info.Participants[i].RiotIDTagline,
				match.Info.Participants[i].ProfileIcon,
				match.Info.Participants[i].ChampionID,
				match.Info.Participants[i].ChampionName,
				match.Info.Participants[i].ChampLevel,
				match.Info.Participants[i].ChampExperience,
				match.Info.Participants[i].TeamID,
				match.Info.Participants[i].TeamPosition,
				match.Info.Participants[i].IndividualPosition,
				match.Info.Participants[i].Lane,
				match.Info.Participants[i].Role,
				match.Info.Participants[i].Kills,
				match.Info.Participants[i].Deaths,
				match.Info.Participants[i].Assists,
				match.Info.Participants[i].Win,
				match.Info.Participants[i].GoldEarned,
				match.Info.Participants[i].GoldSpent,
				match.Info.Participants[i].TotalMinionsKilled,
				match.Info.Participants[i].NeutralMinionsKilled,
				match.Info.Participants[i].TotalDamageDealt,
				match.Info.Participants[i].TotalDamageDealtToChampions,
				match.Info.Participants[i].DamageSelfMitigated,
				match.Info.Participants[i].TotalHeal,
				match.Info.Participants[i].VisionScore,
				match.Info.Participants[i].WardsPlaced,
				match.Info.Participants[i].WardsKilled,
				match.Info.Participants[i].Item0,
				match.Info.Participants[i].Item1,
				match.Info.Participants[i].Item2,
				match.Info.Participants[i].Item3,
				match.Info.Participants[i].Item4,
				match.Info.Participants[i].Item5,
				match.Info.Participants[i].Item6,
				match.Info.Participants[i].Summoner1ID,
				match.Info.Participants[i].Summoner2ID,
				match.Info.Participants[i].FirstBloodKill,
				match.Info.Participants[i].FirstBloodAssist,
				match.Info.Participants[i].FirstTowerKill,
				match.Info.Participants[i].FirstTowerAssist,
				match.Info.Participants[i].DoubleKills,
				match.Info.Participants[i].TripleKills,
				match.Info.Participants[i].QuadraKills,
				match.Info.Participants[i].PentaKills)
			if err != nil {
				fmt.Printf("Error adding participants: %s", err)
			}
		}
	}

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
