package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"mime"
	"net/http"
	"os"
	"path"
	"strings"

	"github.com/natefinch/atomic"
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

	service, err := photoslibrary.New(client)
	if err != nil {
		log.Fatal(err)
	}

	os.MkdirAll("photos", 0755)
	os.MkdirAll("albums", 0755)

	mediaItemIds := downloadPhotos("photos", service)
	prune("photos", mediaItemIds)

	albumIds := downloadAlbums("albums", service)
	prune("albums", albumIds)
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

func exists(filename string) bool {
	_, err := os.Stat(filename)
	return !os.IsNotExist(err)
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

func downloadPhotos(output string, service *photoslibrary.Service) []string {
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

func toSet(items []string) map[string]bool {
	set := make(map[string]bool, len(items))

	for _, item := range items {
		set[item] = true
	}

	return set
}

func prune(output string, ids []string) {
	files, err := ioutil.ReadDir(output)
	if err != nil {
		log.Fatal(err)
	}

	idSet := toSet(ids)

	for _, file := range files {
		if file.IsDir() {
			continue
		}

		id := strings.TrimSuffix(file.Name(), path.Ext(file.Name()))

		if !idSet[id] {
			filename := path.Join(output, file.Name())

			err := os.Remove(filename)
			if err != nil {
				log.Fatal(err)
			}
		}
	}
}

// TODO reuse photolibrary.Album
type myAlbum struct {
	Id           string   `json:"id,omitempty"`
	ProductUrl   string   `json:"productUrl,omitempty"`
	Title        string   `json:"title,omitempty"`
	MediaItemIds []string `json:"mediaItemIds,omitempty"`
}

func downloadAlbums(output string, service *photoslibrary.Service) []string {
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
