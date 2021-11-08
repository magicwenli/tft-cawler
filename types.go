package tft

import "golang.org/x/time/rate"

const (
	queueType = "RANKED_SOLO_5x5"
)

type ApiClient struct {
	apiKey   string
	limiter  *rate.Limiter
	platform string
	region   string
}

type Rank int
type Division int
type Matches []string

const (
	RankChall Rank = iota
	RankGM
	RankM
	RankD
	RankP
	RankG
	RankS
	RankB
	RankI
)

const (
	Div1 Division = iota
	Div2
	Div3
	Div4
)

var Ranks = map[Rank]string{
	RankChall: "CHALLENGER",
	RankGM:    "GRANDMASTER",
	RankM:     "MASTER",
	RankD:     "DIAMOND",
	RankP:     "PLATINUM",
	RankG:     "GOLD",
	RankS:     "SILVER",
	RankB:     "BRONZE",
	RankI:     "IRON",
}

var Divisions = map[Division]string{
	Div1: "I",
	Div2: "II",
	Div3: "III",
	Div4: "IV",
}

type LeagueEntryDto struct {
	SummonerId   string `json:"summonerId"`
	SummonerName string `json:"summonerName"`
	Wins         int    `json:"wins"`
	Losses       int    `json:"losses"`
}

type LeagueListDto struct {
	Tier     string `json:"tier"`
	LeagueID string `json:"leagueId"`
	Queue    string `json:"queue"`
	Name     string `json:"name"`
	Entries  []struct {
		SummonerID   string `json:"summonerId"`
		SummonerName string `json:"summonerName"`
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

type EntryDto struct {
	LeagueID     string `json:"leagueId"`
	QueueType    string `json:"queueType"`
	Tier         string `json:"tier"`
	Rank         string `json:"rank"`
	SummonerID   string `json:"summonerId"`
	SummonerName string `json:"summonerName"`
	LeaguePoints int    `json:"leaguePoints"`
	Wins         int    `json:"wins"`
	Losses       int    `json:"losses"`
	Veteran      bool   `json:"veteran"`
	Inactive     bool   `json:"inactive"`
	FreshBlood   bool   `json:"freshBlood"`
	HotStreak    bool   `json:"hotStreak"`
}

type MatchDto struct {
	Metadata Metadata `json:"metadata"`
	Info     Info     `json:"info"`
}
type Metadata struct {
	DataVersion  string   `json:"data_version"`
	MatchID      string   `json:"match_id"`
	Participants []string `json:"participants"`
}
type Companion struct {
	ContentID string `json:"content_ID"`
	SkinID    int    `json:"skin_ID"`
	Species   string `json:"species"`
}
type Traits struct {
	Name        string `json:"name"`
	NumUnits    int    `json:"num_units"`
	Style       int    `json:"style"`
	TierCurrent int    `json:"tier_current"`
	TierTotal   int    `json:"tier_total,omitempty"`
}
type Units struct {
	CharacterID string `json:"character_id"`
	Items       []int  `json:"items"`
	Name        string `json:"name"`
	Rarity      int    `json:"rarity"`
	Tier        int    `json:"tier"`
}
type Participants struct {
	Companion            Companion `json:"companion"`
	GoldLeft             int       `json:"gold_left"`
	LastRound            int       `json:"last_round"`
	Level                int       `json:"level"`
	Placement            int       `json:"placement"`
	PlayersEliminated    int       `json:"players_eliminated"`
	Puuid                string    `json:"puuid"`
	TimeEliminated       float64   `json:"time_eliminated"`
	TotalDamageToPlayers int       `json:"total_damage_to_players"`
	Traits               []Traits  `json:"traits"`
	Units                []Units   `json:"units"`
}
type Info struct {
	GameDatetime  int64          `json:"game_datetime"`
	GameLength    float64        `json:"game_length"`
	GameVariation string         `json:"game_variation"`
	GameVersion   string         `json:"game_version"`
	Participants  []Participants `json:"participants"`
	QueueID       int            `json:"queue_id"`
	TftSetNumber  int            `json:"tft_set_number"`
}

type SummonerDto struct {
	ID            string `json:"id"`
	AccountID     string `json:"accountId"`
	Puuid         string `json:"puuid"`
	Name          string `json:"name"`
	ProfileIconID int    `json:"profileIconId"`
	RevisionDate  int64  `json:"revisionDate"`
	SummonerLevel int    `json:"summonerLevel"`
}
