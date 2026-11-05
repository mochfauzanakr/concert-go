package util_test

import (
	"bytes"
	"context"
	"mime/multipart"
	"net/textproto"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"concert-go/internal/config"
	"concert-go/internal/util"
)

func TestStorageService_LocalUpload(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "storage_test_*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	cfg := &config.Config{
		LocalStoragePath: tempDir,
	}
	storage := util.NewStorageService(cfg)
	ctx := context.Background()

	t.Run("rejects nil file", func(t *testing.T) {
		_, err := storage.Upload(ctx, nil, "avatars")
		if err == nil {
			t.Fatal("expected error on nil file header")
		}
	})

	t.Run("rejects unsupported extension", func(t *testing.T) {
		header := &multipart.FileHeader{
			Filename: "exploit.exe",
			Size:     100,
		}
		_, err := storage.Upload(ctx, header, "avatars")
		if err == nil {
			t.Fatal("expected error on unsupported extension")
		}
	})

	t.Run("rejects file larger than 2MB", func(t *testing.T) {
		header := &multipart.FileHeader{
			Filename: "large.png",
			Size:     3 * 1024 * 1024,
		}
		_, err := storage.Upload(ctx, header, "avatars")
		if err == nil {
			t.Fatal("expected error on file size > 2MB")
		}
	})

	t.Run("success local upload png", func(t *testing.T) {
		content := []byte("fake image content")
		var body bytes.Buffer
		writer := multipart.NewWriter(&body)

		h := make(textproto.MIMEHeader)
		h.Set("Content-Disposition", `form-data; name="file"; filename="test.png"`)
		h.Set("Content-Type", "image/png")
		part, _ := writer.CreatePart(h)
		_, _ = part.Write(content)
		_ = writer.Close()

		reader := multipart.NewReader(&body, writer.Boundary())
		form, err := reader.ReadForm(1024 * 1024)
		if err != nil {
			t.Fatalf("failed to read multipart form: %v", err)
		}
		fileHeader := form.File["file"][0]

		url, err := storage.Upload(ctx, fileHeader, "avatars")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if !strings.HasPrefix(url, "/uploads/avatars/") {
			t.Errorf("expected URL to start with /uploads/avatars/, got %s", url)
		}

		// Verify file actually exists on disk in tempDir
		savedPath := filepath.Join(tempDir, "avatars", filepath.Base(url))
		if _, err := os.Stat(savedPath); os.IsNotExist(err) {
			t.Errorf("expected saved file at %s, but file was not found", savedPath)
		}
	})
}
