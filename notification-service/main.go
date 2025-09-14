package main

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"sync"
	"time"
)

type Notification struct {
	ID        int       `json:"id"`
	To        string    `json:"to"`
	Message   string    `json:"message"`
	Timestamp time.Time `json:"timestamp"`
}

var (
	notifications []Notification
	mu            sync.Mutex
	counter       int
)

func notifyHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		To      string `json:"to"`
		Message string `json:"message"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	mu.Lock()
	counter++
	notification := Notification{
		ID:        counter,
		To:        req.To,
		Message:   req.Message,
		Timestamp: time.Now(),
	}
	notifications = append(notifications, notification)
	mu.Unlock()

	// Send event to analytics service asynchronously
	go func(to string) {
		event := map[string]string{
			"type": "notification_sent",
			"to":   to,
		}
		data, err := json.Marshal(event)
		if err != nil {
			log.Println("Failed to marshal event:", err)
			return
		}

		resp, err := http.Post(
			"http://localhost:9091/record",
			"application/json",
			bytes.NewBuffer(data),
		)
		if err != nil {
			log.Println("Failed to send event to analytics:", err)
			return
		}
		defer resp.Body.Close()
	}(req.To)

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(notification); err != nil {
		log.Println("Failed to encode response:", err)
	}
}

func listHandler(w http.ResponseWriter, r *http.Request) {
	mu.Lock()
	defer mu.Unlock()

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(notifications); err != nil {
		log.Println("Failed to encode response:", err)
	}
}

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/notify", notifyHandler)
	mux.HandleFunc("/notifications", listHandler)

	port := os.Getenv("PORT")
	if port == "" {
		port = "9090"
	}

	log.Println("Notification Service running on port", port)
	log.Fatal(http.ListenAndServe(":"+port, mux))
}
