package main

import (
	"encoding/json"
	"log"
	"os"
	"path/filepath"
)

type ObjectData struct {
	//  arbitrary data of key value pairs
	Data map[string]string `json:"data"`
}

var emptyData = `{"data": {}}`
var defaultData = ObjectData{Data: map[string]string{}}

func openData(userhash, object, table string) (error, ObjectData) {
	log.Println("openData", userhash, object, table)
	// open the data file for the userhash, object, and table from
	// the ./data/userhash/object/table.json file
	tableFile := "./data/" + userhash + "/" + object + "/" + table + ".json"

	// if the file and parent folders do not exist, create them
	if _, err := os.Stat(tableFile); os.IsNotExist(err) {
		if err := os.MkdirAll(filepath.Dir(tableFile), 0755); err != nil {
			return err, ObjectData{}
		}
		if err := os.WriteFile(tableFile, []byte(emptyData), 0644); err != nil {
			return err, ObjectData{}
		}
	}
	// read the data from the file and unmarshal it into the object data
	data, err := os.ReadFile(tableFile)
	if err != nil {
		return err, ObjectData{}
	}
	parsedData := ObjectData{}
	if err := json.Unmarshal(data, &parsedData); err != nil {
		return err, ObjectData{}
	}
	return nil, parsedData
}
func getData(userhash, object, table, key string) (string, error) {
	log.Println("getData", userhash, object, table, key)
	// open the data file for the userhash, object, and table
	err, data := openData(userhash, object, table)
	if err != nil {
		return "", err
	}
	//  if key is empty, return the entire data
	if key == "" {
		jsonData, err := json.Marshal(data)
		if err != nil {
			return "", err
		}
		return string(jsonData), nil
	}
	// return the value of the key from the data
	log.Println("return:", data.Data[key])
	return data.Data[key], nil
}
func setData(userhash, object, table, key, value string) error {
	log.Println("setData", userhash, object, table, key, value)
	// open the data file for the userhash, object, and table
	err, data := openData(userhash, object, table)
	if err != nil {
		return err
	}
	// set the value of the key in the data
	data.Data[key] = value
	// marshal the data into json and write it to the file
	jsonData, err := json.Marshal(data)
	if err != nil {
		return err
	}
	tableFile := ""
	if userhash == "" {
		tableFile = "./data/default/" + object + "/" + table + ".json"
	} else {
		tableFile = "./data/" + userhash + "/" + object + "/" + table + ".json"
	}
	if err := os.WriteFile(tableFile, jsonData, 0644); err != nil {
		return err
	}
	return nil
}
func deleteData(userhash, object, table, key string) error {
	log.Println("deleteData", userhash, object, table, key)
	// open the data file for the userhash, object, and table
	err, data := openData(userhash, object, table)
	if err != nil {
		return err
	}
	// delete the key from the data
	delete(data.Data, key)
	// marshal the data into json and write it to the file
	jsonData, err := json.Marshal(data)
	if err != nil {
		return err
	}
	tableFile := ""
	if userhash == "" {
		tableFile = "./data/default/" + object + "/" + table + ".json"
	} else {
		tableFile = "./data/" + userhash + "/" + object + "/" + table + ".json"
	}
	if err := os.WriteFile(tableFile, jsonData, 0644); err != nil {
		return err
	}
	return nil
}
