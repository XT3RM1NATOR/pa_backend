package utils

import (
	"bytes"
	"errors"
	"image"
)

func ValidateProjectId(projectId string) error {
	if len(projectId) < 6 || len(projectId) > 30 {
		return errors.New("project ID must be between 6 and 30 characters")
	}

	for _, char := range projectId {
		if !isValidCharacter(char) {
			return errors.New("project ID can only contain lowercase alphanumeric characters and hyphen (-)")
		}
	}

	return nil
}

func isValidCharacter(char rune) bool {
	return (char >= 'a' && char <= 'z') ||
		(char >= '0' && char <= '9') ||
		char == '-'
}

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
