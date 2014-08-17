package fstopsite

import (
	"log"
)

func AssetLoader(name string) string {
	r, err := Asset(name)
	if err != nil {
		log.Printf("Could not load asset %s: %s", name, err)
		return ""
	}

	return string(r)
}
