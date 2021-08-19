package main

import (
	"errors"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"

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

func isGoGetRequest(r *http.Request) bool {
	// Check if the parameter is in the request
	const goGetParam string = "download"
	downloadArr, ok := r.URL.Query()[goGetParam]

	// has it not been specified?
	if !ok || len(downloadArr) < 1 {
		log.WithFields(log.Fields{
			"Download":         r.URL.Query()[goGetParam],
			"Go Get Parameter": goGetParam,
		}).Debug("Is Not a Go Get Request")
		return false
	}

	// Convert the parameter to a boolean
	download, convErr := strconv.ParseBool(downloadArr[0])

	if convErr != nil {
		// Error converting to a boolean, assume false
		log.WithFields(log.Fields{
			"Download":         r.URL.Query()[goGetParam],
			"Go Get Parameter": goGetParam,
		}).Error(convErr)
		return false
	}

	// Is it set to false?
	if !download {
		log.WithFields(log.Fields{
			"Download":         r.URL.Query()[goGetParam],
			"Go Get Parameter": goGetParam,
		}).Debug("Is Not a Go Get Request")
		return false
	}

	// Must be a go get request!
	log.WithFields(log.Fields{
		"Download":         r.URL.Query()[goGetParam],
		"Go Get Parameter": goGetParam,
	}).Debug("Is a Go Get Request")
	return true
}

func redirect(w http.ResponseWriter, r *http.Request, site string) {
	log.WithFields(log.Fields{
		"To": site,
	}).Debug("Redirecting")
	http.Redirect(w, r, site, http.StatusFound)
}