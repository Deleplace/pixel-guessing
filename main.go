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
	"encoding/json"
	"fmt"
	"image"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
)

var (
	projectID = os.Getenv("GOOGLE_CLOUD_PROJECT")
	modelName = "gemini-pro-vision"
	location  = "us-central1"
)

func main() {
	log.Printf("Starting server for project %q, model %q, model location %q \n", projectID, modelName, location)

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "index.html")
	})

	http.Handle("/samples/", http.StripPrefix("/samples/", http.FileServer(http.Dir("samples"))))
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

	http.HandleFunc("/upload", userPictureUpload)
	http.HandleFunc("/resized", resized)
	http.HandleFunc("/guess", guessHandler)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
		log.Printf("Defaulting to port %s", port)
	}

	addr := os.Getenv("ADDR") + ":" + port
	log.Printf("Listening on %s\n", addr)
	err := http.ListenAndServe(addr, nil)

	log.Fatal(err)
}

func userPictureUpload(w http.ResponseWriter, r *http.Request) {
	// TODO ingest file, return list of endpoints to be called?
}

func resized(w http.ResponseWriter, r *http.Request) {
	resizedImg := extractResized(w, r)
	if resizedImg == nil {
		// Error already written to w
		return
	}

	resizedJpegData := toJpegBytes(resizedImg)
	w.Header().Set("Content-Type", "image/jpeg")
	w.Write(resizedJpegData)
}

func extractResized(w http.ResponseWriter, r *http.Request) image.Image {
	ratioStr := r.FormValue("ratio")
	pixelWidthStr := r.FormValue("pixelwidth")
	sample := r.FormValue("sample")

	if ratioStr == "" && pixelWidthStr == "" {
		http.Error(w, "ratio or pixelwidth is required", http.StatusBadRequest)
		return nil
	}
	var ratio float32
	if ratioStr != "" {
		ratio64, err := strconv.ParseFloat(ratioStr, 64)
		if err != nil {
			http.Error(w, "ratio must be a number", http.StatusBadRequest)
			return nil
		}
		ratio = float32(ratio64)
		if ratio <= 0 || ratio > 1 {
			http.Error(w, "ratio must be between 0.0 and 1.0", http.StatusBadRequest)
			return nil
		}
		log.Println("Resizing", sample, "with ratio", ratio)
	}

	var fullImg image.Image
	if sample != "" {
		// User has clicked one of the sample values
		if !strings.HasPrefix(sample, "samples/") || !strings.HasSuffix(sample, ".jpg") || strings.Contains(sample, "..") {
			http.Error(w, "invalid sample", http.StatusBadRequest)
			return nil
		}
		var err error
		fullImg, err = loadJpeg(sample)
		if err != nil {
			log.Printf("unable to open sample %q: %v", sample, err)
			http.Error(w, "unable to open sample", http.StatusInternalServerError)
			return nil
		}
	}
	if picID := r.FormValue("picID"); picID != "" {
		// Referencing a picture already uploaded by the user
		log.Println("Resizing picture", picID, "with ratio", ratio)
		// TODO
	}

	if pixelWidthStr != "" {
		newWidth, err := strconv.Atoi(pixelWidthStr)
		if err != nil {
			http.Error(w, "pixelwidth must be an integer number", http.StatusBadRequest)
			return nil
		}
		ratio = float32(newWidth) / float32(fullImg.Bounds().Max.X)
		log.Println("Resizing", sample, "with width", newWidth, " => ratio", ratio)
	}

	resizedImg := resizeRatio(fullImg, ratio)
	return resizedImg
}

func guessHandler(w http.ResponseWriter, r *http.Request) {
	resizedImg := extractResized(w, r)
	if resizedImg == nil {
		// Error already written to w
		return
	}

	resizedJpegData := toJpegBytes(resizedImg)

	ctx := r.Context()
	answer, err := guess(ctx, resizedJpegData)
	if err != nil {
		log.Printf("unable to generate contents: %v", err)
		http.Error(w, "error accessing Vertex AI", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/json")
	response := Response{
		"answer": answer,
	}
	fmt.Fprint(w, response)

}

type Response map[string]any

func (r Response) String() (s string) {
	b, err := json.MarshalIndent(r, "", "  ")
	if err != nil {
		s = ""
		return
	}
	s = string(b)
	return
}
