package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"sync"
	"time"
)

type Event struct {
	ID        int       `json:"id"`
	Type      string    `json:"type"`
	To        string    `json:"to,omitempty"` 
	Timestamp time.Time `json:"timestamp"`
}

var (
	events []Event
	mu     sync.Mutex
	count  int
)

func recordHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	var e struct {
		Type string `json:"type"`
		To   string `json:"to,omitempty"`
	}
	if err := json.NewDecoder(r.Body).Decode(&e); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	mu.Lock()
	count++
	event := Event{
		ID:        count,
		Type:      e.Type,
		To:        e.To,
		Timestamp: time.Now(),
	}
	events = append(events, event)
	mu.Unlock()

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(event); err != nil {
		log.Println("Failed to encode event:", err)
	}
}

func listEventsHandler(w http.ResponseWriter, r *http.Request) {
	mu.Lock()
	defer mu.Unlock()
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(events); err != nil {
		log.Println("Failed to encode events:", err)
	}
}

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/record", recordHandler)
	mux.HandleFunc("/events", listEventsHandler)

	port := os.Getenv("PORT")
	if port == "" {
		port = "9091" // Port for Analytics
	}

	log.Println("Analytics Service running on port", port)
	log.Fatal(http.ListenAndServe(":"+port, mux))
}
