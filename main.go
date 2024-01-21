package main

import (
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"sync"
)

const port = 3000

var subscriptions = []string{
	"https://www.google.com/sub",      // 备注名称
	"http://www.google.com/sub",       // 备注名称
	...
	// 添加更多子订阅链接
}

func fetchSubscriptionContent(subscription string, wg *sync.WaitGroup, ch chan<- string) {
	defer wg.Done()

	response, err := http.Get(subscription)
	if err != nil {
		fmt.Printf("Error fetching %s: %v\n", subscription, err)
		ch <- "" // Signal an empty content on error
		return
	}
	defer response.Body.Close()

	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		fmt.Printf("Error reading body for %s: %v\n", subscription, err)
		ch <- "" // Signal an empty content on error
		return
	}

	ch <- string(body)
}

func generateMergedSubscription() (string, error) {
	var wg sync.WaitGroup
	ch := make(chan string, len(subscriptions))
	contents := make([]string, 0, len(subscriptions))

	for _, subscription := range subscriptions {
		wg.Add(1)
		go fetchSubscriptionContent(subscription, &wg, ch)
		content := <-ch
		if content != "" {
			decodedContent, err := base64.StdEncoding.DecodeString(content)
			if err != nil {
				fmt.Printf("Error decoding content: %v\n", err)
				return "", err
			}
			contents = append(contents, string(decodedContent))
		}
	}

	wg.Wait()
	close(ch)

	mergedContent := base64.StdEncoding.EncodeToString([]byte(strings.Join(contents, "\n")))

	return mergedContent, nil
}

func main() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "Hello world!")
	})

	http.HandleFunc("/sub", func(w http.ResponseWriter, r *http.Request) {
		mergedSubscription, err := generateMergedSubscription()
		if err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		fmt.Fprint(w, mergedSubscription)
	})

	fmt.Printf("Server is running on port %d\n", port)
	http.ListenAndServe(fmt.Sprintf(":%d", port), nil)
}
