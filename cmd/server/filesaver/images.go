package filesaver

import (
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/google/uuid"
)

const (
	filePath    = "./static/"
	hostAddress = ""
)

func SaveImage(fileData []byte, fileExtension string) (url string, err error) {
	log.Println("Save image")

	currentTime := time.Now()
	strDate := currentTime.Format("2006-01-02")
	relativePath := strDate + "/"

	imgName := uuid.New().String()

	dir := filepath.Join(filePath, relativePath)

	err = os.MkdirAll(dir, 0755)
	if err != nil {
		log.Println("os.Mkdir Error ", err)
		return "", err
	}

	fileURL := filePath + relativePath + imgName + fileExtension

	err = ioutil.WriteFile(fileURL, fileData, 0644)
	if err != nil {
		log.Println("Error in saveImage when writing file ", err)
		return "", err
	}

	resultURL := hostAddress + relativePath + imgName + fileExtension

	return resultURL, nil
}
