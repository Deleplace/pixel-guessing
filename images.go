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
