package main

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

type responseObject []event

type event struct {
	Name             string        `json:"name"`
	Type             string        `json:"type"`
	Customer         string        `json:"customer"`
	Date             string        `json:"date"`
	BusinessTypes    []string      `json:"business_types"`
	FiscalYear       []string      `json:"fiscal_year"`
	Turnover         []string      `json:"turnover"`
	AccountingPeriod []string      `json:"accounting_period"`
	Description      description   `json:"description"`
	RelatedLinks     []relatedLink `json:"related_links"`
}

type description struct {
	Content []paragraph `json:"content"`
}

type paragraph struct {
	Content []textNode `json:"content"`
}

type textNode struct {
	Value string `json:"value"`
}

type relatedLink struct {
	Name string `json:"name"`
	Link string `json:"link"`
}

// Possible values Privat, Företag
var customer = "Företag"

// Possible values Aktiebolag, Förening, Handelsbolag, Kommanditbolag, Enskild näringsidkare
var businessType = "Aktiebolag"

// Possible values Januari-December
var fiscalYear = "Augusti"

// Possible values 0-1, 1-40, Över 40
var turnover = "0-1"

// Possible values Helår, Kvartal, Månad
var accountingPeriod = "Helår"

func main() {
	fmt.Println("Starting the import...")

	result, err := getDatesFromSkatteverket()
	if err != nil {
		fmt.Println("An error occurred:", err)
		return
	}

	result = filterImportantDates(result, customer, businessType, fiscalYear, turnover, accountingPeriod)

	err = createIcsFile(result)
	if err != nil {
		fmt.Println("An error occurred:", err)
		return
	}

	fmt.Println("Import done!")
}

func createIcsFile(response responseObject) error {
	file, err := os.Create("skatteverket.ics")
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = file.WriteString("BEGIN:VCALENDAR\n" +
		"PRODID:martintroedsson\n" +
		"VERSION:2.0\n" +
		"CALSCALE:GREGORIAN\n")
	if err != nil {
		return err
	}

	for _, event := range response {
		_, err = file.WriteString("BEGIN:VEVENT\n" +
			"TRANSP:TRANSPARENT\n" +
			"UID:" + strconv.Itoa(rand.Int()) + "@martintroedsson\n" +
			"DTSTAMP:" + time.Now().Format("20060102T150405Z") + "\n" +
			"DTSTART;VALUE=DATE:" + strings.ReplaceAll(event.Date, "-", "") + "\n" +
			"SEQUENCE:0\n" +
			"CLASS:PUBLIC\n" +
			"SUMMARY:" + event.Name + "\n" +
			"DESCRIPTION:" + event.Name + "\n" +
			"END:VEVENT\n")
		if err != nil {
			return err
		}
	}

	_, err = file.WriteString("END:VCALENDAR")
	if err != nil {
		return err
	}

	return nil
}

func filterImportantDates(events responseObject, customer string, businessType string, fiscalYear string, turnover string, accountingPeriod string) responseObject {
	var filtered responseObject

	for _, e := range events {
		if e.Customer != customer {
			continue
		}
		if businessType != "" && len(e.BusinessTypes) > 0 && !contains(e.BusinessTypes, businessType) {
			continue
		}
		if fiscalYear != "" && len(e.FiscalYear) > 0 && !contains(e.FiscalYear, fiscalYear) {
			continue
		}
		if turnover != "" && len(e.Turnover) > 0 && !contains(e.Turnover, turnover) {
			continue
		}
		if accountingPeriod != "" && len(e.AccountingPeriod) > 0 && !contains(e.AccountingPeriod, accountingPeriod) {
			continue
		}

		if strings.Contains(strings.ToLower(e.Name), "fastighetsdeklaration") ||
			strings.Contains(strings.ToLower(e.Name), "moms på ioss") ||
			strings.Contains(strings.ToLower(e.Name), "moms på oss") ||
			strings.Contains(strings.ToLower(e.Name), "beslut om fastighetstaxering") ||
			strings.Contains(strings.ToLower(e.Name), "särskild fastighetstaxering") ||
			strings.Contains(strings.ToLower(e.Name), "ansökan om utbetalning för grön teknik") ||
			strings.Contains(strings.ToLower(e.Name), "ansökan om rot- och rututbetalning") ||
			strings.Contains(strings.ToLower(e.Name), "periodisk sammanställning på blankett") {
			continue
		}

		filtered = append(filtered, e)
	}

	return filtered
}

func getDatesFromSkatteverket() (responseObject, error) {
	var result responseObject

	var url = "https://www7.skatteverket.se/portal-wapi/open/content/important-dates"
	fmt.Println("Using URL:", url)

	client := http.Client{}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return result, err
	}
	req.Header.Add("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/87.0.4280.141 Safari/537.36")
	response, err := client.Do(req)
	if err != nil {
		return result, err
	}
	defer response.Body.Close()

	decoder := json.NewDecoder(response.Body)
	result = responseObject{}
	err = decoder.Decode(&result)
	if err != nil {
		return result, err
	}

	return result, nil
}

func contains(list []string, value string) bool {
	value = strings.ToLower(value)
	for _, v := range list {
		if strings.Contains(strings.ToLower(v), value) {
			return true
		}
	}
	return false
}
