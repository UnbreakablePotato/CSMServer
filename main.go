package main

import (
	"log"
	"net/http"
)

// lane will be from
func GetTierList(w http.ResponseWriter, req *http.Request) {
	//use struct data for lane assignment
	// tierlist.lane := req.PathValue("lane")
}

func GetTierListAll(w http.ResponseWriter, req *http.Request) {

}

func GetChampionData(w http.ResponseWriter, req *http.Request) {

}

func GetMatchupData(w http.ResponseWriter, req *http.Request) {

}

func PostMatchupNotes(w http.ResponseWriter, req *http.Request) {

}

func EditMatchupNotes(w http.ResponseWriter, req *http.Request) {

}

func main() {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /tierlist", GetTierListAll)
	mux.HandleFunc("GET /tierlist?lane={lane}", GetTierList)
	mux.HandleFunc("GET /champions/{chammpion}/build", GetChampionData)
	mux.HandleFunc("GET /champions/{champion}/build?opponent={opponent}", GetMatchupData)
	mux.HandleFunc("POST /notes/{champion}?opponent={opponent}/{id}", PostMatchupNotes)
	mux.HandleFunc("PUT /notes/{champion}?opponent={opponent}/{id}", EditMatchupNotes)

	err := http.ListenAndServe(":8080", mux)
	log.Fatal(err)
}
