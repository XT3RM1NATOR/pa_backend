package utils

import (
	"bytes"
	"errors"
	"image"
)

func ValidatePhoto(photoBytes []byte) error {
	img, _, err := image.Decode(bytes.NewReader(photoBytes))
	if err != nil {
		return err
	}

	if len(photoBytes) > 1024*1024 {
		return errors.New("photo size cannot exceed 1MB")
	}

	bounds := img.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()

	if width != height {
		return errors.New("photo must be square")
	}

	if width > 256 || height > 256 {
		return errors.New("photo dimensions cannot exceed 256x256 pixels")
	}

	return nil
}
