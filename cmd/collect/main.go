package main

import (
	"context"
	"log"
	"os"
	"sync"
	lol "tft-cawler"
	"time"

	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo" //MongoDB的Go驱动包
	"go.mongodb.org/mongo-driver/mongo/options"
)

var (
	client   lol.ApiClient
	ec       chan error
	sc       chan string
	mc       chan string
	database *mongo.Database
	bFalse   bool
	wg       sync.WaitGroup
)

const pageCount = 20
const DEVMODE = true
const startMatch = "NA1_4092101718"

type idGenerate struct {
	ID  int
	KEY string
}

func Next(coll string) (int, error) {
	update := bson.M{"$inc": bson.M{"id": 1}}
	opt := options.FindOneAndUpdateOptions{Upsert: &bFalse}
	opt2 := &options.InsertOneOptions{}
	idG := idGenerate{ID: 1}
	err := database.Collection("idgen").
		FindOneAndUpdate(context.Background(), bson.M{"key": coll}, update, &opt).Decode(&idG)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			_, err := database.Collection("idgen").InsertOne(context.Background(), bson.M{"key": coll, "id": 1}, opt2)
			if err != nil {
				ec <- err
				return 0, nil
			} else {
				return 0, nil
			}
		} else {
			ec <- err
			return 0, err
		}
	} else {
		return idG.ID, nil
	}

}

func InsertSummonerByPuuid(puuid string) ([]string, error) {
	log.Println(puuid)
	coll := "summoner"
	res := database.Collection(coll).FindOne(context.Background(), bson.M{"puuid": puuid}, options.FindOne().SetProjection(bson.M{"puuid": 1, "_id": 0}))
	if res.Err() == mongo.ErrNoDocuments {
		summoner, err := client.GetSummonerByPuuid(puuid)
		if err != nil {
			ec <- err
			return nil, nil
		}
		summonerEntry, err := client.GetEntriesBySummoner(summoner.ID)
		if err != nil {
			ec <- err
			return nil, nil
		}
		next, _ := Next(coll)
		field := struct {
			ID            int
			Puuid         string
			SummonerLevel int
			SummonerTire  string
			SummonerRank  string
		}{next, puuid, summoner.SummonerLevel, summonerEntry.Tier, summonerEntry.Rank}
		buf, err := bson.Marshal(field)
		if err != nil {
			return nil, err
		}
		_, err = database.Collection(coll).InsertOne(context.Background(), buf)
		if err != nil && !mongo.IsDuplicateKeyError(err) {
			return nil, err
		}
		log.Println("insert summoner success: " + summoner.Name)
		matches, _ := client.GetMatchesForPuuid(summoner.Puuid, pageCount)
		return *matches, nil
	}
	log.Println(puuid + " already scanned")

	// This should be nil if the document was found. In that case we don't want to do anything
	// and we are good.
	return nil, res.Err()
}

func InsertMatch(matchId string) ([]string, error) {
	coll := "matches"
	res := database.Collection(coll).FindOne(context.Background(), bson.M{"matchid": matchId}, options.FindOne().SetProjection(bson.M{"matchid": 1, "_id": 0}))
	if res.Err() == mongo.ErrNoDocuments {
		match, err := client.GetMatches(matchId)
		if err != nil {
			ec <- err
			return nil, nil
		}
		next, _ := Next(coll)
		field := struct {
			Match   lol.MatchDto
			ID      int
			MatchID string
		}{*match, next, matchId}
		buf, err := bson.Marshal(field)
		if err != nil {
			return nil, err
		}
		_, err = database.Collection(coll).InsertOne(context.Background(), buf)
		if err != nil && !mongo.IsDuplicateKeyError(err) {
			return nil, err
		}
		log.Println("insert match success: " + match.Metadata.MatchID)
		puuids := match.Metadata.Participants
		return puuids, nil
	}
	log.Println(matchId + " already scanned")
	// This should be nil if the document was found. In that case we don't want to do anything
	// and we are good.
	return nil, res.Err()
}

func main() {
	log.Println("===== Collector started =====")
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found")
	}
	mongoUrl := os.Getenv("MONGODB_URI")
	user := os.Getenv("MONGODB_USER")
	password := os.Getenv("MONGODB_PASS")
	collectionName := os.Getenv("MONGODB_COLLECTION")
	apiKey := os.Getenv("RGAPIKEY")

	if len(apiKey) == 0 || len(mongoUrl) == 0 || len(user) == 0 || len(password) == 0 || len(collectionName) == 0 {
		log.Fatal("not all needed environment variables are set")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	mongoOpts := options.Client().ApplyURI(mongoUrl)
	mongoOpts.SetAppName("LolMatchCollector")
	mongoOpts.SetAuth(options.Credential{
		AuthMechanism: "SCRAM-SHA-256",
		AuthSource:    "admin",
		Username:      user,
		Password:      password,
	})

	mongoClient, err := mongo.Connect(ctx, mongoOpts)
	if err != nil {
		log.Fatal(err)
	}
	defer func() {
		if err = mongoClient.Disconnect(ctx); err != nil {
			panic(err)
		}
	}()

	database = mongoClient.Database("admin")
	if DEVMODE {
		database.Collection("summoner").Drop(context.Background())
		database.Collection("matches").Drop(context.Background())
		database.Collection("idgen").Drop(context.Background())
	}
	client = lol.NewApiClient(apiKey)
	ec = make(chan error)
	sc = make(chan string, 40000)
	mc = make(chan string, 40000)
	go errorHandler()
	wg.Add(1)
	go summonerHandler()
	wg.Add(1)
	go matchHandler()

	summoner, err := InsertMatch(startMatch)
	if err != nil {
		ec <- err
	}
	for _, s := range summoner {
		res := database.Collection("summoner").FindOne(context.Background(), bson.M{"puuid": s})
		if res.Err() == mongo.ErrNoDocuments {
			sc <- s
		} else {
			log.Println(s + " already scanned")
		}
	}

	wg.Wait()

	for len(sc) > 0 {
		puuid := <-sc
		time.Sleep(300 * time.Millisecond)
		_, err := InsertSummonerByPuuid(puuid)
		if err != nil {
			ec <- err
		}
	}
	for len(mc) > 0 {
		time.Sleep(300 * time.Millisecond)
		matchId := <-mc
		_, err := InsertMatch(matchId)
		if err != nil {
			ec <- err
		}
	}
}

func summonerHandler() {
	for i := 0; i < 64; i = i + 1 {
		puuid := <-sc
		matches, err := InsertSummonerByPuuid(puuid)
		if err != nil {
			ec <- err
		}
		for _, m := range matches {
			time.Sleep(300 * time.Millisecond)
			res := database.Collection("matches").FindOne(context.Background(), bson.M{"matchid": m})
			if res.Err() == mongo.ErrNoDocuments {
				mc <- m
			} else {
				log.Println(m + " already scanned")
			}
		}
	}
	wg.Done()
}

func matchHandler() {
	for i := 0; i < 64; i = i + 1 {
		matchId := <-mc
		summoner, err := InsertMatch(matchId)
		if err != nil {
			ec <- err
		}
		for _, s := range summoner {
			time.Sleep(300 * time.Millisecond)
			res := database.Collection("summoner").FindOne(context.Background(), bson.M{"puuid": s})
			if res.Err() == mongo.ErrNoDocuments {
				sc <- s
			} else {
				log.Println(s + " already scanned")
			}
		}
	}
	wg.Done()
}

func errorHandler() {
	i := 0
	for {
		err := <-ec
		i++
		log.Printf("received error \"%v\", %v/5 before stopping\n", err, i)
		if i == 5 {
			log.Fatalln("too many errors. stopping.")
		}
	}
}
