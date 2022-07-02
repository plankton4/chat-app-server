package network

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"time"

	"github.com/plankton4/chat-app-server/cmd/server/config"
)

const (
	maxUploadSize = 10 * 1024 * 1024 // 10MB
	uploadURL     = "http://" + config.ImageSaverAddress + "/uploadimage"
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

	base64file := base64.StdEncoding.EncodeToString(fileBuffer)

	values := map[string]string{
		"B64":       base64file,
		"Extension": ".jpg",
	}
	jsonData, err := json.Marshal(values)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, uploadURL, bytes.NewBuffer(jsonData))
	if err != nil {
		//return nil, errors.WithMessage(err, "cannot prepare apple public keys request")
		log.Fatalln("Error when creating req ", err)
	}

	log.Println("Send request to imageSaver")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	defer func() {
		if resp.Body != nil {
			resp.Body.Close()
		}
	}()

	responseBody, err := ioutil.ReadAll(resp.Body)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	log.Println("RESP ", string(responseBody))

	w.WriteHeader(http.StatusOK)
	w.Write(responseBody)
}
