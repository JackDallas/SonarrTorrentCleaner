package main

import (
	"encoding/json"
	"fmt"
	"time"

	bolt "go.etcd.io/bbolt"
)

var (
	bucketName = []byte("torrents")
)

func main() {
	println("Starting Sonarr Torrent Cleaner...")
	println("")
	running := true
	println("Loading config")
	config, err := LoadOrCreateConfig()
	if err != nil {
		println("Error loading config: " + err.Error())
		panic(err)
	}
	println("Config loaded")

	println("Opening Database")
	db, err := bolt.Open("queue.db", 0666, nil)
	if err != nil {
		panic(err)
	}
	defer db.Close()
	println("Database Opened")

	println("Initialising Database")
	err = db.Update(func(tx *bolt.Tx) error {
		tx.CreateBucketIfNotExists(bucketName)
		return nil
	})
	if err != nil {
		panic(err)
	}

	println("Starting app loop")
	for running {
		println("Processing Queue")
		queue, err := config.GetCurrentQueue()

		if err == nil {
			println("Looping through current queue")
			//Loop the queue items
			for _, currentQueueItem := range queue {
				//Ignore Queued Items
				if currentQueueItem.Status == "Queued" {
					continue
				}
				// Check its a torrent
				if currentQueueItem.Protocol == "torrent" {
					println("Processing Item: " + currentQueueItem.Title + " [" + currentQueueItem.DownloadID + "]")
					// Parse the itemID to a byte array
					itemID := []byte(currentQueueItem.DownloadID)

					if currentQueueItem.Sizeleft == 0 {
						println("Item is complete, removing from Sonarr")
						err = config.DeleteFromQueue(currentQueueItem.ID, false)
						if err != nil {
							println("Error removing item from Sonarr: " + err.Error())
							continue
						}
					}
					err = db.Update(func(tx *bolt.Tx) error {
						// Get the Bucket
						b := tx.Bucket(bucketName)
						// Get the item from the bucket
						prevItem := b.Get(itemID)
						// if not found, add to db
						if prevItem == nil {
							println("Item not in db, adding")
							// Create a db entry
							dbItem := SonarrQueueItemDBEntry{
								Item:        currentQueueItem,
								LastChecked: time.Now(),
							}
							// data bytes to json
							data, err := json.Marshal(dbItem)
							if err != nil {
								println("Error marshalling queue item: " + err.Error())
							} else {
								// Add to db
								tx.Bucket(bucketName).Put(itemID, data)
							}
						} else {
							// If the item is already in our db, unmarshal it
							var prevQueueItem SonarrQueueItemDBEntry
							err = json.Unmarshal(prevItem, &prevQueueItem)
							if err != nil {
								println("Error un-marshalling queue item: " + err.Error())
							} else {
								if currentQueueItem.Sizeleft == 0 {
									println("Item complete, removing from db")
									err = config.DeleteFromQueue(currentQueueItem.ID, false)
									if err != nil {
										println("Error deleting queue item: " + err.Error())
									} else {
										//Delete from db if complete
										tx.Bucket(bucketName).Delete(itemID)
									}
								} else {
									// Check if the item has progressed
									if currentQueueItem.Sizeleft > prevQueueItem.Item.Sizeleft {
										// If the item has progressed, update the db
										println("Item progress made, updating lastChecked")
										// Create a db entry
										dbItem := SonarrQueueItemDBEntry{
											Item:        currentQueueItem,
											LastChecked: time.Now(),
										}
										data, err := json.Marshal(dbItem)
										if err != nil {
											println("Error marshalling queue item: " + err.Error())
										} else {
											// Add to db
											tx.Bucket(bucketName).Put(itemID, data)
										}
									} else {
										// If the item has not progressed, check how long its been since it has
										println("No item progress made, checking time changed against timeout")
										timeSinceMinutes := time.Since(prevQueueItem.LastChecked).Minutes()
										fmt.Printf("Time since Last Progress %f minutes, Timeout is set to %f\n", timeSinceMinutes, config.NoProgressTimeoutMinutes)
										if timeSinceMinutes > config.NoProgressTimeoutMinutes {
											println("Item being timed out, removing from queue and blacklisting torrent")
											config.DeleteFromQueue(currentQueueItem.ID, true)
											tx.Bucket(bucketName).Delete(itemID)
										}
									}
								}
							}
						}
						return nil
					})
					if err != nil {
						println("Error adding to db: " + err.Error())
					}
				}
			}
		} else {
			println("Error getting current queue: " + err.Error())
		}
		println("Processing complete, sleeping for ", config.CheckTimeMinutes, " minutes")
		time.Sleep(time.Minute * time.Duration(config.CheckTimeMinutes))
	}

}
