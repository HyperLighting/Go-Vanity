package main

import (
	"io/ioutil"
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
