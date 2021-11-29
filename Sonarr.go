package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"

	log "github.com/sirupsen/logrus"
)

var queueEndpoint = "/api/queue"

func (c Config) GetCurrentQueue() (SonarrQueue, error) {
	log.Info("Getting Sonarr queue from ", c.SonarrURL+queueEndpoint)
	req, err := http.NewRequest("GET", c.SonarrURL+queueEndpoint, nil)
	if err != nil {
		return SonarrQueue{}, err
	}
	req.Header.Add("X-Api-Key", c.SonarrAPIKey)

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return SonarrQueue{}, err
	}
	defer res.Body.Close()

	if res.StatusCode != 200 {
		ioBody, err := ioutil.ReadAll(res.Body)
		if err != nil {
			return SonarrQueue{}, err
		} else {
			return SonarrQueue{}, errors.New("Sonarr returned non-200 status code, Body: " + string(ioBody))
		}
	}

	log.Info("Decoding Sonarr queue response")
	var queue SonarrQueue
	err = json.NewDecoder(res.Body).Decode(&queue)
	if err != nil {
		ioBody, err := ioutil.ReadAll(res.Body)
		if err != nil {
			return SonarrQueue{}, err
		} else {
			return SonarrQueue{}, errors.New("Cannot decode response from Sonarr, Body: " + string(ioBody))
		}
	}

	//Sonarr reterns an item per episode, we want an item per torrent, filter to get out unique items by DownloadID
	var uniqueQueue = make(map[string]SonarrQueueItem)
	for _, item := range queue {
		uniqueQueue[item.DownloadID] = item
	}
	queue = nil
	//Back to an array
	for _, item := range uniqueQueue {
		queue = append(queue, item)
	}
	return queue, err
}

func (c Config) DeleteFromQueue(id int, blacklist ...bool) error {
	if len(blacklist) == 0 {
		blacklist[0] = false
	}

	SonarrQueueItemDelete := SonarrQueueItemDelete{id, blacklist[0]}
	var jsonStr, err = json.Marshal(SonarrQueueItemDelete)
	if err != nil {
		return err
	}

	client := &http.Client{}
	req, err := http.NewRequest("DELETE", c.SonarrURL+queueEndpoint, bytes.NewBuffer(jsonStr))
	if err != nil {
		return err
	}
	req.Header.Add("X-Api-Key", c.SonarrAPIKey)

	resp, err := client.Do(req)
	if err != nil {
		return err
	}

	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	if resp.StatusCode != 200 {
		return errors.New(string(respBody))
	}

	return nil
}
