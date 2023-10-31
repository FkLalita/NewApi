package main

import (
	"encoding/json"
	"html/template"
	"log"
	"net/http"
	"sync"
	"time"
)

var (
	topHeadlines []NewsArticle
	mu           sync.RWMutex
)

type NewsArticle struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	URL         string `json:"url"`
	apiKey      string   //your api key
}

func fetchTopHeadlines() {
	for {
		// Define the API endpoint for top headlines
		topHeadlinesURL := "https://newsapi.org/v2/top-headlines?sources=techcrunch&apiKey=" + apiKey   
		response, err := http.Get(topHeadlinesURL)
		if err != nil {
			log.Printf("Failed to fetch top headlines: %v", err)
			return
		}

		if response.StatusCode != http.StatusOK {
			response.Body.Close() // Close the response body before returning
			log.Printf("HTTP request to top headlines failed with status code: %d", response.StatusCode)
			return
		}

		var data struct {
			Articles []NewsArticle `json:"articles"`
		}
		err = json.NewDecoder(response.Body).Decode(&data)
		response.Body.Close() // Close the response body

		if err != nil {
			log.Printf("Failed to decode JSON response: %v", err)
			return
		}

		// Update the top headlines with the fetched data
		mu.Lock()
		topHeadlines = data.Articles
		mu.Unlock()

		// Wait for a specified interval before updating again
		time.Sleep(5 * time.Minute)
	}
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
	mu.RLock()
	defer mu.RUnlock()

	// Render the top headlines in your HTML template
	tmpl, err := template.ParseFiles("static/index.html")
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		log.Printf("Failed to parse template: %v", err)
		return
	}

	data := struct {
		TopHeadlines []NewsArticle
	}{
		TopHeadlines: topHeadlines,
	}

	if err := tmpl.Execute(w, data); err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		log.Printf("Failed to execute template: %v", err)
	}
}

func main() {
	go fetchTopHeadlines()

	// Serve HTTP requests
	http.HandleFunc("/", indexHandler)
	log.Println("Starting server...")

	err := http.ListenAndServe(":8000", nil)
	if err != nil {
		log.Fatal(err)
	}
}
