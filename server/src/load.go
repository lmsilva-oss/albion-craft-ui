package src

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/spf13/viper"
)

// tightly coupled to the dump's formatting
var localizationKeys = []string{"EN-US", "DE-DE", "FR-FR", "RU-RU", "PL-PL", "ES-ES", "PT-BR", "IT-IT", "ZH-CN", "KO-KR", "JA-JP"}

type Item struct {
	LocalizationNameVariable        string
	LocalizationDescriptionVariable string
	LocalizedNames                  map[string]string
	LocalizedDescriptions           map[string]string
	Index                           string
	UniqueName                      string
}

func loadItemsJSON() []Item {
	location := viper.GetString("ao-data-api.itemsJSON")
	client := http.Client{}

	resp, err := client.Get(location)
	if err != nil {
		log.Fatal(err)
	}

	defer resp.Body.Close()

	var items []Item
	err = json.NewDecoder(resp.Body).Decode(&items)
	if err != nil {
		log.Fatal(err)
	}

	return items
}

// TODO: use AO data Swagger client to fetch market data
func setupAODataClient() {
	host := viper.GetString("ao-data-api.host")
	fmt.Println(host)
}

func Start() {
	items := loadItemsJSON()

	for _, item := range items {
		localizationKey := localizationKeys[6] // PT_BR
		fmt.Println(item.UniqueName, item.LocalizedNames[localizationKey], item.LocalizedDescriptions[localizationKey])
	}

	setupAODataClient()
}
