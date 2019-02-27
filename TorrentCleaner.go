package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"time"
)

var (
	config Config
)

type ConfigFile struct {
	WaitTime           string `json:"WaitTime"`
	ZeroPercentTimeout string `json:"ZeroPercentTimeout"`
	SonarrURL          string `json:"SonarrURL"`
	SonarrAPIKey       string `json:"SonarrAPIKey"`
	Blacklist          bool   `json:"Blacklist"`
}
type Config struct {
	WaitTime           time.Duration `json:"WaitTime"`
	ZeroPercentTimeout time.Duration `json:"ZeroPercentTimeout"`
	SonarrURL          string        `json:"SonarrURL"`
	SonarrAPIKey       string        `json:"SonarrAPIKey"`
	Blacklist          bool          `json:"Blacklist"`
}

func NewConfig() Config {
	WaitTime, _ := time.ParseDuration("4h00m")
	ZeroPercentTimeout, _ := time.ParseDuration("1h00m")
	SonarrURL := "http://localhost"
	SonarrAPIKey := ""
	Blacklist := true
	return Config{WaitTime, ZeroPercentTimeout, SonarrURL, SonarrAPIKey, Blacklist}
}
func NewConfigFromFile(file string) Config {
	config := NewConfig()
	var configFileStruct ConfigFile

	configFile, err := ioutil.ReadFile(file)
	if os.IsNotExist(err) {
		fmt.Println("no config found using defaults")
		return config
	} else {
		err = json.Unmarshal(configFile, &configFileStruct)
		if err != nil {
			log.Fatal(err)
		}

		if configFileStruct.WaitTime != "" {
			WaitTime, err := time.ParseDuration(configFileStruct.WaitTime)
			if err == nil {
				config.WaitTime = WaitTime
			} else {
				log.Println(err)
				log.Printf("Waittime set incorrectly in config using default, (%s) is incorrect", configFileStruct.WaitTime)
			}
		} else {
			log.Println("Waittime not set in config using default")
		}
		if configFileStruct.ZeroPercentTimeout != "" {
			ZeroPercentTimeout, err := time.ParseDuration(configFileStruct.ZeroPercentTimeout)
			if err == nil {
				config.ZeroPercentTimeout = ZeroPercentTimeout
			} else {
				log.Println(err)
				log.Printf("ZeroPercentTimeout set incorrectly in config using default, (%s) is incorrect", configFileStruct.ZeroPercentTimeout)
			}
		} else {
			log.Println("ZeroPercentTimeout not set in config using default")
		}
		if configFileStruct.SonarrURL != "" {
			config.SonarrURL = configFileStruct.SonarrURL
		} else {
			log.Println("SonarrURL not set in config using default")
		}
		if configFileStruct.SonarrAPIKey != "" {
			config.SonarrAPIKey = configFileStruct.SonarrAPIKey
		} else {
			log.Println("SonarrAPIKey not set in config using default")
		}
		config.Blacklist = configFileStruct.Blacklist
	}

	return config
}

func main() {
	f, err := os.OpenFile("TorrentCleaner.log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("error opening file: %v", err)
	}
	defer f.Close()
	log.SetOutput(f)
	log.Println("Starting Torrent Cleaner....")
	log.Printf("Time Is : %v", time.Now())
	config = NewConfigFromFile("config.json")

	file, err := ioutil.ReadFile("old_queue.json")
	if os.IsNotExist(err) {
		log.Println("No previous queue file found")
		queue, err := GetCurrentQueue()
		if err != nil {
			log.Fatalln("Error getting Sonarr Queue")
			log.Fatalln(err.Error())
			os.Exit(1)
		} else {
			log.Println("Got queue from Sonarr, saving to file...")
			queueJSON, _ := json.Marshal(queue)
			err = ioutil.WriteFile("old_queue.json", queueJSON, 0644)
			if err != nil {
				log.Fatal(err.Error())
			}
		}
	} else {
		log.Printf("Got old queue, looking for torrents over %v old", config.WaitTime)
		currentQueue, err := GetCurrentQueue()
		if err != nil {
			log.Fatal(err)
		}

		var oldQueue SonarrQueue

		err = json.Unmarshal(file, &oldQueue)
		if err != nil {
			log.Fatal(err)
		}

		currentTime := time.Now()

		for i, queueItem := range currentQueue.QueueContainers {
			err, oldQueueObject := containsID(oldQueue.QueueContainers, queueItem.Queue.ID)
			if err == nil {
				if queueItem.Queue.Status == "Downloading" {
					timeSinceLastSeen := currentTime.Sub(oldQueueObject.LastSeen)
					if timeSinceLastSeen > config.WaitTime {
						if oldQueueObject.Queue.Sizeleft == queueItem.Queue.Sizeleft {
							log.Printf("Remove %s-%v:%v", oldQueueObject.Queue.Series.Title, oldQueueObject.Queue.Episode.SeasonNumber, oldQueueObject.Queue.Episode.EpisodeNumber)
							removeFromSonarr(oldQueue, currentQueue, oldQueueObject)
							if err != nil {
								log.Fatalf(err.Error())
							}
						} else {
							//Torrent has progressed bump it's last time
							log.Printf("Progress made on %s-%v:%v bumping last time", oldQueueObject.Queue.Series.Title, oldQueueObject.Queue.Episode.SeasonNumber, oldQueueObject.Queue.Episode.EpisodeNumber)
							oldQueue.QueueContainers[i].LastSeen = currentTime
						}
					} else if timeSinceLastSeen > config.ZeroPercentTimeout && queueItem.Queue.Size == queueItem.Queue.Sizeleft {
						log.Printf("0%% progress made on %s-%v:%v in 1 hour, removing", oldQueueObject.Queue.Series.Title, oldQueueObject.Queue.Episode.SeasonNumber, oldQueueObject.Queue.Episode.EpisodeNumber)
						err = removeFromSonarr(oldQueue, currentQueue, oldQueueObject)
						if err != nil {
							log.Fatalf(err.Error())
						}
					} else {
						log.Printf("%s-%v:%v Lower than timeout, skipping", oldQueueObject.Queue.Series.Title, oldQueueObject.Queue.Episode.SeasonNumber, oldQueueObject.Queue.Episode.EpisodeNumber)
					}
				}
			} else {
				oldQueue.QueueContainers = append(oldQueue.QueueContainers, queueItem)
			}
		}
		log.Println("Queue file updated, saving to file...")
		queueJSON, _ := json.Marshal(oldQueue)
		err = ioutil.WriteFile("old_queue.json", queueJSON, 0644)
		if err != nil {
			log.Fatal(err.Error())
		}
	}
}

func removeFromSonarr(oldQueue SonarrQueue, currentQueue SonarrQueue, oldQueueObject QueueObjectContainer) error {
	url := fmt.Sprintf("%s/api/queue/%d?apikey=%s&blacklist=%t", config.SonarrURL, oldQueueObject.Queue.ID, config.SonarrAPIKey, config.Blacklist)
	req, err := http.NewRequest("DELETE", url, nil)
	// handle err
	if err != nil {
		log.Fatal(err)
		return err
	}
	resp, err := http.DefaultClient.Do(req)
	// handle err
	if err != nil {
		log.Fatal(err)
		return err
	}
	if resp.StatusCode < 300 {
		log.Printf("Removed %s-%v:%v from sonarr", oldQueueObject.Queue.Series.Title, oldQueueObject.Queue.Episode.SeasonNumber, oldQueueObject.Queue.Episode.EpisodeNumber)
		err, oldQueue.QueueContainers = removeByID(oldQueue.QueueContainers, oldQueueObject.Queue.ID)
		err, currentQueue.QueueContainers = removeByID(currentQueue.QueueContainers, oldQueueObject.Queue.ID)
	} else {
		log.Fatalf("Error removing %s-%v:%v from sonarr! %+v", oldQueueObject.Queue.Series.Title, oldQueueObject.Queue.Episode.SeasonNumber, oldQueueObject.Queue.Episode.EpisodeNumber, resp)
		return errors.New("Error removing episode from Sonarr")
	}
	return nil
}

func removeByID(list []QueueObjectContainer, ID int) (error, []QueueObjectContainer) {
	for i, a := range list {
		if a.Queue.ID == ID {
			return nil, append(list[:i], list[i+1:]...)
		}
	}
	return errors.New("Not Found"), list
}

func containsID(list []QueueObjectContainer, ID int) (error, QueueObjectContainer) {
	for _, a := range list {
		if a.Queue.ID == ID {
			return nil, a
		}
	}
	return errors.New("Not Found"), QueueObjectContainer{}
}

func GetCurrentQueue() (SonarrQueue, error) {
	url := fmt.Sprintf("%s/api/queue?apikey=%s", config.SonarrURL, config.SonarrAPIKey)
	resp, err := http.Get(url)
	if err != nil {
		log.Fatal(err)
	}
	var queue []QueueObject

	decoder := json.NewDecoder(resp.Body)

	err = decoder.Decode(&queue)
	if err != nil {
		log.Fatal(err)
		return SonarrQueue{}, err
	}

	var queueContainers []QueueObjectContainer

	currentTime := time.Now()

	for _, queueItem := range queue {
		queueContainers = append(queueContainers, QueueObjectContainer{queueItem, currentTime})
	}
	return SonarrQueue{queueContainers, currentTime}, nil
}

type SonarrQueue struct {
	QueueContainers []QueueObjectContainer `json:"QueueContainers"`
	Time            time.Time              `json:"Time"`
}

type QueueObjectContainer struct {
	Queue    QueueObject `json:"Queue"`
	LastSeen time.Time   `json:"LastSeen"`
}

//An object in the activity queue in sonarr
type QueueObject struct {
	LastCheckedTime time.Time `json:"LastCheckedTime"`
	Series          struct {
		Title       string `json:"title"`
		SortTitle   string `json:"sortTitle"`
		SeasonCount int    `json:"seasonCount"`
		Status      string `json:"status"`
		Overview    string `json:"overview"`
		Network     string `json:"network"`
		AirTime     string `json:"airTime"`
		Images      []struct {
			CoverType string `json:"coverType"`
			URL       string `json:"url"`
		} `json:"images"`
		Seasons []struct {
			SeasonNumber int  `json:"seasonNumber"`
			Monitored    bool `json:"monitored"`
		} `json:"seasons"`
		Year              int           `json:"year"`
		Path              string        `json:"path"`
		ProfileID         int           `json:"profileId"`
		LanguageProfileID int           `json:"languageProfileId"`
		SeasonFolder      bool          `json:"seasonFolder"`
		Monitored         bool          `json:"monitored"`
		UseSceneNumbering bool          `json:"useSceneNumbering"`
		Runtime           int           `json:"runtime"`
		TvdbID            int           `json:"tvdbId"`
		TvRageID          int           `json:"tvRageId"`
		TvMazeID          int           `json:"tvMazeId"`
		FirstAired        time.Time     `json:"firstAired"`
		LastInfoSync      time.Time     `json:"lastInfoSync"`
		SeriesType        string        `json:"seriesType"`
		CleanTitle        string        `json:"cleanTitle"`
		ImdbID            string        `json:"imdbId"`
		TitleSlug         string        `json:"titleSlug"`
		Certification     string        `json:"certification"`
		Genres            []string      `json:"genres"`
		Tags              []interface{} `json:"tags"`
		Added             time.Time     `json:"added"`
		Ratings           struct {
			Votes int     `json:"votes"`
			Value float64 `json:"value"`
		} `json:"ratings"`
		QualityProfileID int `json:"qualityProfileId"`
		ID               int `json:"id"`
	} `json:"series"`
	Episode struct {
		SeriesID                 int       `json:"seriesId"`
		EpisodeFileID            int       `json:"episodeFileId"`
		SeasonNumber             int       `json:"seasonNumber"`
		EpisodeNumber            int       `json:"episodeNumber"`
		Title                    string    `json:"title"`
		AirDate                  string    `json:"airDate"`
		AirDateUtc               time.Time `json:"airDateUtc"`
		Overview                 string    `json:"overview"`
		HasFile                  bool      `json:"hasFile"`
		Monitored                bool      `json:"monitored"`
		AbsoluteEpisodeNumber    int       `json:"absoluteEpisodeNumber"`
		UnverifiedSceneNumbering bool      `json:"unverifiedSceneNumbering"`
		LastSearchTime           time.Time `json:"lastSearchTime"`
		ID                       int       `json:"id"`
	} `json:"episode"`
	Quality struct {
		Quality struct {
			ID         int    `json:"id"`
			Name       string `json:"name"`
			Source     string `json:"source"`
			Resolution int    `json:"resolution"`
		} `json:"quality"`
		Revision struct {
			Version int `json:"version"`
			Real    int `json:"real"`
		} `json:"revision"`
	} `json:"quality"`
	Size                    float64       `json:"size"`
	Title                   string        `json:"title"`
	Sizeleft                float64       `json:"sizeleft"`
	Timeleft                string        `json:"timeleft"`
	EstimatedCompletionTime time.Time     `json:"estimatedCompletionTime"`
	Status                  string        `json:"status"`
	TrackedDownloadStatus   string        `json:"trackedDownloadStatus"`
	StatusMessages          []interface{} `json:"statusMessages"`
	DownloadID              string        `json:"downloadId"`
	Protocol                string        `json:"protocol"`
	ID                      int           `json:"id"`
}
