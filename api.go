package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	"golang.org/x/oauth2"
)

// GetClient returns an authenticated Google API client.
func GetClient(config *oauth2.Config) *http.Client {
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
