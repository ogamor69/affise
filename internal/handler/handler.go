package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
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

	var wg sync.WaitGroup
	timeoutCtx, timeoutCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer timeoutCancel()

	if waitErr := waitGroupWithContext(timeoutCtx, &wg); waitErr != nil {
		http.Error(w, "Some URLs failed to fetch", http.StatusInternalServerError)
		return
	}

	resp := responsePayload{Data: data}
	jsonData, err := json.Marshal(resp)
	if err != nil {
		http.Error(w, "Failed to generate JSON response", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(jsonData)
}

// Func for processing a list of URLs
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

// Func for performing HTTP requests to URL
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
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		ch <- ""
		return
	}

	ch <- string(body)
}

// Func for waiting for all operations in the group to complete
func waitGroupWithContext(ctx context.Context, wg *sync.WaitGroup) error {
	done := make(chan struct{})
	go func() {
		defer close(done)
		wg.Wait()
	}()

	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-done:
		return nil
	}
}

// Path: internal/handler/handler_test.go
