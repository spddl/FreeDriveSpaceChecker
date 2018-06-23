package config

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
)

// Config Struktur
type Config struct { // https://mholt.github.io/json-to-go/
	Drive  string `json:"drive"`
	Alarm1 struct {
		Value        float64 `json:"value"`
		Sound        string  `json:"sound"`
		Notification bool    `json:"notification"`
	} `json:"alarm1"`
	Alarm2 struct {
		Value        float64 `json:"value"`
		Sound        string  `json:"sound"`
		Notification bool    `json:"notification"`
	} `json:"alarm2"`
	Alarm3 struct {
		Value        float64 `json:"value"`
		Sound        string  `json:"sound"`
		Notification bool    `json:"notification"`
	} `json:"alarm3"`
}

func Exists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return true, err
}

// Init zumn initialisieren der Conifg
func Init() Config {
	found, e := Exists("./config.json")
	if e != nil {
		fmt.Printf("File error: %v\n", e)
		os.Exit(1)
	}

	if !found {
		return NewConfig() // wenn es keine config.json gibt dann erstelle sie
	} else {
		return LoadConfigFile() // sonst lese sie
	}
}

// LoadConfigFile läd die Config
func LoadConfigFile() Config {
	file, e := ioutil.ReadFile("./config.json")
	if e != nil {
		fmt.Printf("File error: %v\n", e)
		os.Exit(1)
	}

	var jsontype Config
	err := json.Unmarshal(file, &jsontype)
	if err != nil {
		fmt.Println("error:", err)
	}
	return jsontype
}

// SaveConfigFile Speichert die Config
func SaveConfigFile(jsonBlob Config) bool { // https://github.com/spddl/csgo-reporter/blob/master/Config/Config.go#L147
	bytes, err := json.Marshal(jsonBlob)
	if err != nil {
		panic("Config konnte nicht gespeichert werden (json error)")
	} else {
		err = ioutil.WriteFile("./config.json", bytes, 0644)
		if err == nil {
			// fmt.Println("gespeichert.")
			return true
		} else {
			panic(err)
		}
	}
}

// NewConfig erstelle eine Config
func NewConfig() Config {
	jsonBlob := json.RawMessage(`{
		"drive": "C:\\",
		"alarm1": {
			"value": 4096,
			"sound": "./sound1.mp3",
			"notification": true
		},
		"alarm2": {
			"value": 2048,
			"sound": "./sound2.mp3",
			"notification": true
		},
		"alarm3": {
			"value": 1024,
			"sound": "./sound3.mp3",
			"notification": true
		}
	}`)

	var jsontype Config
	err := json.Unmarshal(jsonBlob, &jsontype)
	if err != nil {
		panic(err)
	}
	return jsontype
}
