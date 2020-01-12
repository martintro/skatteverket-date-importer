package main

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"os"
	"strconv"
	"time"
)

var months = map[string]string{
	"jan": "01",
	"feb": "02",
	"mar": "03",
	"apr": "04",
	"maj": "05",
	"jun": "06",
	"jul": "07",
	"aug": "08",
	"sep": "09",
	"okt": "10",
	"nov": "11",
	"dec": "12",
}

type responseObject struct {
	Years yearWithDates `json:"viktigaDatum,omitempty"`
}

type yearWithDates struct {
	Upcoming []event `json:"kommande,omitempty"`
	Earlier  []event `json:"tidigare,omitempty"`
}

type event struct {
	Year           int    `json:"ar,omitempty"`
	Day            int    `json:"dag,omitempty"`
	MonthName      string `json:"manad,omitempty"`
	MonthShortName string `json:"manadShort,omitempty"`
	Category       string `json:"kategori,omitempty"`
	Type           string `json:"typ,omitempty"`
	WebSkvPath     string `json:"url,omitempty"`
}

// Possible values UPP_TILL_EN_MILJON, UPP_TILL_FYRTIO_MILJONER, OVER_FYRTIO_MILJONER
var revenue = "UPP_TILL_EN_MILJON"

// Possible values HELAR, KVARTAL, MANAD
var vatDeclarationPeriod = "HELAR"

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

func createIcsFile(events *responseObject) error {
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

	for _, event := range events.Years.Upcoming {
		_, err = file.WriteString("BEGIN:VEVENT\n" +
			"TRANSP:TRANSPARENT\n" +
			"UID:" + strconv.Itoa(rand.Int()) + "@martintroedsson\n" +
			"DTSTAMP:" + time.Now().Format("20060102T150405Z") + "\n" +
			"DTSTART;VALUE=DATE:" + strconv.Itoa(event.Year) + months[event.MonthShortName] + fmt.Sprintf("%02d", event.Day) + "\n" +
			"SEQUENCE:0\n" +
			"CLASS:PUBLIC\n" +
			"SUMMARY:" + event.Category + "\n" +
			"DESCRIPTION:" + event.Type + " " + event.Category + ". https://www.skatteverket.se" + event.WebSkvPath + "\n" +
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

func getDatesFromSkatteverket(financialYearLastMonth string, paysSalary string, revenue string,
	vatDeclarationPeriod string) (*responseObject, error) {
	var result *responseObject

	response, err := http.Get("https://www.skatteverket.se/privat/etjansterochblanketter/" +
		"viktigadatum.4.5d699354142b230302034e.12.3810a01c150939e893f2b2bd.portlet" +
		"?struts.portlet.action=/vd/json-get-viktiga-datum&malgrupp=foretag" +
		"&foretagsform=AKTIEBOLAG_FORENINGAR&omsattning=" + revenue + "&redovisning=" +
		vatDeclarationPeriod + "&rakenskapsar-slut=" + financialYearLastMonth +
		"&arbetsgivare=" + paysSalary + "&spara-filterval=false")
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
