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

// SaveBase64ToImage decodes base64 image data and saves it to a file
func SaveBase64ToImage(base64Data, outputPath string) error {
	// Decode base64 data
	decoded, err := base64.StdEncoding.DecodeString(base64Data)
	if err != nil {
		return err
	}

	// Write to file
	err = ioutil.WriteFile(outputPath, decoded, 0644)
	if err != nil {
		return err
	}

	return nil
}
