package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
)

type ObjectData struct {
	//  arbitrary data of key value pairs
	Data map[string]string `json:"data"`
}

var emptyData = `{"data": {}}`

// File locking to prevent concurrent access issues
var fileLocks = make(map[string]*sync.Mutex)
var locksMutex sync.Mutex

// Validation patterns
var validNamePattern = regexp.MustCompile(`^[a-zA-Z0-9_-]+$`)

// validateInput validates object, table, and key names
func validateInput(object, table, key string) error {
	if object == "" || table == "" {
		return fmt.Errorf("object and table cannot be empty")
	}

	if !validNamePattern.MatchString(object) {
		return fmt.Errorf("invalid object name: %s", object)
	}

	if !validNamePattern.MatchString(table) {
		return fmt.Errorf("invalid table name: %s", table)
	}

	if key != "" && !validNamePattern.MatchString(key) {
		return fmt.Errorf("invalid key name: %s", key)
	}

	return nil
}

// getFileLock returns a mutex for the given file path
func getFileLock(filePath string) *sync.Mutex {
	locksMutex.Lock()
	defer locksMutex.Unlock()

	if fileLocks[filePath] == nil {
		fileLocks[filePath] = &sync.Mutex{}
	}
	return fileLocks[filePath]
}

// sanitizePath prevents directory traversal attacks
func sanitizePath(userhash, object, table string) (string, error) {
	// Remove any path traversal attempts
	userhash = strings.ReplaceAll(userhash, "..", "")
	userhash = strings.ReplaceAll(userhash, "/", "")
	userhash = strings.ReplaceAll(userhash, "\\", "")

	object = strings.ReplaceAll(object, "..", "")
	object = strings.ReplaceAll(object, "/", "")
	object = strings.ReplaceAll(object, "\\", "")

	table = strings.ReplaceAll(table, "..", "")
	table = strings.ReplaceAll(table, "/", "")
	table = strings.ReplaceAll(table, "\\", "")

	if userhash == "" {
		userhash = "default"
	}

	return filepath.Join("./data", userhash, object, table+".json"), nil
}

func openData(userhash, object, table string) (ObjectData, error) {
	log.Println("openData", userhash, object, table)

	// Validate input
	if err := validateInput(object, table, ""); err != nil {
		return ObjectData{}, err
	}

	// Get safe file path
	tableFile, err := sanitizePath(userhash, object, table)
	if err != nil {
		return ObjectData{}, err
	}

	// Get file lock to prevent concurrent access
	fileLock := getFileLock(tableFile)
	fileLock.Lock()
	defer fileLock.Unlock()

	// if the file and parent folders do not exist, create them
	if _, err := os.Stat(tableFile); os.IsNotExist(err) {
		if err := os.MkdirAll(filepath.Dir(tableFile), 0755); err != nil {
			return ObjectData{}, fmt.Errorf("failed to create directory: %w", err)
		}
		if err := os.WriteFile(tableFile, []byte(emptyData), 0644); err != nil {
			return ObjectData{}, fmt.Errorf("failed to create file: %w", err)
		}
	}

	// read the data from the file and unmarshal it into the object data
	data, err := os.ReadFile(tableFile)
	if err != nil {
		return ObjectData{}, fmt.Errorf("failed to read file: %w", err)
	}

	parsedData := ObjectData{}
	if err := json.Unmarshal(data, &parsedData); err != nil {
		return ObjectData{}, fmt.Errorf("failed to parse JSON: %w", err)
	}

	return parsedData, nil
}
func getData(userhash, object, table, key string) (string, error) {
	log.Println("getData", userhash, object, table, key)

	// Validate input
	if err := validateInput(object, table, key); err != nil {
		return "", err
	}

	// open the data file for the userhash, object, and table
	data, err := openData(userhash, object, table)
	if err != nil {
		return "", err
	}
	//  if key is empty, return the entire data
	if key == "" {
		jsonData, err := json.Marshal(data)
		if err != nil {
			return "", fmt.Errorf("failed to marshal data: %w", err)
		}
		return string(jsonData), nil
	}
	// return the value of the key from the data
	log.Println("return:", data.Data[key])
	return data.Data[key], nil
}
func setData(userhash, object, table, key, value string) error {
	log.Println("setData", userhash, object, table, key, value)

	// Validate input
	if err := validateInput(object, table, key); err != nil {
		return err
	}

	if key == "" {
		return fmt.Errorf("key cannot be empty")
	}

	// open the data file for the userhash, object, and table
	data, err := openData(userhash, object, table)
	if err != nil {
		return err
	}
	// set the value of the key in the data
	data.Data[key] = value
	// marshal the data into json and write it to the file
	jsonData, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to marshal data: %w", err)
	}

	tableFile, err := sanitizePath(userhash, object, table)
	if err != nil {
		return err
	}

	// Get file lock
	fileLock := getFileLock(tableFile)
	fileLock.Lock()
	defer fileLock.Unlock()

	if err := os.WriteFile(tableFile, jsonData, 0644); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}
	return nil
}
func deleteData(userhash, object, table, key string) error {
	log.Println("deleteData", userhash, object, table, key)

	// Validate input
	if err := validateInput(object, table, key); err != nil {
		return err
	}

	if key == "" {
		return fmt.Errorf("key cannot be empty")
	}

	// open the data file for the userhash, object, and table
	data, err := openData(userhash, object, table)
	if err != nil {
		return err
	}
	// delete the key from the data
	delete(data.Data, key)
	// marshal the data into json and write it to the file
	jsonData, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to marshal data: %w", err)
	}

	tableFile, err := sanitizePath(userhash, object, table)
	if err != nil {
		return err
	}

	// Get file lock
	fileLock := getFileLock(tableFile)
	fileLock.Lock()
	defer fileLock.Unlock()

	if err := os.WriteFile(tableFile, jsonData, 0644); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}
	return nil
}
