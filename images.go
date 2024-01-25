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
	"image"
	"image/jpeg"
	"os"

	"golang.org/x/image/draw"
)

func loadJpeg(filepath string) (image.Image, error) {
	data, err := os.ReadFile(filepath)
	if err != nil {
		return nil, err
	}
	buffer := bytes.NewBuffer(data)
	img, _, err := image.Decode(buffer)
	return img, err
}

func toJpegBytes(img image.Image) []byte {
	buffer := bytes.NewBuffer(nil)
	err := jpeg.Encode(buffer, img, nil)
	if err != nil {
		panic(err) // fails at encoding a JPEG??
	}
	return buffer.Bytes()
}

func resizeRatio(src image.Image, ratio float32) image.Image {
	newWidth := int(ratio * float32(src.Bounds().Max.X))
	newHeight := int(ratio * float32(src.Bounds().Max.Y))
	dst := image.NewRGBA(image.Rect(0, 0, newWidth, newHeight))
	draw.NearestNeighbor.Scale(dst, dst.Rect, src, src.Bounds(), draw.Over, nil)
	return dst
}

func resize(src image.Image, newWidth int) image.Image {
	ratio := float32(newWidth) / float32(src.Bounds().Max.X)
	return resizeRatio(src, ratio)
}
