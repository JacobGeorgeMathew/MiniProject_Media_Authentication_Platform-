package utils

import (
	"image"
	"image/jpeg"
	"image/png"
	"io"
	"strings"
)

// encodeImageToWriter writes img into w using the codec inferred from mimeType.
// Falls back to PNG for any unrecognised MIME type (lossless, safe default).
// Extend the switch for TIFF / WebP once native encoders are wired in.
func EncodeImageToWriter(w io.Writer, img image.Image, mimeType string) error {
	switch {
	case strings.Contains(mimeType, "jpeg") || strings.Contains(mimeType, "jpg"):
		return jpeg.Encode(w, img, &jpeg.Options{Quality: 95})
	default:
		return png.Encode(w, img)
	}
}