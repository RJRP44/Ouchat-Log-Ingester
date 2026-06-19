package main

import (
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const uploadDir = "./uploads"

func uploadHandler(db *Database) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		// Limit total request size to 2GB max per file
		r.Body = http.MaxBytesReader(w, r.Body, 2<<30)

		if err := r.ParseMultipartForm(32 << 20); err != nil {
			http.Error(w, "File too large or invalid request: "+err.Error(), http.StatusBadRequest)
			return
		}

		file, header, err := r.FormFile("video")
		if err != nil {
			http.Error(w, "Missing 'video' field: "+err.Error(), http.StatusBadRequest)
			return
		}

		defer func(file multipart.File) {
			err := file.Close()
			if err != nil {
				log.Printf("Error closing file: %v", err)
			}
		}(file)

		ext := strings.ToLower(filepath.Ext(header.Filename))
		if ext != ".mp4" {
			http.Error(w, "Unsupported file extension: "+ext, http.StatusUnsupportedMediaType)
			return
		}

		dstPath, cat, timestamp, err := buildDestPath(header.Filename)
		if err != nil {
			http.Error(w, "Invalid filename format: "+err.Error(), http.StatusBadRequest)
			return
		}

		if err := os.MkdirAll(filepath.Dir(dstPath), 0755); err != nil {
			http.Error(w, "Server error while creating directory", http.StatusInternalServerError)
			return
		}

		dst, err := os.Create(dstPath)
		if err != nil {
			http.Error(w, "Server error while creating file", http.StatusInternalServerError)
			return
		}
		defer func(dst *os.File) {
			err := dst.Close()
			if err != nil {
				log.Fatal(err)
			}
		}(dst)

		written, err := io.Copy(dst, file)
		if err != nil {
			http.Error(w, "Error while writing file", http.StatusInternalServerError)
			return
		}

		relPath, _ := filepath.Rel(uploadDir, dstPath)
		log.Printf("File received: %s (%d bytes)", relPath, written)

		//Add to database
		err = db.InsertVideo(cat, timestamp)
		if err != nil {
			_ = fmt.Errorf("error inserting video to db: %v", err)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		_, _ = fmt.Fprintf(w, `{"status":"ok","path":"%s","size":%d}`, filepath.ToSlash(relPath), written)
	}
}

func buildDestPath(filename string) (path string, id string, t time.Time, err error) {
	filename = filepath.Base(filename) // strip any path component
	ext := filepath.Ext(filename)
	base := strings.TrimSuffix(filename, ext)

	parts := strings.SplitN(base, "_", 2)
	if len(parts) != 2 {
		return "", "", time.Time{}, fmt.Errorf("expected format id_YYYY-MM-DD_HH-MM-SS, got %q", filename)
	}
	id, timestamp := parts[0], parts[1]
	if id == "" {
		return "", "", time.Time{}, fmt.Errorf("empty id in filename %q", filename)
	}

	t, err = time.Parse("2006-01-02_15-04-05", timestamp)
	if err != nil {
		return "", "", time.Time{}, fmt.Errorf("invalid timestamp in filename %q: %w", filename, err)
	}

	dir := filepath.Join(
		uploadDir,
		id,
		fmt.Sprintf("%04d", t.Year()),
		fmt.Sprintf("%02d", t.Month()),
		fmt.Sprintf("%02d", t.Day()),
	)
	finalName := fmt.Sprintf("%02d-%02d-%02d%s", t.Hour(), t.Minute(), t.Second(), ext)

	return filepath.Join(dir, finalName), id, t, nil
}
