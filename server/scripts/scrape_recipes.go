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
const MAX_RECORDS = 7095
const READ_BUFFER = 5000
const WRITE_BUFFER = 300

const AO2D_TIME_BETWEEN_REQUESTS = 50 * time.Millisecond
const HTTP_FETCHERS = 2
const HTTP_TIMEOUT = 2 * time.Second
const baseURL = "https://www.albiononline2d.com/en/craftcalculator/api/"

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

	ids := 0
	client := http.Client{
		Timeout: HTTP_TIMEOUT,
	}

	fetch := func(workerID int, itemIDChan chan string, outputChan chan ScrapedItem, doneChan chan bool) {
		for itemID := range itemIDChan {
			if strings.Contains(itemID, "LABOURER_CONTRACT") {
				log.Printf("skipping %s (LABOURER_CONTRACT)", itemID)
				continue
			}

			log.Printf("(worker #%d) fetching recipe for %s (%d/%d)", workerID, itemID, ids, MAX_RECORDS)
			time.Sleep(AO2D_TIME_BETWEEN_REQUESTS)

			resp, err := client.Get(baseURL + itemID)
			if err != nil {
				log.Printf("Error fetching: %v", err)
				continue
			}

			body, err := io.ReadAll(resp.Body)
			if err != nil {
				log.Printf("(%s) Error reading response body: %v", itemID, err)
				continue
			}

			err = resp.Body.Close()
			if err != nil {
				log.Printf("(%s) Error closing response body: %v", itemID, err)
				continue
			}

			result := string(body)

			output := ScrapedItem{
				ItemID:       itemID,
				ScrapeResult: result,
			}

			outputChan <- output
		}
		doneChan <- true
	}

	batchFetch := func(itemIDsChan chan string, outputChan chan ScrapedItem) {
		inputChan := make(chan string, HTTP_FETCHERS)
		doneChan := make(chan bool)

		for i := 0; i < HTTP_FETCHERS; i++ {
			go fetch(i, inputChan, outputChan, doneChan)
		}

		for itemID := range itemIDsChan {
			ids += 1
			inputChan <- itemID
		}
		close(inputChan)

		for i := 0; i < HTTP_FETCHERS; i++ {
			<-doneChan
		}
		close(outputChan)
	}

	itemIDsChan := make(chan string, READ_BUFFER)
	outputChan := make(chan ScrapedItem, WRITE_BUFFER)

	go readFile(filename, itemIDsChan)
	go batchFetch(itemIDsChan, outputChan)

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
