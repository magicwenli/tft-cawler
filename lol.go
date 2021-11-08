package tft

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"golang.org/x/time/rate"
)

// NewApiClient Creates a new api client using the provided api key
func NewApiClient(apiKey string) ApiClient {
	// Limits us to 96 requests every 2 minutes, just 4 below the bucket allows
	limiter := rate.NewLimiter(4, 4)
	return ApiClient{
		apiKey:   apiKey,
		limiter:  limiter,
		platform: "https://na1.api.riotgames.com",
		region:   "https://americas.api.riotgames.com",
	}
}

// rawRequest makes the actual http call, and will wait before doing so according to the limiter
func (c *ApiClient) rawRequest(url string) (*http.Response, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Add("X-Riot-Token", c.apiKey)

	err = c.limiter.Wait(context.Background())
	if err != nil {
		return nil, err
	}

	return http.DefaultClient.Do(req)
}

// request will call rawRequest and retry on 429s to get around service-level rate limiting
func (c *ApiClient) request(server, endpoint string) (*http.Response, error) {
	for i := 1; i < 4; i++ {
		res, err := c.rawRequest(server + endpoint)
		if err != nil {
			return nil, err
		}
		switch res.StatusCode {
		case 200, 404:
			return res, err
		case 400, 401, 403:
			log.Fatalln(res.Status, server+endpoint)
		case 429:
			_ = res.Body.Close()

			retryAfter, err := strconv.Atoi(res.Header.Get("Retry-After"))
			if err != nil {
				retryAfter = 15
			}

			if i == 3 {
				break
			}

			retryAfter = retryAfter*i + 5

			log.Println("Received a 429, waiting", retryAfter, "seconds")
			time.Sleep(time.Duration(retryAfter) * time.Second)
		case 500, 502, 503, 504:
			_ = res.Body.Close()
			if i == 3 {
				break
			}
			log.Println("Received", res.Status, ", waiting 1 minutes")
			time.Sleep(1 * time.Minute)
		default:
			log.Fatalln("Received unexpected status code", res.Status)
		}
	}
	log.Fatalln("too many 429s or 5xxs on", server+endpoint)
	return nil, nil
}

func (c *ApiClient) GetEntriesBySummoner(summonerId string) (*EntryDto, error) {
	endpoint := fmt.Sprintf("/tft/league/v1/entries/by-summoner/%v", summonerId)

	res, err := c.request(c.platform, endpoint)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	buf, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}
	var players []EntryDto

	if res.StatusCode == 404 {
		return nil, nil
	}

	if err := json.Unmarshal(buf, &players); err != nil {
		return nil, err
	}

	return &players[0], nil
}

func (c *ApiClient) GetEntries(r Rank, d Division, page int) ([]EntryDto, error) {
	// Make the endpoint string
	if page < 1 {
		return nil, errors.New("page number cannot be less than 1")
	}

	if (r == RankChall || r == RankGM || r == RankM) && d != Div1 {
		return nil, errors.New("apex tiers only use division 1")
	}
	endpoint := fmt.Sprintf("/tft/league/v1/entries/%v/%v?page=%v", Ranks[r],
		Divisions[d], page)

	res, err := c.request(c.platform, endpoint)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	buf, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}
	var players []EntryDto

	if res.StatusCode == 404 {
		return nil, nil
	}

	if err := json.Unmarshal(buf, &players); err != nil {
		return nil, err
	}

	return players, nil
}

func (c *ApiClient) GetLeagueList(r Rank) (*LeagueListDto, error) {
	if r > RankM {
		return nil, errors.New("apex tiers only use division 1")
	}
	endpoint := fmt.Sprintf("/tft/league/v1/%v", strings.ToLower(Ranks[r]))
	res, err := c.request(c.platform, endpoint)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	buf, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}
	var players *LeagueListDto

	if res.StatusCode == 404 {
		return nil, nil
	}

	if err := json.Unmarshal(buf, &players); err != nil {
		return nil, err
	}

	return players, nil
}

func (c *ApiClient) GetMatchesForPuuid(puuid string, count int) (*Matches, error) {
	endpoint := fmt.Sprintf("/tft/match/v1/matches/by-puuid/%v/ids?count=%v", puuid, count)
	res, err := c.request(c.region, endpoint)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	buf, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}
	var matches *Matches

	if res.StatusCode == 404 {
		return nil, nil
	}

	if err := json.Unmarshal(buf, &matches); err != nil {
		return nil, err
	}

	return matches, nil
}

func (c *ApiClient) GetMatches(matchId string) (*MatchDto, error) {
	endpoint := fmt.Sprintf("/tft/match/v1/matches/%v", matchId)
	res, err := c.request(c.region, endpoint)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	buf, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}
	var matches *MatchDto

	if res.StatusCode == 404 {
		return nil, nil
	}

	if err := json.Unmarshal(buf, &matches); err != nil {
		return nil, err
	}

	return matches, nil
}

func (c *ApiClient) GetSummonerByName(summonerName string) (*SummonerDto, error) {

	endpoint := fmt.Sprintf("/tft/summoner/v1/summoners/by-name/%v", summonerName)
	res, err := c.request(c.platform, endpoint)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	buf, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}
	var summoner *SummonerDto

	if res.StatusCode == 404 {
		return nil, nil
	}

	if err := json.Unmarshal(buf, &summoner); err != nil {
		return nil, err
	}

	return summoner, nil
}

func (c *ApiClient) GetSummonerByPuuid(Puuid string) (*SummonerDto, error) {

	endpoint := fmt.Sprintf("/tft/summoner/v1/summoners/by-puuid/%v", Puuid)
	res, err := c.request(c.platform, endpoint)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	buf, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}
	var summoner *SummonerDto

	if res.StatusCode == 404 {
		return nil, nil
	}

	if err := json.Unmarshal(buf, &summoner); err != nil {
		return nil, err
	}

	return summoner, nil
}

func (c *ApiClient) GetSummoner(summonerId string) (*SummonerDto, error) {

	endpoint := fmt.Sprintf("/tft/summoner/v1/summoners/%v", summonerId)
	res, err := c.request(c.platform, endpoint)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	buf, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}
	var summoner *SummonerDto

	if res.StatusCode == 404 {
		return nil, nil
	}

	if err := json.Unmarshal(buf, &summoner); err != nil {
		return nil, err
	}

	return summoner, nil
}
