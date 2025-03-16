package imaging

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"image"
	"image/jpeg"
	"net/http"

	"golang.org/x/image/draw"
)

func GetBase64Thumb(url string) (string, error) {
	// Step 1: Fetch the image
	resp, err := http.Get(url)
	if err != nil {
		return "", fmt.Errorf("failed to download image: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("failed to download image: status code %d", resp.StatusCode)
	}

	// Step 2: Decode the image
	img, _, err := image.Decode(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to decode image: %w", err)
	}

	// Step 3: Calculate new dimensions while maintaining aspect ratio
	maxSize := 128
	bounds := img.Bounds()
	width, height := bounds.Dx(), bounds.Dy()

	if width > height {
		// Landscape → Limit width to maxSize
		if width > maxSize {
			height = (height * maxSize) / width
			width = maxSize
		}
	} else {
		// Portrait or square → Limit height to maxSize
		if height > maxSize {
			width = (width * maxSize) / height
			height = maxSize
		}
	}

	// Step 4: Resize while keeping aspect ratio
	dst := image.NewRGBA(image.Rect(0, 0, width, height))
	draw.CatmullRom.Scale(dst, dst.Bounds(), img, img.Bounds(), draw.Over, nil)

	// Step 5: Encode as JPEG in memory
	var buf bytes.Buffer
	err = jpeg.Encode(&buf, dst, &jpeg.Options{Quality: 90})
	if err != nil {
		return "", fmt.Errorf("failed to encode image to JPEG: %w", err)
	}

	// Step 6: Convert to Base64
	base64Str := base64.StdEncoding.EncodeToString(buf.Bytes())

	return base64Str, nil
}
