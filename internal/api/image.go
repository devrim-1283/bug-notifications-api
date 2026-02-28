package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"strings"
	"time"
)

const (
	MaxImageSize = 5 * 1024 * 1024 // 5MB
)

// allowedImageTypes maps MIME types to whether they're allowed.
var allowedImageTypes = map[string]bool{
	"image/jpeg": true,
	"image/png":  true,
	"image/webp": true,
	"image/gif":  true,
}

// allowedExtensions maps lowercase extensions to whether they're allowed.
var allowedExtensions = map[string]bool{
	".jpg":  true,
	".jpeg": true,
	".png":  true,
	".webp": true,
	".gif":  true,
}

// magicBytes maps file signatures to MIME types for content-based detection.
var magicBytes = []struct {
	offset int
	magic  []byte
	mime   string
}{
	{0, []byte{0xFF, 0xD8, 0xFF}, "image/jpeg"},
	{0, []byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A}, "image/png"},
	{0, []byte("RIFF"), "image/webp"},   // RIFF....WEBP
	{0, []byte("GIF87a"), "image/gif"},
	{0, []byte("GIF89a"), "image/gif"},
}

// imageUploadResponse is the R2 API response.
type imageUploadResponse struct {
	Success bool   `json:"success"`
	URL     string `json:"url"`
}

// validateImage checks file size, extension, and magic bytes.
// Returns the validated file data and detected MIME type, or an error.
func validateImage(fh *multipart.FileHeader) ([]byte, string, error) {
	// 1. Size check
	if fh.Size > MaxImageSize {
		return nil, "", fmt.Errorf("image must be at most %dMB", MaxImageSize/(1024*1024))
	}
	if fh.Size == 0 {
		return nil, "", fmt.Errorf("image file is empty")
	}

	// 2. Extension check
	name := strings.ToLower(fh.Filename)
	dotIdx := strings.LastIndex(name, ".")
	if dotIdx == -1 {
		return nil, "", fmt.Errorf("image must have a file extension (jpg, png, webp, gif)")
	}
	ext := name[dotIdx:]
	if !allowedExtensions[ext] {
		return nil, "", fmt.Errorf("invalid image type %q, allowed: jpg, png, webp, gif", ext)
	}

	// 3. Read file content
	f, err := fh.Open()
	if err != nil {
		return nil, "", fmt.Errorf("failed to read image file")
	}
	defer f.Close()

	data, err := io.ReadAll(io.LimitReader(f, MaxImageSize+1))
	if err != nil {
		return nil, "", fmt.Errorf("failed to read image file")
	}
	if len(data) > MaxImageSize {
		return nil, "", fmt.Errorf("image must be at most %dMB", MaxImageSize/(1024*1024))
	}

	// 4. Magic bytes check - detect actual content type
	detectedMime := detectMimeFromBytes(data)
	if detectedMime == "" || !allowedImageTypes[detectedMime] {
		return nil, "", fmt.Errorf("file content is not a valid image (jpg, png, webp, gif)")
	}

	return data, detectedMime, nil
}

// detectMimeFromBytes checks file magic bytes to determine the real content type.
func detectMimeFromBytes(data []byte) string {
	for _, m := range magicBytes {
		end := m.offset + len(m.magic)
		if len(data) >= end && bytes.Equal(data[m.offset:end], m.magic) {
			// Extra check for WebP: bytes 8-12 must be "WEBP"
			if m.mime == "image/webp" && (len(data) < 12 || string(data[8:12]) != "WEBP") {
				continue
			}
			return m.mime
		}
	}
	return ""
}

// uploadToR2 sends the image to the R2 Image Processor API and returns the public URL.
func uploadToR2(apiURL, apiKey string, fileData []byte, filename string) (string, error) {
	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)

	part, err := writer.CreateFormFile("file", filename)
	if err != nil {
		return "", fmt.Errorf("create form file: %w", err)
	}
	if _, err := part.Write(fileData); err != nil {
		return "", fmt.Errorf("write file data: %w", err)
	}
	writer.Close()

	req, err := http.NewRequest(http.MethodPost, apiURL+"/upload", &buf)
	if err != nil {
		return "", fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.Header.Set("X-API-Key", apiKey)

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("r2 upload request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 1024))
		return "", fmt.Errorf("r2 upload failed with status %d: %s", resp.StatusCode, string(body))
	}

	var result imageUploadResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("decode r2 response: %w", err)
	}
	if !result.Success || result.URL == "" {
		return "", fmt.Errorf("r2 upload returned no url")
	}

	return result.URL, nil
}
