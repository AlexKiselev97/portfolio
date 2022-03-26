package main

import (
	"fmt"
	"strings"

	stockFinnhub "github.com/tonymackay/finnhub-go"
)

func getStockPrice(client stockFinnhub.Client, symbol string) float32 {
	symbol1 := strings.ToUpper(symbol)
	quote, err := client.Quote(symbol1)
	if err != nil {
		fmt.Printf("%v\n", err)
	}

	return quote.Current
}

func main() {
	token := readToken()
	if len(token) == 0 {
		return
	}

	portfolio := readPortfolio()
	if len(portfolio) == 0 {
		return
	}

	binanceResponse := readDataFromBinance()
	client := stockFinnhub.NewClient(token)

	rubPrice := getPriceFromBinance(binanceResponse, "USDTRUB")
	totalBalance := float32(0)
	for _, record := range portfolio {
		var price float32
		if record.RecordType == "UsStock" {
			price = getStockPrice(client, record.Symbol)
		} else if record.RecordType == "MoexStock" {
			price = getPriceFromMoex(record.Symbol, rubPrice)
		} else if record.RecordType == "crypto" {
			price = getPriceFromBinance(binanceResponse, record.Symbol)
		} else {
			continue
		}

		fmt.Printf("Symbol: %s\n", record.Symbol)
		fmt.Printf(" Price: $%f\n", price)
		fmt.Printf(" Amount: %d\n", record.Amount)

		value := price * float32(record.Amount)
		fmt.Printf(" In portfolio: $%f\n", value)

		totalBalance = totalBalance + value
	}

	fmt.Printf("\nBalance: $%f, RUB %f", totalBalance, totalBalance*rubPrice)
}
