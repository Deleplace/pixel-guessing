package main

import (
	"image"
	"log"
	"math/rand"
)

type imgID string

var userImages = map[imgID]image.Image{}

// In this sample app, we don't wish to save many user images in memory.
// When maxImg will be reached, we'll remove 1 older image.
const maxImg = 8

func save(img image.Image) imgID {
	for len(userImages) >= maxImg {
		deleteOneUserImage()
	}

	id := randomImgID()
	userImages[id] = img
	return id
}

func load(id string) image.Image {
	img := userImages[imgID(id)]
	if img == nil {
		log.Println("Could not find stored image", id)
	} else {
		log.Println("Found stored image", id)
	}
	return img
}

const alphanum = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789"

func randomImgID() imgID {
	// Not crypto secure — it's fine
	const n = 4
	a := make([]byte, n)
	for i := range a {
		a[i] = alphanum[rand.Intn(len(alphanum))]
	}
	return imgID(a)
}

func deleteOneUserImage() {
	for id := range userImages {
		log.Println("Deleting stored image", id)
		delete(userImages, id)
		return
	}
}
