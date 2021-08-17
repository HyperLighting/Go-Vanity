package main

import (
	"errors"
	"io/ioutil"
	"net/http"
	"os"

	log "github.com/sirupsen/logrus"
)

func readLocalFile(fileName string) (out []byte, err error) {
	// Open the File
	file, errorOpen := os.Open(fileName)

	// Error openning file?
	if errorOpen != nil {
		log.WithFields(log.Fields{
			"File": fileName,
		}).Error(errorOpen)
		return nil, errorOpen
	}

	// Convert to Byte Array
	byteValue, errorConvert := ioutil.ReadAll(file)

	// Error Converting?
	if errorConvert != nil {
		log.WithFields(log.Fields{
			"File": fileName,
		}).Error(errorConvert)
		return nil, errorConvert
	}

	return byteValue, nil
}

func readRemoteFile(source string) (out []byte, err error) {
	// Get response from the URL
	resp, err := http.Get(source)

	if err != nil {
		return out, err
	}

	// Defer Closing
	defer resp.Body.Close()

	// Check the status code
	if resp.StatusCode != http.StatusOK {
		return out, errors.New("status code not ok")
	}

	// Convert response to a byte array
	byteValue, errorConvert := ioutil.ReadAll(resp.Body)

	if errorConvert != nil {
		log.WithFields(log.Fields{
			"Source": source,
		}).Error(errorConvert)
		return nil, errorConvert
	}

	return byteValue, nil
}
