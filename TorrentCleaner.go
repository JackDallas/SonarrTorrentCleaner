package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"

	bolt "go.etcd.io/bbolt"
)

var (
	bucketName = []byte("torrents")
	indexBytes []byte
)

func main() {
	log.SetFormatter(&log.TextFormatter{
		FullTimestamp: true,
	})

	log.Info("Starting Sonarr Torrent Cleaner...")
	log.Info("")
	running := true
	log.Info("Loading config")
	config, err := LoadOrCreateConfig()
	if err != nil {
		log.Info("Error loading config: " + err.Error())
		panic(err)
	}
	log.Info("Config loaded")

	log.Info("Opening Database")
	db, err := bolt.Open("queue.db", 0666, nil)
	if err != nil {
		panic(err)
	}
	defer db.Close()
	log.Info("Database Opened")

	log.Info("Initialising Database")
	err = db.Update(func(tx *bolt.Tx) error {
		tx.CreateBucketIfNotExists(bucketName)
		return nil
	})
	if err != nil {
		panic(err)
	}

	//

	go SonarrCheckerLoop(&running, &config, db)
	if config.WebServer {
		RunWebServer(db, &config)
	} else {
		WaitForCtrlC()
	}
}

//https://jjasonclark.com/waiting_for_ctrl_c_in_golang/
func WaitForCtrlC() {
	var end_waiter sync.WaitGroup
	end_waiter.Add(1)
	signal_channel := make(chan os.Signal, 1)
	signal.Notify(signal_channel, os.Interrupt)
	go func() {
		<-signal_channel
		end_waiter.Done()
	}()
	end_waiter.Wait()
}

type IndexTemplates struct {
	RootPath string
}

func RunWebServer(db *bolt.DB, config *Config) {
	tmpl, err := template.ParseFiles("./static/index.html")
	if err != nil {
		log.Fatal(err)
	}

	var ibytes bytes.Buffer
	err = tmpl.Execute(&ibytes, &IndexTemplates{config.WebRoot})
	if err != nil {
		log.Fatal(err)
	}
	indexBytes = ibytes.Bytes()

	r := mux.NewRouter()
	r.HandleFunc("/api/queue", func(w http.ResponseWriter, r *http.Request) {
		var items []SonarrQueueItemDBEntry

		db.View(func(tx *bolt.Tx) error {
			// Get the Bucket
			b := tx.Bucket(bucketName)
			b.ForEach(func(k, v []byte) error {
				var item SonarrQueueItemDBEntry
				json.Unmarshal(v, &item)
				items = append(items, item)
				return nil
			})
			return nil
		})

		jsonData, err := json.Marshal(items)
		if err != nil {
			log.Error("Error marshalling json")
			log.Error(err.Error())
			fmt.Fprintf(w, "error")
		} else {
			w.Write(jsonData)
		}
	})

	spa := spaHandler{
		staticPath: "static",
		indexPath:  "index.html",
		webRoot:    config.WebRoot,
	}

	r.PathPrefix("/").Handler(spa)
	// listen to port
	srv := &http.Server{
		Handler: r,
		Addr:    fmt.Sprintf("%s:%s", config.BindIP, config.BindPort),
		// Good practice: enforce timeouts for servers you create!
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}
	srv.ListenAndServe()
}

// Shamlessly stolen from mux examples https://github.com/gorilla/mux#examples
type spaHandler struct {
	staticPath string
	indexPath  string
	webRoot    string
}

func (h spaHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// get the absolute path to prevent directory traversal
	path, err := filepath.Abs(r.URL.Path)
	if err != nil {
		// if we failed to get the absolute path respond with a 400 bad request
		// and stop
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	path = strings.Replace(path, h.webRoot, "", 1)

	// prepend the path with the path to the static directory
	path = filepath.Join(h.staticPath, path)

	// check whether a file exists at the given path
	_, err = os.Stat(path)
	if os.IsNotExist(err) || strings.HasSuffix(path, h.staticPath) {
		// file does not exist, serve index.html
		// http.ServeFile(w, r, filepath.Join(h.staticPath, h.indexPath))
		// file does not exist, serve index.html template
		w.Write(indexBytes)
		return
	} else if err != nil {
		// if we got an error (that wasn't that the file doesn't exist) stating the
		// file, return a 500 internal server error and stop
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	r.URL.Path = strings.Replace(path, h.staticPath, "", -1)
	// otherwise, use http.FileServer to serve the static dir
	http.FileServer(http.Dir(h.staticPath)).ServeHTTP(w, r)
}
