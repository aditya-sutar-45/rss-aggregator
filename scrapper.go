package main

import (
	"context"
	"log"
	"sync"
	"time"

	"github.com/aditya-sutar-45/rss-aggregator/internal/database"
)

// long running job
// db -> connection to teh database
// concurrency -> how many different go routines we want to do the scrapping
// timeBetweenRequests -> how much time between each request
func startScrapping(db *database.Queries, concurrency int, timeBetweenRequests time.Duration) {
	log.Printf("scrapping on %v go routines every %s Duration", concurrency, timeBetweenRequests)

	ticker := time.NewTicker(timeBetweenRequests)
	// run this for loop every time a new value if fed into the channel of a ticker
	// so run this for loop every "timeBetweenRequests" Duration
	for ; ; <-ticker.C {
		feeds, err := db.GetNextFeedsToFetch(
			context.Background(),
			int32(concurrency),
		)
		if err != nil {
			log.Println("error fetchign feeds: ", err)
			continue
		}

		wg := &sync.WaitGroup{}
		for _, feed := range feeds {
			wg.Add(1)

			go scrapeFeed(wg, db, feed)
		}
		wg.Wait()
	}
}

func scrapeFeed(wg *sync.WaitGroup, db *database.Queries, feed database.Feed) {
	defer wg.Done()

	_, err := db.MarkFeedAsFetched(context.Background(), feed.ID)
	if err != nil {
		log.Fatalln("error in marking feed as fetched: ", err)
		return
	}

	rssFeed, err := urlToFeed(feed.Url)
	if err != nil {
		log.Fatalln("error error fetching feed: ", err)
		return
	}

	for _, item := range rssFeed.Channel.Item {
		log.Println("found post: ", item.Title, " on feed: ", feed.Name)
	}
	log.Printf("feed %s collected, %v posts found", feed.Name, len(rssFeed.Channel.Item))
}
