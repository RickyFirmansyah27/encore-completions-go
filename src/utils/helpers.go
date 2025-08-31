package utils

import (
	"encoding/base64"
	"io/ioutil"
)

// EncodeImageToBase64 reads an image file and encodes it to base64
func EncodeImageToBase64(imagePath string) (string, error) {
	// Read the image file
	imageData, err := ioutil.ReadFile(imagePath)
	if err != nil {
		return "", err
	}

	// Encode to base64
	encoded := base64.StdEncoding.EncodeToString(imageData)
	return encoded, nil
}
