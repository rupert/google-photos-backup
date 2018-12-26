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

func main() {
	usage := `Google Photos Backup

Usage:
  google-photos-backup <backup_dir>
  google-photos-backup -h | --help
  google-photos-backup --version

Options:
  -h --help                Show this screen
  --version                Show version
  --credentials-file=FILE  File containing Google OAuth credentials (as downloaded from the Google Cloud Console) [default: credentials.json]
  --token-file=FILE        File where the Google OAuth token should be cached [default: token.json]`

	arguments, _ := docopt.ParseArgs(usage, os.Args, "0.0.0")
	backupDir, _ := arguments.String("BackupDir")
	credentialsFile, _ := arguments.String("CredentialsFile")
	tokenFile, _ := arguments.String("TokenFile")

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
