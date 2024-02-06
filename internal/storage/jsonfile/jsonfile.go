package jsonfile

import (
	"encoding/json"
	"github.com/terratensor/kremlin-parser/internal/parser"
	"log"
	"os"
)

func checkError(message string, err error) {
	if err != nil {
		log.Fatal(message, err)
	}
}

func WriteJsonFile(entries parser.Entries, outputPath string) {

	// Create file
	file, err := os.Create(outputPath)
	checkError("Cannot create file", err)
	defer file.Close()

	aJson, err := json.MarshalIndent(entries, "", "\t")
	if err != nil {
		log.Fatal(err)
	}
	_, err = file.Write(aJson)
	checkError("Cannot write to the file", err)
}
