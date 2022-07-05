package network

import (
	"io"
	"log"
	"net/http"

	"github.com/plankton4/chat-app-server/cmd/server/filesaver"
)

const (
	maxUploadSize = 10 * 1024 * 1024 // 10MB
)

func WorkUploadImage(w http.ResponseWriter, r *http.Request) {
	log.Println("ContentLength ", r.ContentLength)
	if r.Method != http.MethodPost {
		w.Header().Set("Allow", http.MethodPost)
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	// Не больше 10мб принимаем
	if err := r.ParseMultipartForm(maxUploadSize); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	file, fileHeader, err := r.FormFile("file")
	fileHeaderBuffer := make([]byte, 512)

	// Copy the headers into the FileHeader buffer
	if _, err := file.Read(fileHeaderBuffer); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// set position back to start.
	if _, err := file.Seek(0, io.SeekStart); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	fileBuffer := make([]byte, fileHeader.Size)
	_, err = file.Read(fileBuffer)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	imageUrl, err := filesaver.SaveImage(fileBuffer, ".jpg")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	log.Println("SAVED IMAGE URL ", imageUrl)

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(imageUrl))
}
