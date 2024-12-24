package httpserver

import (
	"crypto/md5"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
)

func StartServer(port string, filePath string, logConnection func(serverType, clientAddr, request string)) error {
	http.Handle("/", http.FileServer(http.Dir("./static")))

	http.HandleFunc("/upload", func(w http.ResponseWriter, r *http.Request) {
		clientIP := r.RemoteAddr
		logConnection("HTTP", clientIP, r.URL.Path)
		uploadHandler(w, r, filePath)
	})

	log.Printf("HTTP server listening on port %s...", port)
	return http.ListenAndServe(":"+port, nil)
}

func uploadHandler(w http.ResponseWriter, r *http.Request, filePath string) {
	clientIP := r.RemoteAddr
	log.Printf("Received request to /upload from %s", clientIP)

	// Validate HTTP method
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		log.Printf("Invalid request method from %s: %s", clientIP, r.Method)
		return
	}

	// Ensure the file storage directory exists
	if err := os.MkdirAll(filePath, os.ModePerm); err != nil {
		http.Error(w, "Failed to create directory", http.StatusInternalServerError)
		log.Printf("Failed to create directory %s: %v", filePath, err)
		return
	}

	// Parse and process the uploaded file
	file, fileHeader, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "Failed to read file", http.StatusBadRequest)
		log.Printf("Failed to read uploaded file from %s: %v", clientIP, err)
		return
	}
	defer file.Close()
	log.Printf("Processing uploaded file from %s: %s", clientIP, fileHeader.Filename)

	// Compute MD5 hash and save file as base64
	hash := md5.New()
	base64Buffer := base64.NewEncoder(base64.StdEncoding, hash)
	content, _ := io.ReadAll(file)
	base64Buffer.Write(content)
	base64Buffer.Close()

	md5Sum := hex.EncodeToString(hash.Sum(nil))
	outputPath := filepath.Join(filePath, md5Sum+".b64")
	err = os.WriteFile(outputPath, []byte(base64.StdEncoding.EncodeToString(content)), 0644)
	if err != nil {
		http.Error(w, "Failed to save file", http.StatusInternalServerError)
		log.Printf("Failed to save file from %s to %s: %v", clientIP, outputPath, err)
		return
	}

	log.Printf("File saved successfully from %s: %s (MD5: %s)", clientIP, outputPath, md5Sum)

	// Create response
	response := map[string]string{
		"file_id": md5Sum,
		"source_ip": clientIP,
	}
	w.Header().Set("Content-Type", "application/json")
	jsonResponse, err := json.Marshal(response)
	if err != nil {
		http.Error(w, "Failed to generate JSON response", http.StatusInternalServerError)
		log.Printf("Failed to generate JSON response for %s: %v", clientIP, err)
		return
	}

	// Send the JSON response to the client
	w.WriteHeader(http.StatusOK)
	w.Write(jsonResponse)
	log.Printf("File upload completed successfully from %s, file_id: %s", clientIP, md5Sum)
}
