package db

import "database/sql"

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

/*
func CreateMatchupTable(db *sql.DB) (sql.Result, error) {

	sql := `CREATE TABLE IF NOT EXISTS matchup (
	);`

	return db.Exec(sql)
}

*/
