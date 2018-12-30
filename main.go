package main

import (
	"io/ioutil"
	"log"
	"os"
	"path"

	"github.com/docopt/docopt-go"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/photoslibrary/v1"
)

const version = "0.0.0"

func main() {
	usage := `Google Photos Backup

Usage:
  google-photos-backup [options] <backup-dir>
  google-photos-backup -h | --help
  google-photos-backup --version

Options:
  -h --help                Show this screen
  --version                Show version
  --credentials-file=FILE  File containing Google OAuth credentials (as downloaded from the Google Cloud Console) [default: credentials.json]
  --token-file=FILE        File where the user's Google OAuth token should be cached [default: token.json]`

	opts, _ := docopt.ParseArgs(usage, os.Args[1:], version)
	backupDir, _ := opts.String("<backup-dir>")
	credentialsFile, _ := opts.String("--credentials-file")
	tokenFile, _ := opts.String("--token-file")

	b, err := ioutil.ReadFile(credentialsFile)
	if err != nil {
		log.Fatal(err)
	}

	config, err := google.ConfigFromJSON(b, photoslibrary.PhotoslibraryReadonlyScope)
	if err != nil {
		log.Fatal(err)
	}

	client := GetClient(config, tokenFile)

	service, err := photoslibrary.New(client)
	if err != nil {
		log.Fatal(err)
	}

	photosDir := path.Join(backupDir, "photos")
	albumsDir := path.Join(backupDir, "albums")

	os.MkdirAll(photosDir, 0755)
	os.MkdirAll(albumsDir, 0755)

	mediaItemIds := DownloadPhotos(photosDir, service)
	Prune(photosDir, mediaItemIds)

	albumIds := DownloadAlbums(albumsDir, service)
	Prune(albumsDir, albumIds)
}
