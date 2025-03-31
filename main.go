// main
package main

import (
	"encoding/json"
	"log"
	"math/rand"
	"net/http"
	"os"
	"strings"
	"sync"

	"github.com/gocolly/colly"
)

// ContentPhrase is model
type ContentPhrase struct {
	Message string `json:"message,omitempty"`
	Author  string `json:"author,omitempty"`
}

// Phrases is list
type Phrases struct {
	lock   chan struct{}
	Phrase []ContentPhrase `json:"phrase,omitempty"`
	wg     *sync.WaitGroup
}

// ReponseMessage send Response in Format JSON
type ReponseMessage struct {
	Status  int    `json:"status,omitempty"`
	Message string `json:"message,omitempty"`
	Data    any    `json:"data,omitempty"`
}

func main() {
	c := colly.NewCollector()
	phrases := &Phrases{
		lock: make(chan struct{}, 1),
		wg:   &sync.WaitGroup{},
	}
	port := os.Getenv("PORT")
	if port == "" {
		port = "3000"
	}
	mux := http.NewServeMux()
	mux.HandleFunc("/phrase", phraseOfTheDay(c, phrases))
	log.Println("server run to port:", port)
	http.ListenAndServe(":"+port, AddingCors(mux))
}

// AddingCors is middleware that set access origin
func AddingCors(next http.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Content-Type", "application/json")
		log.Println(r.URL.Host)
		next.ServeHTTP(w, r)
	}
}

func phraseOfTheDay(c *colly.Collector, phrases *Phrases) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		res := &ReponseMessage{
			Status:  http.StatusOK,
			Message: "ok",
		}

		c.OnHTML("li", func(e *colly.HTMLElement) {
			phrases.wg.Add(1)
			go addPhrase(e.Text, phrases)
		})
		c.Visit("https://www.tinyrockets.app/blog/frases-motivadoras")
		phrases.wg.Wait()
		res.Data = phrases.Phrase
		//fmt.Println(phrases.phrase[randomNumber(len(phrases.phrase))])
		json.NewEncoder(w).Encode(res)
	}
}

func randomNumber(lengthPhrases int) int {
	return rand.Intn(lengthPhrases)
}

func addPhrase(text string, phrases *Phrases) {
	defer phrases.wg.Done()
	data := strings.Split(text, " - ")
	if len(data) != 2 {
		return
	}
	phrases.lock <- struct{}{}
	phrases.Phrase = append(phrases.Phrase, ContentPhrase{
		Message: data[0],
		Author:  data[1],
	})
	<-phrases.lock
}
