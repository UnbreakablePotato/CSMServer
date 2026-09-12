package crawler

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	api "github.com/UnbreakablePotato/CSMServer/internal/API"
	"github.com/joho/godotenv"
	_ "modernc.org/sqlite"
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
	mu             sync.RWMutex
}

var queue = Queue{
	VisitedMatches: make(map[string]bool),
	PendingMatches: make(chan string, 1000), // Buffered so InitialRequest doesn't block
	PendingPuuids:  make(chan string, 1000),
	MatchData:      make(chan api.Game, 100),
}

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
	client := http.Client{}
	for {
		puuid := <-queue.PendingPuuids

		fullUrl := "https://europe.api.riotgames.com/lol/match/v5/matches/by-puuid/" + puuid + "/ids?start=0&count=5&api_key=" + apiKey
		for {
			req, err := http.NewRequest("GET", fullUrl, nil)
			if err != nil {
				fmt.Printf("Error: %s\n", err)

			}

			var res *http.Response

			var httperr error
			res, httperr = client.Do(req)
			//happens upon multiple errors
			if httperr != nil {
				fmt.Printf("Network error: %s\n", httperr)
				time.Sleep(5 * time.Second)
				continue
			}

			//Happens only when rate limit hit
			if res.StatusCode == 429 {
				retry := res.Header.Get("Retry-After")
				retrySeconds, _ := strconv.Atoi(retry)
				fmt.Printf("Rate limit hit! Sleeping for %s seconds...\n", retry)
				res.Body.Close()
				time.Sleep(time.Duration(retrySeconds)*time.Second + time.Second)
				continue
			}
			defer res.Body.Close()
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
			break
		}
	}
}

func ExtractMatchData() {
	client := http.Client{}
	for {
		var intermediateGame api.Game
		matchID := <-queue.PendingMatches
		fullUrl := "https://europe.api.riotgames.com/lol/match/v5/matches/" + matchID + "?api_key=" + apiKey
		for {
			req, err := http.NewRequest("GET", fullUrl, nil)
			if err != nil {
				fmt.Printf("Error: %s\n", err)
				//return err
			}

			var res *http.Response

			var httperr error
			res, httperr = client.Do(req)
			//httperr wil trigger ipon multiple 429 errors
			if httperr != nil {
				fmt.Printf("Error: %s\n", httperr)
				time.Sleep(5 * time.Second)
				continue
			}

			if res.StatusCode == 429 {
				retry := res.Header.Get("Retry-After")
				retrySeconds, _ := strconv.Atoi(retry)
				fmt.Printf("Rate limit hit! Sleeping for %s seconds...\n", retry)
				res.Body.Close()
				time.Sleep(time.Duration(retrySeconds)*time.Second + time.Second)
				continue
			}
			defer res.Body.Close()
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
			break
		}

	}

	//queue.PendingMatches = append(queue.PendingMatches, intermediateGame.Metadata.MatchID)
}

func AddBuildToDB(db *sql.DB, match api.Game) error {
	createTableQuery := `CREATE TABLE IF NOT EXISTS builds (
		champion_id INTEGER,
		champion_name TEXT,
		position TEXT,
		item_0 INTEGER,
		item_1 INTEGER,
		item_2 INTEGER,
		item_3 INTEGER,
		item_4 INTEGER,
		item_5 INTEGER,
		item_6 INTEGER,
		summoner_spell_1 INTEGER,
		summoner_spell_2 INTEGER,
		keystone INTEGER,
		perk_1 INTEGER,
		perk_2 INTEGER,
		perk_3 INTEGER,
		perk_4 INTEGER,
		perk_5 INTEGER,
		perk_6 INTEGER,
		games INTEGER,
		wins INTEGER,
		UNIQUE(
			champion_id, position, 
			item_0, item_1, item_2, item_3, item_4, item_5, item_6, 
			summoner_spell_1, summoner_spell_2, 
			keystone, perk_1, perk_2, perk_3, perk_4, perk_5, perk_6
		)
	);`

	if _, err := db.Exec(createTableQuery); err != nil {
		return fmt.Errorf("failed to create builds table: %w", err)
	}

	for i := range match.Info.Participants {
		participant := match.Info.Participants[i]

		if len(participant.Perks.Styles) < 2 {
			return fmt.Errorf("participant has insufficient perk styles")
		}

		primary := participant.Perks.Styles[0]
		secondary := participant.Perks.Styles[1]

		if len(primary.Selections) < 4 || len(secondary.Selections) < 2 {
			return fmt.Errorf("participant has insufficient perk selections")
		}

		keystone := primary.Selections[0].Perk
		perk1 := primary.Selections[1].Perk
		perk2 := primary.Selections[2].Perk
		perk3 := primary.Selections[3].Perk

		perk4 := secondary.Selections[0].Perk
		perk5 := secondary.Selections[1].Perk

		perk6 := 0

		query := `
        INSERT INTO builds (
            champion_id,
            champion_name,
            position,

            item_0,
            item_1,
            item_2,
            item_3,
            item_4,
            item_5,
            item_6,

            summoner_spell_1,
            summoner_spell_2,

            keystone,
            perk_1,
            perk_2,
            perk_3,
            perk_4,
            perk_5,
            perk_6,

            games,
            wins
        )
        VALUES (
            ?, ?, ?,
            ?, ?, ?, ?, ?, ?, ?,
            ?, ?,
            ?, ?, ?, ?, ?, ?, ?,
            1, ?
        )
        ON CONFLICT (
            champion_id,
            position,
            item_0,
            item_1,
            item_2,
            item_3,
            item_4,
            item_5,
            item_6,
            summoner_spell_1,
            summoner_spell_2,
            keystone,
            perk_1,
            perk_2,
            perk_3,
            perk_4,
            perk_5,
            perk_6
        )
        DO UPDATE SET
            games = games + 1,
            wins = wins + excluded.wins;
    `

		wins := 0

		if participant.Win {
			wins = 1
		}

		_, err := db.Exec(
			query,

			participant.ChampionID,
			participant.ChampionName,
			participant.TeamPosition,

			participant.Item0,
			participant.Item1,
			participant.Item2,
			participant.Item3,
			participant.Item4,
			participant.Item5,
			participant.Item6,

			participant.Summoner1ID,
			participant.Summoner2ID,

			keystone,
			perk1,
			perk2,
			perk3,
			perk4,
			perk5,
			perk6,

			wins,
		)
		if err != nil {
			return err
		}
	}

	return nil
}

func AddGameToDB(db *sql.DB) {
	for {
		match := <-queue.MatchData

		builderr := AddBuildToDB(db, match)
		if builderr != nil {
			fmt.Printf("Error adding build: %s\n", builderr)
		}

		createTableGameQuery := `CREATE TABLE IF NOT EXISTS games (
			match_id,
		 	data_version,
		  	end_of_game_result,
		    game_creation,

			game_duration,
			game_end_timestamp,
			game_id, game_mode,
			game_name,
			game_start_timestamp,
			game_type,
			game_version,
			map_id,
			platform_id,
			queue_id,
			tournament_code
			);`

		_, err := db.Exec(createTableGameQuery)
		if err != nil {
			fmt.Printf("Error creating table: %s\n", err)
			// You might want to return or exit here if the table fails to create
		}

		query := `INSERT INTO games (match_id,
		 	data_version,
		  	end_of_game_result,
		    game_creation,

			game_duration,
			game_end_timestamp,
			game_id, game_mode,
			game_name,
			game_start_timestamp,
			game_type,
			game_version,
			map_id,
			platform_id,
			queue_id,
			tournament_code)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?);`
		_, err = db.Exec(query,
			match.Metadata.MatchID,
			match.Metadata.DataVersion,
			match.Info.EndOfGameResult,
			match.Info.GameCreation,
			match.Info.GameDuration,
			match.Info.GameEndTimestamp,
			match.Info.GameID,
			match.Info.GameMode,
			match.Info.GameName,
			match.Info.GameStartTimestamp,
			match.Info.GameType,
			match.Info.GameVersion,
			match.Info.MapID,
			match.Info.PlatformID,
			match.Info.QueueID,
			match.Info.TournamentCode)
		if err != nil {
			fmt.Printf("Error adding game: %s", err)
		}

		createTableQuery := `CREATE TABLE IF NOT EXISTS participants (
			match_id TEXT,
			participant_id INTEGER,
			puuid TEXT,
			summoner_id TEXT,
			summoner_name TEXT,
			summoner_level INTEGER,
			riot_id_game_name TEXT,
			riot_id_tagline TEXT,
			profile_icon INTEGER,
			champion_id INTEGER,
			champion_name TEXT,
			champion_level INTEGER,
			champion_experience INTEGER,
			team_id INTEGER,
			team_position TEXT,
			individual_position TEXT,
			lane TEXT,
			role TEXT,
			kills INTEGER,
			deaths INTEGER,
			assists INTEGER,
			win BOOLEAN,
			gold_earned INTEGER,
			gold_spent INTEGER,
			total_minions_killed INTEGER,
			neutral_minions_killed INTEGER,
			total_damage_dealt INTEGER,
			total_damage_dealt_to_champions INTEGER,
			total_damage_taken INTEGER,
			damage_self_mitigated INTEGER,
			total_heal INTEGER,
			vision_score INTEGER,
			wards_placed INTEGER,
			ward_killed INTEGER,
			item_0 INTEGER,
			item_1 INTEGER,
			item_2 INTEGER,
			item_3 INTEGER,
			item_4 INTEGER,
			item_5 INTEGER,
			item_6 INTEGER,
			summoner_spell_1_id INTEGER,
			summoner_spell_2_id INTEGER,
			first_blood_kill BOOLEAN,
			first_blood_assist BOOLEAN,
			first_tower_kill BOOLEAN,
			first_tower_assist BOOLEAN,
			double_kills INTEGER,
			triple_kills INTEGER,
			quadra_kills INTEGER,
			penta_kills INTEGER,
			PRIMARY KEY (match_id, puuid)
			);`

		_, err = db.Exec(createTableQuery)
		if err != nil {
			fmt.Printf("Error creating table: %s\n", err)
			// You might want to return or exit here if the table fails to create
		}

		query = `INSERT OR IGNORE INTO participants (
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
				match.Info.Participants[i].TotalDamageTaken,
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
				fmt.Printf("Error adding participants into participant table: %s", err)
			}
		}
	}

}

/*
	log start time

	if rate limit has not been hit do goroutines

	if rate limit hit wait unti
*/

func Crawl(db *sql.DB) {
	err := InitialRequest()
	if err != nil {
		fmt.Printf("InitialRequest Error: %s\n", err)
	}
	go ExtractMatchIds(&queue)
	go ExtractMatchData()
	go AddGameToDB(db)

}
