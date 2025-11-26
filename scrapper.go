package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

	"github.com/aditya-sutar-45/rss-aggregator/internal/database"
	"github.com/google/uuid"
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
		// log.Println("found post: ", item.Title, " on feed: ", feed.Name)

		description := sql.NullString{}
		if item.Description != "" {
			description.String = item.Description
			description.Valid = true
		}

		pubTime, err := parseDate(item.PubDate)
		if err != nil {
			log.Println("error: ", err)
			continue
		}

		_, err = db.CreatePost(context.Background(), database.CreatePostParams{
			ID:          uuid.New(),
			CreatedAt:   time.Now().UTC(),
			UpdatedAt:   time.Now().UTC(),
			Title:       item.Title,
			Description: description,
			PublishedAt: pubTime,
			Url:         item.Link,
			FeedID:      feed.ID,
		})
		if err != nil {
			if strings.Contains(err.Error(), "duplicate key") {
				continue
			}
			log.Println("failed to create post: ", err)
		}
	}

	log.Printf("feed %s collected, %v posts found", feed.Name, len(rssFeed.Channel.Item))
}

func parseDate(s string) (time.Time, error) {
	formats := []string{
		time.RFC1123Z,    // "Wed, 26 Nov 2025 10:28:03 +0000"
		time.RFC1123,     // "Wed, 26 Nov 2025 10:28:03 GMT"
		time.RFC822Z,     // "26 Nov 25 10:28 +0000"
		time.RFC822,      // "26 Nov 25 10:28 GMT"
		time.RFC3339,     // "2024-08-28T16:35:03+05:30"
		time.RFC3339Nano, // "2024-08-28T16:35:03.123456Z"
		time.DateOnly,
		"02 Jan 2006 15:04:05 -0700", // common custom format
	}

	for _, layout := range formats {
		t, err := time.Parse(layout, s)
		if err == nil {
			return t, nil
		}
	}

	return time.Time{}, fmt.Errorf("unsuported format: %s", s)
}
