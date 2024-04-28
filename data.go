package main

import (
	"encoding/json"
	"io/ioutil"
	"os"
)

type ObjectData struct {
	//  arbitrary data of key value pairs
	Data map[string]string `json:"data"`
}

var emptyData = `{"data": {}}`
var defaultData = ObjectData{Data: map[string]string{}}

func verifyFile(datapath string, objectfile string) error {
	// check if objectfile exists
	filePath := "data/" + datapath + "/" + objectfile + ".json"
	// if folder  does not exist, create it
	if _, err := os.Stat("data/" + datapath); os.IsNotExist(err) {
		err := os.Mkdir("data/"+datapath, 0755)
		if err != nil {
			return err
		}
	}
	// if file does not exist, create it
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		file, err := os.Create(filePath)
		if err != nil {
			return err
		}
		defer file.Close()
		_, err = file.WriteString(emptyData)
		if err != nil {
			return err
		}
	}
	return nil
}

func readEntireData(datapath string, objectfile string) (ObjectData, error) {
	// read data from objectfile as json
	filePath := "data/" + datapath + "/" + objectfile + ".json"
	file, err := os.Open(filePath)
	if err != nil {
		return defaultData, err
	}
	defer file.Close()
	// return data as ObjectData
	byteValue, err := ioutil.ReadAll(file)
	if err != nil {
		return defaultData, err
	}

	var data ObjectData
	err = json.Unmarshal(byteValue, &data)
	if err != nil {
		return defaultData, err
	}

	return data, nil
}

func readData(datapath string, objectfile string) (ObjectData, error) {
	// read data from objectfile as json
	filePath := "data/" + datapath + "/" + objectfile + ".json"
	file, err := os.Open(filePath)
	if err != nil {
		return ObjectData{}, err
	}
	defer file.Close()

	byteValue, err := ioutil.ReadAll(file)
	if err != nil {
		return ObjectData{}, err
	}

	var data ObjectData
	err = json.Unmarshal(byteValue, &data)
	if err != nil {
		return ObjectData{}, err
	}

	return data, nil
}

func writeData(datapath string, objectfile string, key string, value string) error {
	// read data from objectfile
	data, err := readData(datapath, objectfile)
	if err != nil {
		return err
	}

	// set key value pair
	data.Data[key] = value

	// write data to objectfile
	filePath := "data/" + datapath + "/" + objectfile + ".json"
	file, err := os.OpenFile(filePath, os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	byteValue, err := json.Marshal(data)
	if err != nil {
		return err
	}

	_, err = file.Write(byteValue)
	if err != nil {
		return err
	}

	return nil
}

func deleteData(datapath string, objectfile string, key string) error {
	// read data from objectfile
	data, err := readData(datapath, objectfile)
	if err != nil {
		return err
	}

	// delete key value pair
	delete(data.Data, key)

	// write data to objectfile
	filePath := "data/" + datapath + "/" + objectfile + ".json"
	file, err := os.OpenFile(filePath, os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	byteValue, err := json.Marshal(data)
	if err != nil {
		return err
	}

	_, err = file.Write(byteValue)
	if err != nil {
		return err
	}

	return nil
}
