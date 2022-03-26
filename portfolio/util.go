package main

import (
	"encoding/xml"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
)

type PortfolioRecord struct {
	RecordType string
	Symbol     string
	Amount     int
}

func readPortfolio() []PortfolioRecord {
	dat, err := os.ReadFile("portfolio.txt")
	if errors.Is(err, os.ErrNotExist) {
		fmt.Print("portfolio.txt is not found")
		return []PortfolioRecord{}
	}

	var result []PortfolioRecord
	for _, x := range strings.Split(string(dat), "\r\n") {
		var r PortfolioRecord
		splitString := strings.Split(string(x), " ")
		r.RecordType = splitString[0]
		r.Symbol = splitString[1]
		r.Amount, _ = strconv.Atoi(splitString[2])
		result = append(result, r)
	}

	return result
}

func readToken() string {
	dat, err := os.ReadFile("token.txt")
	if errors.Is(err, os.ErrNotExist) {
		fmt.Print("Put your token into token.txt file")
		return ""
	}

	return string(dat)
}

func readDataFromBinance() string {
	response, err := http.Get("http://binance.com/api/v3/ticker/price")
	if err != nil {
		fmt.Print(err.Error())
		return ""
	}

	responseData, err := ioutil.ReadAll(response.Body)
	if err != nil {
		log.Fatal(err)
	}

	return string(responseData)
}

func getPriceFromBinance(responseData string, ticker string) float32 {

	dataString := string(responseData)
	dataString = dataString[:len(dataString)-1]
	dataString = dataString[1:]
	splitData := strings.Split(dataString, "},{")
	for _, symbolAndPrice := range splitData {
		splitString := strings.Split(symbolAndPrice, ",")
		symbol := strings.Split(splitString[0], ":")[1]
		if symbol != ("\"" + ticker + "\"") {
			continue
		}

		price := strings.Split(splitString[1], ":")[1]
		price = price[:len(price)-1]
		price = price[1:]
		value, _ := strconv.ParseFloat(price, 32)
		return float32(value)
	}

	return 0
}

type StockRecord struct {
	XMLName xml.Name `xml:"row"`
	Ticker  string   `xml:"SECID,attr"`
	Price   string   `xml:"PREVADMITTEDQUOTE,attr"`
}

type Stocks struct {
	XMLName xml.Name      `xml:"rows"`
	Stocks  []StockRecord `xml:"row"`
}

type Data struct {
	XMLName xml.Name `xml:"data"`
	Id      string   `xml:"id,attr"`
	Content Stocks   `xml:"rows"`
}

type MoexResponse struct {
	XMLName xml.Name `xml:"document"`
	Content Data     `xml:"data"`
}

func getPriceFromMoex(ticker string, rubPrice float32) float32 {
	response, err := http.Get("https://iss.moex.com/iss/engines/stock/markets/shares/boards/TQBR/securities.xml?iss.meta=off&iss.only=securities&securities.columns=SECID,PREVADMITTEDQUOTE")

	if err != nil {
		fmt.Print(err.Error())
		os.Exit(1)
	}

	responseData, err := ioutil.ReadAll(response.Body)
	if err != nil {
		log.Fatal(err)
	}

	var moexDoc MoexResponse
	xml.Unmarshal(responseData, &moexDoc)

	// we iterate through every user within our users array and
	// print out the user Type, their name, and their facebook url
	// as just an example
	for i := 0; i < len(moexDoc.Content.Content.Stocks); i++ {
		if moexDoc.Content.Content.Stocks[i].Ticker != ticker {
			continue
		}

		value, _ := strconv.ParseFloat(moexDoc.Content.Content.Stocks[i].Price, 32)
		return float32(value) / rubPrice
	}

	return 0
}
