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
	"bytes"
	"encoding/json"
	"fmt"
	"image"
	"io"
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
	data, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "unable to read request body", http.StatusBadRequest)
		log.Println("Receiving user picture:", err)
		return
	}

	r.Body.Close()
	log.Println("Receiving user picture of size", len(data))
	buffer := bytes.NewBuffer(data)
	img, _, err := image.Decode(buffer)
	if err != nil {
		log.Println("decoding user provided image:", err)
		http.Error(w, "we're very sorry, but we were unable to decode this image :(", http.StatusBadRequest)
		return
	}
	imgID := save(img)
	// Image ID + original width is enough info for the webpage to then call
	// /resized?imgid=abcd&pixelwidth=8
	// /guess?imgid=abcd&pixelwidth=8
	// /resized?imgid=abcd&pixelwidth=10
	// /guess?imgid=abcd&pixelwidth=10
	// etc.
	// etc.
	response := Response{
		"imageID": imgID,
		"width":   img.Bounds().Max.X,
		"height":  img.Bounds().Max.Y,
	}
	fmt.Fprint(w, response)
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
		log.Println("Resizing with ratio", ratio)
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
	if imgID := r.FormValue("imgid"); imgID != "" {
		// Referencing a picture already uploaded by the user
		log.Println("Resizing picture", imgID)
		fullImg = load(imgID)
		if fullImg == nil {
			http.Error(w, "no such image: "+imgID, http.StatusBadRequest)
			return nil
		}
	}

	if pixelWidthStr != "" {
		newWidth, err := strconv.Atoi(pixelWidthStr)
		if err != nil {
			http.Error(w, "pixelwidth must be an integer number", http.StatusBadRequest)
			return nil
		}
		ratio = float32(newWidth) / float32(fullImg.Bounds().Max.X)
		log.Println("Resizing with width", newWidth, " => ratio", ratio)
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
