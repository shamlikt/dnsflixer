package httpserver

import (
	"crypto/md5"
	"encoding/base64"
	"encoding/hex"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"encoding/json"
)

func StartServer(port string, filePath string) error {
	http.Handle("/", http.FileServer(http.Dir("./static")))

	
	http.HandleFunc("/upload", func(w http.ResponseWriter, r *http.Request) {
		uploadHandler(w, r, filePath)
	})
	return http.ListenAndServe(":"+port, nil)
}


func uploadHandler(w http.ResponseWriter, r *http.Request, filePath string) {
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	if err := os.MkdirAll(filePath, os.ModePerm); err != nil {
		http.Error(w, "Failed to create directory", http.StatusInternalServerError)
		return
	}

	file, _, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "Failed to read file", http.StatusBadRequest)
		return
	}
	defer file.Close()

	hash := md5.New()
	base64Buffer := base64.NewEncoder(base64.StdEncoding, hash)
	content, _ := io.ReadAll(file)
	base64Buffer.Write(content)
	base64Buffer.Close()

	md5Sum := hex.EncodeToString(hash.Sum(nil))
	outputPath := filepath.Join(filePath, md5Sum+".b64")
	os.WriteFile(outputPath, []byte(base64.StdEncoding.EncodeToString(content)), 0644)


	response := map[string]string{
		"file_id": md5Sum,
	}

	w.Header().Set("Content-Type", "application/json")
	jsonResponse, err := json.Marshal(response)
	if err != nil {
		http.Error(w, "Failed to generate JSON response", http.StatusInternalServerError)
		return
	}

	// Send the JSON response to the client
	w.WriteHeader(http.StatusOK)
	w.Write(jsonResponse)
	
}
