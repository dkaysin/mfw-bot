package main

import (
	"encoding/json"
	"io/ioutil"
)

func GetData(d *Data) *Data {
	data, err := ioutil.ReadFile("data.json")
	if err != nil {
		panic(err)
	}

	err = json.Unmarshal(data, d)
	if err != nil {
		panic(err)
	}

	return d
}
