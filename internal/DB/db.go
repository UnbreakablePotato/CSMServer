package db

import (
	"database/sql"

	_ "modernc.org/sqlite"
)

func CreateChampionTable(db *sql.DB) (sql.Result, error) {

	query := `CREATE TABLE IF NOT EXISTS champion (
		ChampionId INTEGER NOT NULL,
		ChampionIcon TEXT,
		RecommendedPerk1 INTEGER NOT NULL,
		RecommendedPerk2 INTEGER NOT NULL,
		RecommendedPerk3 INTEGER NOT NULL,
		RecommendedPerk4 INTEGER NOT NULL,
		RecommendedPerk5 INTEGER NOT NULL,
		RecommendedPerk6 INTEGER NOT NULL,
		RecommendedPerk7 INTEGER NOT NULL,
		RecommendedPerk8 INTEGER NOT NULL,
		RecommendedPerk9 INTEGER NOT NULL,
		RecommendedAbility1 TEXT NOT NULL,
		RecommendedStart1 INTEGER NOT NULL,
		RecommendedStart2 INTEGER NOT NULL,
		RecommendedItem1 INTEGER NOT NULL,
		RecommendedItem2 INTEGER NOT NULL,
		RecommendedItem3 INTEGER NOT NULL,
		RecommendedItem4 INTEGER NOT NULL,
		RecommendedItem5 INTEGER NOT NULL,
		RecommendedItem6 INTEGER NOT NULL,
		RecommendedItem7 INTEGER,
	);`

	return db.Exec(query)
}

func CreateGameTable(db *sql.DB) (sql.Result, error) {

	query := `CREATE TABLE IF NOT EXISTS games (
    match_id              TEXT PRIMARY KEY,
    data_version          TEXT NOT NULL,

    end_of_game_result    TEXT,
    game_creation         INTEGER NOT NULL,
    game_duration         INTEGER NOT NULL,
    game_end_timestamp    INTEGER NOT NULL,
    game_id               INTEGER NOT NULL UNIQUE,
    game_mode             TEXT NOT NULL,
    game_name             TEXT,
    game_start_timestamp  INTEGER NOT NULL,
    game_type             TEXT NOT NULL,
    game_version          TEXT NOT NULL,
    map_id                INTEGER NOT NULL,
    platform_id           TEXT NOT NULL,
    queue_id              INTEGER NOT NULL,
    tournament_code       TEXT
	);`

	return db.Exec(query)
}

func CreateParticipantsTable(db *sql.DB) (sql.Result, error) {

	query := `CREATE TABLE IF NOT EXISTS participants (
    match_id                         TEXT NOT NULL,
    participant_id                   INTEGER NOT NULL,
    puuid                            TEXT NOT NULL,

    summoner_id                      TEXT,
    summoner_name                    TEXT,
    summoner_level                   INTEGER,
    riot_id_game_name                TEXT,
    riot_id_tagline                  TEXT,
    profile_icon                     INTEGER,

    champion_id                      INTEGER NOT NULL,
    champion_name                    TEXT NOT NULL,
    champion_level                   INTEGER,
    champion_experience              INTEGER,

    team_id                          INTEGER NOT NULL,
    team_position                    TEXT,
    individual_position              TEXT,
    lane                             TEXT,
    role                             TEXT,

    kills                            INTEGER NOT NULL DEFAULT 0,
    deaths                           INTEGER NOT NULL DEFAULT 0,
    assists                          INTEGER NOT NULL DEFAULT 0,
    win                              INTEGER NOT NULL CHECK (win IN (0, 1)),

    gold_earned                      INTEGER,
    gold_spent                       INTEGER,
    total_minions_killed             INTEGER,
    neutral_minions_killed           INTEGER,

    total_damage_dealt               INTEGER,
    total_damage_dealt_to_champions  INTEGER,
    total_damage_taken               INTEGER,
    damage_self_mitigated            INTEGER,
    total_heal                       INTEGER,

    vision_score                     INTEGER,
    wards_placed                     INTEGER,
    wards_killed                     INTEGER,

    item_0                           INTEGER,
    item_1                           INTEGER,
    item_2                           INTEGER,
    item_3                           INTEGER,
    item_4                           INTEGER,
    item_5                           INTEGER,
    item_6                           INTEGER,

    summoner_spell_1_id              INTEGER,
    summoner_spell_2_id              INTEGER,

    first_blood_kill                 INTEGER CHECK (first_blood_kill IN (0, 1)),
    first_blood_assist               INTEGER CHECK (first_blood_assist IN (0, 1)),
    first_tower_kill                 INTEGER CHECK (first_tower_kill IN (0, 1)),
    first_tower_assist               INTEGER CHECK (first_tower_assist IN (0, 1)),

    double_kills                     INTEGER,
    triple_kills                     INTEGER,
    quadra_kills                     INTEGER,
    penta_kills                      INTEGER,

    PRIMARY KEY (match_id, participant_id),
    FOREIGN KEY (match_id)
        REFERENCES games(match_id)
        ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_participants_puuid
    ON participants(puuid);

CREATE INDEX IF NOT EXISTS idx_participants_champion
    ON participants(champion_id);

CREATE INDEX IF NOT EXISTS idx_participants_match_team
    ON participants(match_id, team_id);`

	return db.Exec(query)
}

type Build struct {
	ID           int
	ChampionID   int
	ChampionName string
	Position     string

	Item0 int
	Item1 int
	Item2 int
	Item3 int
	Item4 int
	Item5 int
	Item6 int

	SummonerSpell1 int
	SummonerSpell2 int

	Keystone int
	Perk1    int
	Perk2    int
	Perk3    int
	Perk4    int
	Perk5    int
	Perk6    int

	Games int
	Wins  int
}

func GetMostPopularBuild(db *sql.DB, championID int, position string) (*Build, error) {
	query := `
		SELECT
			id,
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
		FROM builds
		WHERE champion_id = ?
		  AND position = ?
		ORDER BY games DESC
		LIMIT 1;
	`

	var build Build

	err := db.QueryRow(query, championID, position).Scan(
		&build.ID,
		&build.ChampionID,
		&build.ChampionName,
		&build.Position,
		&build.Item0,
		&build.Item1,
		&build.Item2,
		&build.Item3,
		&build.Item4,
		&build.Item5,
		&build.Item6,
		&build.SummonerSpell1,
		&build.SummonerSpell2,
		&build.Keystone,
		&build.Perk1,
		&build.Perk2,
		&build.Perk3,
		&build.Perk4,
		&build.Perk5,
		&build.Perk6,
		&build.Games,
		&build.Wins,
	)

	if err != nil {
		return nil, err
	}

	return &build, nil
}

/*
func CreateMatchupTable(db *sql.DB) (sql.Result, error) {

	sql := `CREATE TABLE IF NOT EXISTS matchup (
	);`

	return db.Exec(sql)
}

*/
