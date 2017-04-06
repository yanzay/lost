package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/boltdb/bolt"
	"github.com/yanzay/log"
	"github.com/yanzay/lost"
	"github.com/yanzay/lost/parser"
)

type SerialList struct {
	Serials map[int]string
}

var (
	userID   = os.Getenv("LOST_USER_ID")
	password = os.Getenv("LOST_PASSWORD")
)

var db *bolt.DB
var subscriptions = []int{279, 162, 282}

func main() {
	flag.Parse()
	db = initDB()
	defer db.Close()
	log.Infof("Parsing subscriptions: %v", subscriptions)
	serials := parseSerials(subscriptions)
	if len(serials) == 0 {
		log.Info("No new episodes")
	}
	for _, serial := range serials {
		log.Infof("Saving serial: %s", serial.Name)
		saveSerial(serial)
	}
}

func initDB() *bolt.DB {
	var err error
	db, err = bolt.Open("my.db", 0600, nil)
	if err != nil {
		log.Fatal(err)
	}
	err = db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists([]byte("serials"))
		return err
	})
	if err != nil {
		log.Fatal(err)
	}
	return db
}

func parseSerials(subscriptions []int) []lost.Serial {
	f, err := os.Open("serials.json")
	if err != nil {
		log.Fatal(err)
	}
	list := &SerialList{}
	decoder := json.NewDecoder(f)
	decoder.Decode(list)
	pars := parser.NewParser(userID, password)
	serials := make([]lost.Serial, 0, len(subscriptions))
	for _, id := range subscriptions {
		saved, err := loadSerial(id)
		if err != nil {
			log.Warning(err)
		}
		episodes, err := pars.ListAllEpisodes(id)
		if err != nil {
			log.Error(err)
			continue
		}
		serial := lost.Serial{ID: id, Name: list.Serials[id], Episodes: episodes}
		serial = mergeEpisodes(*saved, serial)
		for _, episode := range serial.Episodes {
			if episode.Link == "" {
				episode.Link, err = pars.GetLink(id, episode.Season, episode.Number)
				time.Sleep(200 * time.Millisecond)
				if err != nil {
					log.Error(err)
					continue
				}
			}
		}
		serials = append(serials, serial)
	}
	return serials
}

func mergeEpisodes(saved, parsed lost.Serial) lost.Serial {
	for _, parsedEpisode := range parsed.Episodes {
		for _, savedEpisode := range saved.Episodes {
			if savedEpisode.Number == parsedEpisode.Number && savedEpisode.Season == parsedEpisode.Season {
				if savedEpisode.Link != "" {
					parsedEpisode.Link = savedEpisode.Link
				}
			}
		}
	}
	return parsed
}

func loadSerial(id int) (*lost.Serial, error) {
	var serialized []byte
	err := db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("serials"))
		serialized = b.Get([]byte(fmt.Sprint(id)))
		return nil
	})
	if err != nil {
		return nil, err
	}
	serial := &lost.Serial{}
	err = json.Unmarshal(serialized, &serial)
	return serial, err
}

func saveSerial(serial lost.Serial) {
	filterWithoutLinks(serial)
	serialized, err := json.Marshal(serial)
	if err != nil {
		log.Error(err)
	}
	log.Debug(string(serialized))
	err = db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("serials"))
		return b.Put([]byte(fmt.Sprint(serial.ID)), serialized)
	})
	if err != nil {
		log.Error(err)
	}
}

func filterWithoutLinks(serial lost.Serial) {
	episodes := make([]*lost.Episode, 0)
	for _, episode := range serial.Episodes {
		if episode.Link != "" {
			episodes = append(episodes, episode)
		}
	}
	serial.Episodes = episodes
}
