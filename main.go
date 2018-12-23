package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/photoslibrary/v1"
)

func main() {
	b, err := ioutil.ReadFile("credentials.json")

	if err != nil {
		log.Fatal(err)
	}

	config, err := google.ConfigFromJSON(b, photoslibrary.PhotoslibraryReadonlyScope)

	if err != nil {
		log.Fatal(err)
	}

	client := getClient(config)

	fmt.Println(client)

	photoLibraryService, err := photoslibrary.New(client)

	if err != nil {
		log.Fatal(err)
	}

	mediaItemsService := photoslibrary.NewMediaItemsService(photoLibraryService)
	call := mediaItemsService.Search(&photoslibrary.SearchMediaItemsRequest{PageSize: 100})
	response, err := call.Do()

	if err != nil {
		log.Fatal(err)
	}

	for _, mediaItem := range response.MediaItems {
		fmt.Println(mediaItem.BaseUrl)
	}
}

func getClient(config *oauth2.Config) *http.Client {
	filename := "token.json"
	token, err := readToken(filename)
	if err != nil {
		token = getTokenFromWeb(config)
		writeToken(filename, token)
	}
	return config.Client(context.Background(), token)
}

func getTokenFromWeb(config *oauth2.Config) *oauth2.Token {
	url := config.AuthCodeURL("state-token", oauth2.AccessTypeOffline)
	fmt.Printf("Go to the following link in your browser then type the "+
		"authorization code: \n%v\n", url)

	var code string
	if _, err := fmt.Scan(&code); err != nil {
		log.Fatalf("Unable to read authorization code: %v", err)
	}

	token, err := config.Exchange(context.TODO(), code)
	if err != nil {
		log.Fatalf("Unable to retrieve token from web: %v", err)
	}

	return token
}

func writeToken(path string, token *oauth2.Token) {
	f, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		log.Fatalf("Unable to cache oauth token: %v", err)
	}
	defer f.Close()
	json.NewEncoder(f).Encode(token)
}

func readToken(file string) (*oauth2.Token, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	token := &oauth2.Token{}
	err = json.NewDecoder(f).Decode(token)
	return token, err
}
