package main

import (
	"fmt"

	"github.com/sherifabdlnaby/configuro"
)

//Config Is our main Application Config Struct.
type Config struct {
	// Primitive Types
	Number      int
	NumberList  []int `config:"number_list"`
	Word        string
	AnotherWord string            `config:"another_word"`
	WordMap     map[string]string `config:"word_map"`
	// Nested Objects (Ptr and None)
	Database *Database
	Logger   Logger
}

//Database A sub-config struct
type Database struct {
	Hosts    []string
	Username string
	Password string
}

//Logger Another sub-config struct
type Logger struct {
	Level string
	Debug bool
}

func main() {

	// Create Configuro Object
	Loader, err := configuro.NewConfig()
	if err != nil {
		panic(err)
	}

	// Create our Config holding Struct
	config := &Config{Word: "default value in struct."}

	// Load Our Config.
	err = Loader.Load(config)
	if err != nil {
		panic(err)
	}

	// Print Result.
	fmt.Printf(`
Config Struct:
	Number: %d
	NumberList: %v
	--------------
	Word: %s
	AnotherWord: %s
	WordMap: %v
	--------------
	Database:
		Hosts: %v
		Username: %s
		Password: %s
	--------------
	Logger:
		Level: %s
		Debug: %t
	`, config.Number, config.NumberList, config.Word, config.AnotherWord, config.WordMap, config.Database.Hosts,
		config.Database.Username, config.Database.Password, config.Logger.Level, config.Logger.Debug)
}
