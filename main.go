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
	"http://node2.lunes.host:27180/sub",     // Lunes-IE-8118158
	"http://node4.lunes.host:1139/sub",      // Lunes-CA-6887668
	"http://free-2.witchly.cloud:25720/sub", // Witchly-FI-WXXUUX
	"http://51.161.130.134:10328/sub",       // Sanilds-AU-6887668
	"http://95.214.55.215:1540/sub",         // RudraCloud-PL-wxxuux
	"http://uk-bot-01.scarcehost.uk:4698/sub",                     // scarehost-GB-wxxuux
	"http://infra.chromanodes.eu:25635/sub",                       // chromanodes-CH-8118158
	"http://server.nexcord.com:10393/sub",                         // nexcord-DE-wxxuux
	"http://45.140.142.188:4246/sub",                              // solonodes-NL-6887668
	"http://myappsg.onrender.com/sub",                             // render-SG-xxuuwx@gmail
	"https://marvelous-selective-humor.glitch.me/sub",             //Glitch-US-8118158
	"https://raw.githubusercontent.com/eoovve/test/main/sub.txt",  // saclingo-wwxoot
	"https://wxxuux-testargo.hf.space/sub",                        // wxxuux-Testargo
	"http://wwxoo.serv00.net:1110/sub",                            // Serv00-xysun-xray-argo
	"http://xysun.ct8.pl:1231/sub",                                // ct8-xysun-xray-argo
	"https://raw.githubusercontent.com/eoovve/test/main/sub1.txt", // Codesphere-de+us
	// 添加更多订阅链接
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
	contents := make([]string, len(subscriptions))

	for _, subscription := range subscriptions {
		wg.Add(1)
		go fetchSubscriptionContent(subscription, &wg, ch)
	}

	go func() {
		wg.Wait()
		close(ch)
	}()

	for range subscriptions {
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

	// 重新进行base64编码
	mergedContent := base64.StdEncoding.EncodeToString([]byte(strings.Join(contents, "\n")))

	return mergedContent, nil
}

func main() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "Hello world!")
	})

	http.HandleFunc("/sum", func(w http.ResponseWriter, r *http.Request) {
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
