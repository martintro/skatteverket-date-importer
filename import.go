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

type responseObject struct {
	Events []event `json:"viktigaDatum,omitempty"`
}

type event struct {
	Type       string   `json:"type,omitempty"`
	Category   string   `json:"category,omitempty"`
	WebSkvPath string   `json:"uri,omitempty"`
	Dates      []string `json:"dates,omitempty"`
}

// Possible values UPP_TILL_EN_MILJON, MER_AN_EN_MILJON_TILL_FYRTIO_MILJONER, OVER_FYRTIO_MILJONER
var revenue = "UPP_TILL_EN_MILJON"

// Possible values AR, KVARTAL, MANAD
var vatDeclarationPeriod = "AR"

// Possible values true, false
var paysSalary = "true"

// Possible values 1-12
var financialYearLastMonth = "8"

func main() {
	fmt.Println("Starting the import...")

	result, err := getDatesFromSkatteverket(financialYearLastMonth, paysSalary, revenue, vatDeclarationPeriod)
	if err != nil {
		fmt.Println("An error occurred:", err)
		return
	}

	err = createIcsFile(result)
	if err != nil {
		fmt.Println("An error occurred:", err)
		return
	}

	fmt.Println("Import done!")
}

func createIcsFile(response *responseObject) error {
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

	for _, event := range response.Events {
		for _, date := range event.Dates {
			_, err = file.WriteString("BEGIN:VEVENT\n" +
				"TRANSP:TRANSPARENT\n" +
				"UID:" + strconv.Itoa(rand.Int()) + "@martintroedsson\n" +
				"DTSTAMP:" + time.Now().Format("20060102T150405Z") + "\n" +
				"DTSTART;VALUE=DATE:" + strings.ReplaceAll(date, "-", "") + "\n" +
				"SEQUENCE:0\n" +
				"CLASS:PUBLIC\n" +
				"SUMMARY:" + event.Category + "\n" +
				"DESCRIPTION:" + event.Type + " " + event.Category + ". https://www.skatteverket.se" + event.WebSkvPath + "\n" +
				"END:VEVENT\n")
			if err != nil {
				return err
			}
		}
	}

	_, err = file.WriteString("END:VCALENDAR")
	if err != nil {
		return err
	}

	return nil
}

func getDatesFromSkatteverket(financialYearLastMonth string, paysSalary string, revenue string,
	vatDeclarationPeriod string) (*responseObject, error) {
	var result *responseObject

	var url = "https://skatteverket.se/viktiga-datum-api/api/v1/viktiga-datum-foretag?foretagsform=AKTIEBOLAG_FORENINGAR&" +
		"omsattning=" + revenue + "&momsredovisningsperiod=" + vatDeclarationPeriod + "&rakenskapsaretsSistaManad=" + financialYearLastMonth + "&" +
		"arbetsgivare=" + paysSalary + "&tidigareDatum=false"
	fmt.Println("Using URL: " + url)
	response, err := http.Get(url)
	if err != nil {
		return result, err
	}
	defer response.Body.Close()

	decoder := json.NewDecoder(response.Body)
	result = &responseObject{}
	err = decoder.Decode(result)
	if err != nil {
		return result, err
	}

	return result, nil
}
