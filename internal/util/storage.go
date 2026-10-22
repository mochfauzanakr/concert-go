package util

import (
	"bytes"
	"context"
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"concert-go/internal/config"
	"concert-go/internal/exception"

	"github.com/google/uuid"
)

// StorageService abstracts asset uploads (Cloudinary or local storage)
type StorageService interface {
	Upload(ctx context.Context, fileHeader *multipart.FileHeader, folder string) (string, error)
}

type autoStorageServiceImpl struct {
	cfg *config.Config
}

// NewStorageService returns an auto-fallback storage service
func NewStorageService(cfg *config.Config) StorageService {
	if cfg.LocalStoragePath != "" {
		_ = os.MkdirAll(cfg.LocalStoragePath, 0755)
	}
	return &autoStorageServiceImpl{cfg: cfg}
}

var allowedImageExtensions = map[string]bool{
	".jpg":  true,
	".jpeg": true,
	".png":  true,
	".webp": true,
}

const maxImageUploadSize = 2 * 1024 * 1024 // 2 MB

func (s *autoStorageServiceImpl) Upload(ctx context.Context, fileHeader *multipart.FileHeader, folder string) (string, error) {
	if fileHeader == nil {
		return "", exception.BadRequest("no file provided")
	}

	if fileHeader.Size > maxImageUploadSize {
		return "", exception.BadRequest("file size exceeds maximum allowed limit of 2MB")
	}

	ext := strings.ToLower(filepath.Ext(fileHeader.Filename))
	if !allowedImageExtensions[ext] {
		return "", exception.BadRequest("invalid file format: only jpg, jpeg, png, webp are supported")
	}

	// Try Cloudinary if API key is provided
	if s.cfg.CloudinaryCloudName != "" && s.cfg.CloudinaryAPIKey != "" && s.cfg.CloudinaryAPISecret != "" {
		url, err := s.uploadToCloudinary(ctx, fileHeader, folder)
		if err == nil {
			return url, nil
		}
		log.Printf("[Storage] Cloudinary upload failed (%v), falling back to local storage", err)
	}

	return s.uploadToLocal(fileHeader, folder)
}

func (s *autoStorageServiceImpl) uploadToLocal(fileHeader *multipart.FileHeader, folder string) (string, error) {
	cleanFolder := strings.Trim(folder, "/\\")
	targetDir := filepath.Join(s.cfg.LocalStoragePath, cleanFolder)
	if err := os.MkdirAll(targetDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create upload directory: %w", err)
	}

	ext := strings.ToLower(filepath.Ext(fileHeader.Filename))
	fileName := fmt.Sprintf("%s%s", uuid.New().String(), ext)
	filePath := filepath.Join(targetDir, fileName)

	src, err := fileHeader.Open()
	if err != nil {
		return "", fmt.Errorf("failed to read uploaded file: %w", err)
	}
	defer src.Close()

	dst, err := os.Create(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to save file: %w", err)
	}
	defer dst.Close()

	if _, err := io.Copy(dst, src); err != nil {
		return "", fmt.Errorf("failed to write file content: %w", err)
	}

	return fmt.Sprintf("/uploads/%s/%s", cleanFolder, fileName), nil
}

func (s *autoStorageServiceImpl) uploadToCloudinary(ctx context.Context, fileHeader *multipart.FileHeader, folder string) (string, error) {
	src, err := fileHeader.Open()
	if err != nil {
		return "", err
	}
	defer src.Close()

	timestamp := fmt.Sprintf("%d", time.Now().Unix())
	cloudFolder := s.cfg.CloudinaryUploadFolder
	if folder != "" {
		cloudFolder = fmt.Sprintf("%s/%s", s.cfg.CloudinaryUploadFolder, strings.Trim(folder, "/\\"))
	}

	toSign := fmt.Sprintf("folder=%s&timestamp=%s%s", cloudFolder, timestamp, s.cfg.CloudinaryAPISecret)
	hash := sha1.Sum([]byte(toSign))
	signature := hex.EncodeToString(hash[:])

	var body bytes.Buffer
	writer := multipart.NewWriter(&body)

	_ = writer.WriteField("api_key", s.cfg.CloudinaryAPIKey)
	_ = writer.WriteField("timestamp", timestamp)
	_ = writer.WriteField("folder", cloudFolder)
	_ = writer.WriteField("signature", signature)

	part, err := writer.CreateFormFile("file", fileHeader.Filename)
	if err != nil {
		return "", err
	}
	if _, err := io.Copy(part, src); err != nil {
		return "", err
	}
	_ = writer.Close()

	apiURL := fmt.Sprintf("https://api.cloudinary.com/v1_1/%s/image/upload", s.cfg.CloudinaryCloudName)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, apiURL, &body)
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())

	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		respBody, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("cloudinary error %d: %s", resp.StatusCode, string(respBody))
	}

	var res struct {
		SecureURL string `json:"secure_url"`
		URL       string `json:"url"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&res); err != nil {
		return "", err
	}

	if res.SecureURL != "" {
		return res.SecureURL, nil
	}
	return res.URL, nil
}
