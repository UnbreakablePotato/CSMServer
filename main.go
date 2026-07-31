package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	api "github.com/UnbreakablePotato/CSMServer/internal/API"
)

// lane will be from
func GetTierList(w http.ResponseWriter, req *http.Request) {
	//use struct data for lane assignment
	// tierlist.lane := req.PathValue("lane")

	lane := req.URL.Query().Get("lane")

	if lane == "" {

	}
}

func GetTierListAll(w http.ResponseWriter, req *http.Request) {

}

var testChampion = api.Champion{
	ChampionId:              1,
	ChampionIcon:            "path as a string",
	RecommendedPerks:        []int{1, 2, 3},
	RecommendedSpells:       []int{4, 5},
	RecommendedAbilityOrder: []string{"W", "Q", "E"},
	RecommendedStartItems:   []int{6, 7},
	RecommendedItems:        []int{8, 9, 10, 11, 12, 13},
}

func GetChampionData(w http.ResponseWriter, req *http.Request) {
	// IMPORTANT! the user can type anything into the request, if the request contains
	// an unknown champion return an error instead

	w.Header().Set("Content-Type", "application/json")

	if err := json.NewEncoder(w).Encode(testChampion); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
	}

}

var matchup = api.Matchup{
	ChampionName:      "Riven",
	EnemyChampionName: "Darius",
	EnemyChampIcon:    "Some other path",
	Build:             testChampion,
}

func GetMatchupData(w http.ResponseWriter, req *http.Request) {
	opponent := req.URL.Query().Get("opponent")
	fmt.Printf("%s\n", opponent)
	/*
		If opponent exists in db get data for champ vs opponent
			create matchup struct
		else
			return a 404 error
	*/
	w.Header().Set("Content-Type", "application/json")

	if err := json.NewEncoder(w).Encode(matchup); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
	}
}

func PostMatchupNotes(w http.ResponseWriter, req *http.Request) {

}

func EditMatchupNotes(w http.ResponseWriter, req *http.Request) {

}

func GetMatchupNotes(w http.ResponseWriter, req *http.Request) {

}

func main() {

	db, derr := sql.Open("sqlite", "./my.db")
	if derr != nil {
		fmt.Printf("DB could not open: %s\n", derr)
	}

	if _, err := db.Exec("PRAGMA foreign_keys = ON;"); err != nil {
		fmt.Printf("Error: %s", err)
	}

	mux := http.NewServeMux()

	mux.HandleFunc("GET /tierlist/all", GetTierListAll)
	mux.HandleFunc("GET /tierlist", GetTierList)
	mux.HandleFunc("GET /champions/{champion}/build", GetChampionData)
	///champions/{champion}/build?opponent={opponent}
	mux.HandleFunc("GET /champions/{champion}/build/matchup", GetMatchupData)
	///notes/{champion}?opponent={opponent}/{id}
	mux.HandleFunc("POST /notes/{champion}", PostMatchupNotes)
	mux.HandleFunc("PUT /notes/{champion}", EditMatchupNotes)
	mux.HandleFunc("GET /notes/{champion}", GetMatchupNotes)

	err := http.ListenAndServe(":8080", mux)

	if err != nil {
		db.Close()
	}

	log.Fatal(err)
}
