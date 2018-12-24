package main

import (
	"bytes"
	"fmt"
	"log"
	"mime"
	"net/http"
	"os"
	"path"

	"github.com/natefinch/atomic"
	photoslibrary "google.golang.org/api/photoslibrary/v1"
)

// DownloadPhotos downloads all of the photos and videos from the user's
// Google Photos library. It returns the list of MediaItem ids currently
// in the user's library.
func DownloadPhotos(output string, service *photoslibrary.Service) []string {
	mediaItemsService := photoslibrary.NewMediaItemsService(service)

	more := true
	pageToken := ""
	pageSize := 100
	var ids []string

	for more {
		response, err := mediaItemsService.Search(&photoslibrary.SearchMediaItemsRequest{PageSize: int64(pageSize), PageToken: pageToken}).Do()
		if err != nil {
			log.Fatal(err)
		}

		more = len(response.MediaItems) == pageSize
		pageToken = response.NextPageToken

		for _, mediaItem := range response.MediaItems {
			metadataFilename := path.Join(output, mediaItem.Id+".json")

			mediaExtension, err := getMimeTypeExtension(mediaItem.MimeType)
			if err != nil {
				log.Fatal(err)
			}
			mediaFilename := path.Join(output, mediaItem.Id+mediaExtension)

			url := mediaItem.BaseUrl + "=d"

			ids = append(ids, mediaItem.Id)

			if exists(metadataFilename) && exists(mediaFilename) {
				continue
			}

			fmt.Println(mediaItem.Id)

			json, err := mediaItem.MarshalJSON()
			if err != nil {
				log.Fatal(err)
			}

			err = atomic.WriteFile(metadataFilename, bytes.NewReader(json))
			if err != nil {
				log.Fatal(err)
			}

			err = download(mediaFilename, url)
			if err != nil {
				log.Fatal(err)
			}
		}
	}

	return ids
}

func download(filename string, url string) error {
	response, err := http.Get(url)
	if err != nil {
		return err
	}
	defer response.Body.Close()

	return atomic.WriteFile(filename, response.Body)
}

func getMimeTypeExtension(mimeType string) (string, error) {
	extensions, err := mime.ExtensionsByType(mimeType)
	if err != nil {
		return "", err
	}
	return extensions[0], nil
}

func exists(filename string) bool {
	_, err := os.Stat(filename)
	return !os.IsNotExist(err)
}
