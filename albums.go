package main

import (
	"bytes"
	"encoding/json"
	"log"
	"path"

	"github.com/natefinch/atomic"
	photoslibrary "google.golang.org/api/photoslibrary/v1"
)

// DownloadAlbums downloads the metadata for all of the albums in the user's
// Google Photos library. Note: it doesn't download the actual media files.
// It returns a list of Album ids in the user's library.
func DownloadAlbums(output string, service *photoslibrary.Service) []string {
	albumsService := photoslibrary.NewAlbumsService(service)
	mediaItemsService := photoslibrary.NewMediaItemsService(service)

	response, err := albumsService.List().Do()
	if err != nil {
		log.Fatal(err)
	}

	var ids []string

	for _, album := range response.Albums {
		ids = append(ids, album.Id)
		mediaItemIds := getAlbumMediaItemIds(mediaItemsService, album.Id)
		album := myAlbum{Id: album.Id, Title: album.Title, ProductUrl: album.ProductUrl, MediaItemIds: mediaItemIds}
		json, err := json.Marshal(album)
		if err != nil {
			log.Fatal(err)
		}
		filename := path.Join(output, album.Id+".json")
		atomic.WriteFile(filename, bytes.NewReader(json))
	}

	return ids
}

func getAlbumMediaItemIds(service *photoslibrary.MediaItemsService, id string) []string {
	more := true
	pageToken := ""
	pageSize := 100
	var ids []string

	for more {
		response, err := service.Search(&photoslibrary.SearchMediaItemsRequest{AlbumId: id, PageSize: int64(pageSize), PageToken: pageToken}).Do()
		if err != nil {
			log.Fatal(err)
		}

		more = len(response.MediaItems) == pageSize
		pageToken = response.NextPageToken

		for _, mediaItem := range response.MediaItems {
			ids = append(ids, mediaItem.Id)
		}
	}

	return ids
}

// TODO reuse photolibrary.Album
type myAlbum struct {
	Id           string   `json:"id,omitempty"`
	ProductUrl   string   `json:"productUrl,omitempty"`
	Title        string   `json:"title,omitempty"`
	MediaItemIds []string `json:"mediaItemIds,omitempty"`
}
