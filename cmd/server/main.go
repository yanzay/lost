package main

import "github.com/gorilla/feeds"

func main() {
	feed := &feeds.Feed{
		Title:       "lost",
		Link:        "example.com",
		Description: "torrent feed",
	}
	episodes := loadAllEpisodes()
	items := make([]*feed.Item, 0)
	for _, episode := range episodes {
		newItem := &feed.Item{
			Title:       episode.Name,
			Link:        episode.Link,
			Description: episode.Name,
		}
	}
}
