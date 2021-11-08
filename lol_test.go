package tft

import "testing"

var c = NewApiClient("RGAPI-cdb6a866-bac4-464c-86fb-a154b4249fbd")

func TestApiClient_GetEntries(t *testing.T) {
	entries, err := c.GetEntries(3, 0, 1)
	if err != nil {
		panic(err)
	}
	t.Log(entries)
}

func TestApiClient_GetEntriesBySummoner(t *testing.T) {
	entries, err := c.GetEntriesBySummoner("pVjdi1gQfMXoHsjXOeFp4XLM6PPdcqz7ILD3rnHWazeuIZHZ")
	if err != nil {
		panic(err)
	}
	t.Log(entries)
}

func TestApiClient_GetLeagueList(t *testing.T) {
	entries, err := c.GetLeagueList(0)
	if err != nil {
		panic(err)
	}
	t.Log(entries)
}

func TestApiClient_GetMatchesForPuuid(t *testing.T) {
	entries, err := c.GetMatchesForPuuid("cfoLBzN1qZcWHZzl7khUEhSDMSGOPcpXlJH6dVRGxl1sKEiMgzn00X_Qc-gqhHENToxXHPDPaPKPpA", 20)
	if err != nil {
		panic(err)
	}
	t.Log(entries)
}

func TestApiClient_GetMatchesByPuuid(t *testing.T) {
	entries, err := c.GetMatchesForPuuid("cfoLBzN1qZcWHZzl7khUEhSDMSGOPcpXlJH6dVRGxl1sKEiMgzn00X_Qc-gqhHENToxXHPDPaPKPpA", 20)
	if err != nil {
		panic(err)
	}
	t.Log(entries)
}

func TestApiClient_GetMatches(t *testing.T) {
	entries, err := c.GetMatches("NA1_3367286856")
	if err != nil {
		panic(err)
	}
	t.Log(entries)
}

func TestApiClient_GetSummoner(t *testing.T) {
	summoner := SummonerDto{
		ID:            "7K2rAErpH_yOyPEhz8LX3ngg-_goStKrvdSI7zziURC5CaM",
		AccountID:     "h5xGHY8yOvJE90fZXoi3UcP1kEiNxBAb9j6awG5cbkdheQg",
		Puuid:         "cfoLBzN1qZcWHZzl7khUEhSDMSGOPcpXlJH6dVRGxl1sKEiMgzn00X_Qc-gqhHENToxXHPDPaPKPpA",
		Name:          "API",
		ProfileIconID: 4568,
		RevisionDate:  1635576039000,
		SummonerLevel: 102,
	}

	s1, err := c.GetSummoner(summoner.ID)
	if err != nil {
		panic(err)
	}
	if *s1 != summoner {
		t.Log("GetSummoner Failed")
	}
	s2, err := c.GetSummonerByPuuid(summoner.Puuid)
	if err != nil {
		panic(err)
	}
	if *s2 != summoner {
		t.Log("GetSummonerByPuuid Failed")
	}
	s3, err := c.GetSummonerByName(summoner.Name)
	if err != nil {
		panic(err)
	}
	if *s3 != summoner {
		t.Log("GetSummonerByName Failed")
	}

}
