package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/spf13/viper"
)

type Configuration struct {
	API_Token string
}

var Config Configuration

const groupId int = -218375169

func apiGetWall(count, offset int) (map[string]interface{}, error) {
	/* VK API wall.get call function. Returns json object, contains last { offset } to { offset + offset } posts */
	urlGet := "https://api.vk.com/method/wall.get?"
	urlParameters := url.Values{}

	headers := map[string]string{
		"Authorization": fmt.Sprintf("Bearer %s", Config.API_Token),
	}

	parameters := map[string]string{
		"owner_id": fmt.Sprint(groupId),
		"count":    fmt.Sprint(count),
		"offset":   fmt.Sprint(offset),
		"v":        "5.199",
	}

	for key, value := range parameters {
		urlParameters.Add(key, value)
	}

	urlGet = urlGet + urlParameters.Encode()

	request, err := http.NewRequest("GET", urlGet, nil)
	if err != nil {
		return nil, err
	}

	for key, value := range headers {
		request.Header.Add(key, value)
	}

	client := &http.Client{}
	response, err := client.Do(request)
	if err != nil {
		return nil, err
	}

	defer response.Body.Close()

	groupWallRaw, _ := io.ReadAll(response.Body)
	groupWall := make(map[string]interface{})

	json.Unmarshal(groupWallRaw, &groupWall)

	return groupWall, err
}

func apiGetComms(postId int) (map[string]interface{}, error) {
	/* VK API wall.getComments call function. Returns json object, contains comments from specific post set by postId */
	urlGet := "https://api.vk.com/method/wall.getComments?"
	urlParameters := url.Values{}

	headers := map[string]string{
		"Authorization": fmt.Sprintf("Bearer %s", Config.API_Token),
	}

	parameters := map[string]string{
		"owner_id": fmt.Sprint(groupId),
		"post_id":  fmt.Sprint(postId),
		"v":        "5.199",
	}

	for key, value := range parameters {
		urlParameters.Add(key, value)
	}

	urlGet = urlGet + urlParameters.Encode()

	request, err := http.NewRequest("GET", urlGet, nil)
	if err != nil {
		return nil, err
	}

	for key, value := range headers {
		request.Header.Add(key, value)
	}

	client := &http.Client{}
	response, err := client.Do(request)
	if err != nil {
		return nil, err
	}

	defer response.Body.Close()

	commsRaw, _ := io.ReadAll(response.Body)
	comms := make(map[string]interface{})

	json.Unmarshal(commsRaw, &comms)

	return comms, err
}

func main() {
	mainTime := time.Now()

	viper.AddConfigPath("./")
	viper.SetConfigType("yml")

	viper.ReadInConfig()
	viper.Unmarshal(&Config)

	posts, _ := apiGetWall(1, 0)
	postsCount := posts["response"].(map[string]interface{})["count"]

	envelopes := 0
	offset := 0

	var wg sync.WaitGroup
	var innerWg sync.WaitGroup

	for offset < int(postsCount.(float64)) {
		wg.Add(1)
		/* goroutine for posts */
		go func(offset int) {
			defer wg.Done()

			groupWall, _ := apiGetWall(100, offset)
			for _, item := range groupWall["response"].(map[string]interface{})["items"].([]interface{}) {
				innerWg.Add(1)
				/* goroutine for comments */
				go func() {
					defer innerWg.Done()

					if int(item.(map[string]interface{})["comments"].(map[string]interface{})["count"].(float64)) > 0 {
						comms, _ := apiGetComms(int(item.(map[string]interface{})["id"].(float64)))
						for _, comm := range comms["response"].(map[string]interface{})["items"].([]interface{}) {
							envelopes += strings.Count(strings.ToLower(comm.(map[string]interface{})["text"].(string)), "энвилоуп")
						}
					}
				}()
			}

			innerWg.Wait()
		}(offset)

		offset += 100
	}

	wg.Wait()

	fmt.Printf("Энвилоупс: %d\n", envelopes)
	fmt.Printf("Время выполнения: %s", time.Since(mainTime))
}
