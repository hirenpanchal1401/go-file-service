package service

import (
	"bytes"
	"fmt"
	"image/jpeg"
	"image/png"
	"log"
	"strings"

	"github.com/disintegration/imaging"
)

type ImageProcessor struct{}

func NewImageProcessor() *ImageProcessor {
	return &ImageProcessor{}
}

// CompressImage compresses an image with specified quality
func (ip *ImageProcessor) CompressImage(fileData []byte, mimeType string, quality int) ([]byte, error) {
	// Clamp quality between 60 and 95
	if quality < 60 {
		quality = 60
	}
	if quality > 95 {
		quality = 95
	}

	// Decode image
	img, err := imaging.Decode(bytes.NewReader(fileData))
	if err != nil {
		return nil, fmt.Errorf("failed to decode image: %w", err)
	}

	log.Printf("Image quality: %d", quality)

	// Encode back with compression based on MIME type
	var buf bytes.Buffer
	switch mimeType {
	case "image/jpeg", "image/jpg":
		err = jpeg.Encode(&buf, img, &jpeg.Options{Quality: quality})
	case "image/png":
		encoder := &png.Encoder{
			CompressionLevel: png.DefaultCompression,
		}
		err = encoder.Encode(&buf, img)
	case "image/gif":
		// GIF doesn't support quality settings, return as-is
		buf.Write(fileData)
	default:
		// Unknown format, try JPEG
		err = jpeg.Encode(&buf, img, &jpeg.Options{Quality: quality})
	}

	if err != nil {
		return nil, fmt.Errorf("failed to encode image: %w", err)
	}

	return buf.Bytes(), nil
}

// CompressImageProgressive compresses image progressively until target size is reached
func (ip *ImageProcessor) CompressImageProgressive(fileData []byte, mimeType string, targetSize int64) ([]byte, error) {
	quality := 85
	var result []byte
	var err error

	for quality >= 60 {
		result, err = ip.CompressImage(fileData, mimeType, quality)
		if err != nil {
			return nil, err
		}

		if int64(len(result)) <= targetSize {
			log.Printf("Compression successful at quality %d: %d -> %d bytes", quality, len(fileData), len(result))
			return result, nil
		}

		quality -= 5
	}

	// If we've reached here, return best effort (quality 60)
	return result, nil
}

// GetImageInfo returns information about an image
func (ip *ImageProcessor) GetImageInfo(fileData []byte) (map[string]interface{}, error) {
	img, err := imaging.Decode(bytes.NewReader(fileData))
	if err != nil {
		return nil, fmt.Errorf("failed to decode image: %w", err)
	}

	bounds := img.Bounds()
	width := bounds.Max.X - bounds.Min.X
	height := bounds.Max.Y - bounds.Min.Y

	return map[string]interface{}{
		"width":  width,
		"height": height,
		"size":   len(fileData),
	}, nil
}

// NormalizeImageMimeType converts common MIME types to standard format
func NormalizeImageMimeType(mimeType string) string {
	if strings.Contains(mimeType, "jpeg") || strings.Contains(mimeType, "jpg") {
		return "image/jpeg"
	}
	if strings.Contains(mimeType, "png") {
		return "image/png"
	}
	if strings.Contains(mimeType, "gif") {
		return "image/gif"
	}
	if strings.Contains(mimeType, "webp") {
		return "image/webp"
	}
	if strings.Contains(mimeType, "bmp") {
		return "image/bmp"
	}
	if strings.Contains(mimeType, "tiff") {
		return "image/tiff"
	}
	return mimeType
}
