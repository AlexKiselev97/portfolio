package main

import (
	"encoding/json"
	"encoding/xml"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/tonymackay/finnhub-go"
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
	for _, assetLine := range strings.Split(string(dat), "\r\n") {
		var r PortfolioRecord
		splitString := strings.Split(string(assetLine), " ")
		if len(splitString) != 3 {
			fmt.Printf("Invalid asset: %s\n", assetLine)
			continue
		}

		r.RecordType = splitString[0]
		r.Symbol = splitString[1]
		r.Amount, _ = strconv.Atoi(splitString[2])
		result = append(result, r)
	}

	return result
}

func readFinnhubToken() string {
	token, err := os.ReadFile("token.txt")
	if errors.Is(err, os.ErrNotExist) {
		fmt.Print("Put your finnhub token into token.txt file")
		return ""
	}

	return string(token)
}

func getStockPrice(client finnhub.Client, symbol string) float32 {
	symbol1 := strings.ToUpper(symbol)
	quote, err := client.Quote(symbol1)
	if err != nil {
		fmt.Printf("%v\n", err)
	}

	return quote.Current
}

func getPricesFromBinance() []map[string]interface{} {
	response, err := http.Get("http://binance.com/api/v3/ticker/price")
	if err != nil {
		fmt.Print(err.Error())
		return []map[string]interface{}{}
	}

	responseData, err := ioutil.ReadAll(response.Body)
	if err != nil {
		log.Fatal(err)
	}

	var prices []map[string]interface{}
	err = json.Unmarshal(responseData, &prices)
	if err != nil {
		fmt.Print(err.Error())
		return []map[string]interface{}{}
	}

	return prices
}

func getPriceFromBinance(prices []map[string]interface{}, ticker string) float32 {
	for _, asset := range prices {
		if asset["symbol"].(string) != ticker {
			continue
		}

		price, _ := strconv.ParseFloat(asset["price"].(string), 32)
		return float32(price)
	}

	return 0
}

type Stock struct {
	XMLName xml.Name `xml:"row"`
	Ticker  string   `xml:"SECID,attr"`
	Price   string   `xml:"PREVADMITTEDQUOTE,attr"`
}

type Stocks struct {
	XMLName xml.Name `xml:"rows"`
	Stocks  []Stock  `xml:"row"`
}

type Data struct {
	XMLName xml.Name `xml:"data"`
	Id      string   `xml:"id,attr"`
	Content Stocks   `xml:"rows"`
}

func getPriceFromMoex(ticker string, rubPrice float32) float32 {
	if rubPrice == 0 {
		return 0
	}

	response, err := http.Get("https://iss.moex.com/iss/engines/stock/markets/shares/boards/TQBR/securities.xml?iss.meta=off&iss.only=securities&securities.columns=SECID,PREVADMITTEDQUOTE")
	if err != nil {
		fmt.Print(err.Error())
		return 0
	}

	responseData, err := ioutil.ReadAll(response.Body)
	if err != nil {
		log.Fatal(err)
	}

	var moexDoc struct {
		XMLName xml.Name `xml:"document"`
		Content Data     `xml:"data"`
	}
	xml.Unmarshal(responseData, &moexDoc)
	for _, stock := range moexDoc.Content.Content.Stocks {
		if stock.Ticker != ticker {
			continue
		}

		price, _ := strconv.ParseFloat(stock.Price, 32)
		return float32(price) / rubPrice
	}

	return 0
}
