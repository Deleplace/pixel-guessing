// Copyright 2024 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"image"
	"log"
	"math/rand"
	"runtime"
)

type imgID string

var userImages = map[imgID]image.Image{}

// In this sample app, we don't wish to save many user images in memory.
// When maxImg will be reached, we'll remove 1 older image.
const maxImg = 20

func save(img image.Image) imgID {
	for len(userImages) >= maxImg {
		deleteOneUserImage()
	}

	id := randomImgID()
	userImages[id] = img
	log.Println("Storing image", id)
	printMemoryUsage()
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
		printMemoryUsage()
		return
	}
}

func printMemoryUsage() {
	// We're letting the GC do its work.
	// Not requesting a full GC every time.
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	log.Printf("Mem Alloc = %v MiB", m.Alloc/1024/1024)
	log.Printf("Mem TotalAlloc = %v MiB", m.TotalAlloc/1024/1024)
	log.Printf("Mem Sys = %v MiB", m.Sys/1024/1024)
	log.Printf("Mem NumGC = %v\n", m.NumGC)

}
