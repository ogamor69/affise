package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"sync"
	"time"
)

type requestPayload struct {
	URLs []string `json:"urls"`
}

type responsePayload struct {
	Data map[string]string `json:"data"`
}

// Handler обрабатывает HTTP-запросы, получает список URL-адресов и возвращает данные,
// полученные от каждого из них.
func Handler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Only POST method is allowed", http.StatusMethodNotAllowed)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	var req requestPayload
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON payload", http.StatusBadRequest)
		return
	}

	if len(req.URLs) > 20 {
		http.Error(w, "Request limit exceeded: maximum 20 URLs allowed", http.StatusBadRequest)
		return
	}

	data := processURLs(ctx, cancel, req.URLs)

	resp := responsePayload{Data: data}
	jsonData, err := json.Marshal(resp)
	if err != nil {
		http.Error(w, "Failed to generate JSON response", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(jsonData)
}

// processURLs обрабатывает список URL-адресов, выполняет запросы и возвращает данные
// в виде отображения URL-адресов на полученные данные.
func processURLs(ctx context.Context, cancel context.CancelFunc, urls []string) map[string]string {
	var wg sync.WaitGroup
	data := make(map[string]string)
	mu := &sync.Mutex{}
	sem := make(chan struct{}, 4)

	for _, url := range urls {
		wg.Add(1)
		sem <- struct{}{}

		go func(url string) {
			defer func() {
				<-sem
				if r := recover(); r != nil {
					fmt.Printf("WaitGroup error: %v\n", r)
				}
				wg.Done()
			}()

			ch := make(chan string, 1)
			go fetchURL(ctx, url, ch)

			select {
			case <-ctx.Done():
				return
			case body := <-ch:
				if body == "" {
					cancel()
					return
				}
				mu.Lock()
				data[url] = body
				mu.Unlock()
			}
		}(url)
	}

	wg.Wait()
	return data
}

// fetchURL выполняет HTTP-запрос к указанному URL-адресу и возвращает полученные данные
// в канале ch.
func fetchURL(ctx context.Context, url string, ch chan<- string) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		ch <- ""
		return
	}

	client := &http.Client{Timeout: 1 * time.Second}
	req = req.WithContext(ctx)

	resp, err := client.Do(req)
	if err != nil {
		ch <- ""
		return
	}

	defer func() {
		closeErr := resp.Body.Close
		if closeErr != nil {
			log.Printf("Error closing response body: %v", closeErr())
		}
	}()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		ch <- ""
		return
	}

	ch <- string(body)
}
