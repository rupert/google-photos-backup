package main

import (
	"io/ioutil"
	"log"
	"os"
	"path"
	"strings"
)

// Prune removes deleted photos and videos from the backup directory.
func Prune(output string, ids []string) {
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

func toSet(items []string) map[string]bool {
	set := make(map[string]bool, len(items))

	for _, item := range items {
		set[item] = true
	}

	return set
}
