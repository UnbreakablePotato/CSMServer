package api

type Tierlist struct {
	Lane      int
	Champions []struct {
		ChampionName string
		ChampionId   int
		Lane         string
		WinRate      float32
		PickRate     float32
		BanRate      float32
		NumGames     int
		Rank         int
	}
}

type Champion struct {
	ChampionId              int
	ChampionIcon            string
	RecommendedPerks        []int
	RecommendedSpells       []int    // example [1, 2]
	RecommendedAbilityOrder []string // example ["Q","W","E"]
	RecommendedStartItems   []int    // should consist of a list of start item ids
	RecommendedItems        []int    // should consist of a list of item ids
	BestMatchups            []int    // should consist of a list of champions ids
	WorstMatchups           []int    // should consist of a list of champions ids
}

type LiveGame struct {
	Player []struct {
		GameName       string
		Tag            string
		ChampionIcon   string
		AverageKills   float32
		AverageDeaths  float32
		AverageAssists float32
		SoloRank       string
		SoloRankIcon   string
		Lane           string
		MainRole       string
		WinStreak      bool
	}
}
