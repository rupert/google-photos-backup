package main

import (
	"io/ioutil"
	"log"
	"os"

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

	client := GetClient(config)

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
