package utils

import (
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
)

func ReadVersion() string {
	workingDir, _ := os.Getwd()
	version, err := ioutil.ReadFile(filepath.Join(workingDir, "VERSION"))
	if err != nil {
		log.Println("Warning: can't determine current version:", err)
	}

	return string(version)
}
