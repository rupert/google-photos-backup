package main

import (
	"flag"
	"io/ioutil"
	"log"
	"os"

	"golang.org/x/oauth2/google"
	"google.golang.org/api/photoslibrary/v1"
)

func main() {
	credentialsFilename := flag.String("credentials-file", "credentials.json", "File containing Google oauth credentials (as downloaded from the Google Cloud Console)")
	tokenFilename := flag.String("token-file", "token.json", "File where the API token should be cached")
	flag.Parse()

	b, err := ioutil.ReadFile(*credentialsFilename)
	if err != nil {
		log.Fatal(err)
	}

	config, err := google.ConfigFromJSON(b, photoslibrary.PhotoslibraryReadonlyScope)
	if err != nil {
		log.Fatal(err)
	}

	client := GetClient(config, *tokenFilename)

	service, err := photoslibrary.New(client)
	if err != nil {
		log.Fatal(err)
	}

	os.MkdirAll("photos", 0755)
	os.MkdirAll("albums", 0755)

	mediaItemIds := DownloadPhotos("photos", service)
	Prune("photos", mediaItemIds)

	albumIds := DownloadAlbums("albums", service)
	Prune("albums", albumIds)
}
