package main

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
)

// TODO: these const should come from CLI flags
const MAX_RECORDS = 2   // 7095
const READ_BUFFER = 10  // 5000
const WRITE_BUFFER = 10 // 300
const AO2D_TIME_BETWEEN_REQUESTS = 500 * time.Millisecond
const recipesCSV = "recipes.csv"
const itemsCSV = "items.csv"

type ScrapedItem struct {
	ItemID       string
	ScrapeResult string
}

func (i ScrapedItem) ToCSV() []string {
	return []string{i.ItemID, i.ScrapeResult}
}

func Scrape(filename string) chan ScrapedItem {
	readFile := func(filename string, itemIDsChan chan string) {
		f, err := os.Open(filename)
		if err != nil {
			log.Fatal("Unable to open input file "+filename, err)
		}
		defer f.Close() // TODO: do not use defer

		reader := csv.NewReader(f)
		header, err := reader.Read()

		if err != nil {
			log.Fatal(err)
		}

		var index int

		for idx, val := range header {
			if val == "split_id" {
				index = idx
				break
			}
		}

		for i := 0; i < MAX_RECORDS; i++ {
			record, err := reader.Read()
			// Stop at EOF.
			if err == io.EOF {
				break
			}

			if err != nil {
				log.Fatal(err)
			}

			itemIDbytes := []byte(strings.ReplaceAll(record[index], "'", "\""))

			var itemIDslice []string
			if err := json.Unmarshal(itemIDbytes, &itemIDslice); err != nil {
				fmt.Println(err)
			}

			itemID := strings.Join(itemIDslice, "_")
			log.Printf("read item ID %s (%d/%d)", itemID, i+1, MAX_RECORDS)
			itemIDsChan <- itemID
		}
		close(itemIDsChan)
	}

	fetch := func(itemIDsChan chan string, outputChan chan ScrapedItem) {
		ids := 0
		for itemID := range itemIDsChan {
			ids += 1
			log.Printf("fetching recipe for %s (%d/%d)", itemID, ids, MAX_RECORDS)
			time.Sleep(AO2D_TIME_BETWEEN_REQUESTS)

			const baseURL = "https://www.albiononline2d.com/en/craftcalculator/api/"

			resp, err := http.Get(baseURL + itemID)
			if err != nil {
				log.Printf("Error fetching: %v", err)
			}

			body, err := io.ReadAll(resp.Body)
			if err != nil {
				log.Printf("(%s) Error reading response body: %v", itemID, err)
			}

			err = resp.Body.Close()
			if err != nil {
				log.Printf("(%s) Error closing response body: %v", itemID, err)
			}

			// mock
			// body := []byte("{\n    \"_id\": \"62d830330b1d9b001a9098ad\",\n    \"_attr\": {\n        \"uniquename\": \"T7_POTION_STONESKIN@1\",\n        \"uisprite\": \"POTION_TURQUOISE\",\n        \"abilitypower\": \"186\",\n        \"slottype\": \"potion\",\n        \"consumespell\": \"POTION_RESISTANCE_3_LEVEL1\",\n        \"shopcategory\": \"consumables\",\n        \"shopsubcategory1\": \"potion\",\n        \"craftingcategory\": \"potion\",\n        \"tier\": \"7\",\n        \"dummyitempower\": \"1100\",\n        \"maxstacksize\": \"999\",\n        \"unlockedtocraft\": \"false\",\n        \"unlockedtoequip\": \"true\",\n        \"uicraftsoundstart\": \"Play_ui_action_craft_potion_start\",\n        \"uicraftsoundfinish\": \"Play_ui_action_craft_potion_finish\",\n        \"weight\": \"1.44\",\n        \"enchantmentlevel\": \"1\"\n    },\n    \"craftingrequirements\": [\n        {\n            \"_attr\": {\n                \"silver\": \"0\",\n                \"amountcrafted\": \"5\",\n                \"forcesinglecraft\": \"false\",\n                \"craftingfocus\": \"1272\",\n                \"time\": \"2.9065\"\n            },\n            \"craftresource\": [\n                {\n                    \"_attr\": {\n                        \"uniquename\": \"T7_MULLEIN\",\n                        \"count\": \"72\"\n                    }\n                },\n                {\n                    \"_attr\": {\n                        \"uniquename\": \"T6_FOXGLOVE\",\n                        \"count\": \"36\"\n                    }\n                },\n                {\n                    \"_attr\": {\n                        \"uniquename\": \"T4_BURDOCK\",\n                        \"count\": \"36\"\n                    }\n                },\n                {\n                    \"_attr\": {\n                        \"uniquename\": \"T6_MILK\",\n                        \"count\": \"18\"\n                    }\n                },\n                {\n                    \"_attr\": {\n                        \"uniquename\": \"T7_ALCOHOL\",\n                        \"count\": \"18\"\n                    }\n                },\n                {\n                    \"_attr\": {\n                        \"uniquename\": \"T7_ESSENCE_POTION\",\n                        \"count\": \"3\"\n                    }\n                }\n            ]\n        }\n    ],\n    \"enchantments\": [\n        {\n            \"enchantment\": [\n                {\n                    \"_attr\": {\n                        \"enchantmentlevel\": \"1\",\n                        \"abilitypower\": \"186\",\n                        \"dummyitempower\": \"1100\",\n                        \"consumespell\": \"POTION_RESISTANCE_3_LEVEL1\"\n                    },\n                    \"craftingrequirements\": [\n                        {\n                            \"_attr\": {\n                                \"amountcrafted\": \"5\",\n                                \"craftingfocus\": \"1272\",\n                                \"time\": \"2.9065\"\n                            },\n                            \"craftresource\": [\n                                {\n                                    \"_attr\": {\n                                        \"uniquename\": \"T7_MULLEIN\",\n                                        \"count\": \"72\"\n                                    }\n                                },\n                                {\n                                    \"_attr\": {\n                                        \"uniquename\": \"T6_FOXGLOVE\",\n                                        \"count\": \"36\"\n                                    }\n                                },\n                                {\n                                    \"_attr\": {\n                                        \"uniquename\": \"T4_BURDOCK\",\n                                        \"count\": \"36\"\n                                    }\n                                },\n                                {\n                                    \"_attr\": {\n                                        \"uniquename\": \"T6_MILK\",\n                                        \"count\": \"18\"\n                                    }\n                                },\n                                {\n                                    \"_attr\": {\n                                        \"uniquename\": \"T7_ALCOHOL\",\n                                        \"count\": \"18\"\n                                    }\n                                },\n                                {\n                                    \"_attr\": {\n                                        \"uniquename\": \"T7_ESSENCE_POTION\",\n                                        \"count\": \"3\"\n                                    }\n                                }\n                            ]\n                        }\n                    ],\n                    \"upgraderequirements\": [\n                        {\n                            \"upgraderesource\": [\n                                {\n                                    \"_attr\": {\n                                        \"uniquename\": \"T7_ESSENCE_POTION\",\n                                        \"count\": \"3\"\n                                    }\n                                }\n                            ]\n                        }\n                    ]\n                }\n            ]\n        }\n    ],\n    \"_titles\": {\n        \"en\": \"Major Resistance Potion\",\n        \"de\": \"Großer Resistenztrank\",\n        \"fr\": \"Potion de résistance majeure\",\n        \"ru\": \"Большой эликсир защиты\",\n        \"pl\": \"Większa Mikstura Odporności\",\n        \"es\": \"Poción de resistencia mayor\",\n        \"pt\": \"Poção de Resistência Maior\"\n    },\n    \"type\": \"consumableitem\",\n    \"_calculatedEnchantmentLevel\": \"1\",\n    \"upgraderequirements\": [\n        {\n            \"upgraderesource\": [\n                {\n                    \"_attr\": {\n                        \"uniquename\": \"T7_ESSENCE_POTION\",\n                        \"count\": \"3\"\n                    }\n                }\n            ]\n        }\n    ],\n    \"__v\": 0,\n    \"_calculatedCraftFame\": 3630,\n    \"_calculatedItemValue\": 7200,\n    \"_marketplaceTitle\": \"T7_POTION_STONESKIN@1\",\n    \"title\": \"Major Resistance Potion\",\n    \"detailsHref\": \"/en/item/id/T7_POTION_STONESKIN@1\"\n}")

			result := string(body)

			output := ScrapedItem{
				ItemID:       itemID,
				ScrapeResult: result,
			}

			outputChan <- output
		}
		close(outputChan)
	}

	itemIDsChan := make(chan string, READ_BUFFER)
	outputChan := make(chan ScrapedItem, WRITE_BUFFER)

	go readFile(filename, itemIDsChan)
	go fetch(itemIDsChan, outputChan)

	return outputChan
}

type logWriter struct{}

func (writer logWriter) Write(bytes []byte) (int, error) {
	return fmt.Print(time.Now().UTC().Format("2006-01-02T15:04:05.999Z") + " [DEBUG] " + string(bytes))
}

func writeToFile(filename string, items chan ScrapedItem) {
	file, err := os.Create(filename)
	if err != nil {
		log.Fatal(err)
	}

	defer file.Close()

	w := csv.NewWriter(file)
	headers := []string{"item_id", "scrape_result"}
	w.Write(headers)
	w.Flush()
	if err := w.Error(); err != nil {
		log.Fatal(err)
	}

	count := 0
	for item := range items {
		warnString := " "
		if len(item.ScrapeResult) == 0 {
			warnString = " (empty) "
		}

		log.Printf("got %d%scharacters for item ID %s", len(item.ScrapeResult), warnString, item.ItemID)
		w.Write(item.ToCSV())
		count += 1
		if count > WRITE_BUFFER {
			log.Printf("write buffer full")
			w.Flush()
			if err := w.Error(); err != nil {
				log.Fatal(err)
			}

			count = 0
		}
	}

	w.Flush()
	if err := w.Error(); err != nil {
		log.Fatal(err)
	}
}

func main() {
	log.SetFlags(0)
	log.SetOutput(new(logWriter))

	scrapedItemsChannel := Scrape(itemsCSV)

	writeToFile(recipesCSV, scrapedItemsChannel)
}
