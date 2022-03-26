package main

import (
	"fmt"

	"github.com/tonymackay/finnhub-go"
)

func main() {
	UsStockType := "UsStock"
	MoexStockType := "MoexStock"
	CryptoType := "crypto"

	token := readFinnhubToken()
	if len(token) == 0 {
		return
	}

	portfolio := readPortfolio()
	if len(portfolio) == 0 {
		return
	}

	cryptoPrices := getPricesFromBinance()
	stockClient := finnhub.NewClient(token)

	rubPrice := getPriceFromBinance(cryptoPrices, "USDTRUB")
	totalBalance := float32(0)
	for _, asset := range portfolio {
		var price float32
		switch asset.RecordType {
		case UsStockType:
			price = getStockPrice(stockClient, asset.Symbol)
		case MoexStockType:
			price = getPriceFromMoex(asset.Symbol, rubPrice)
		case CryptoType:
			price = getPriceFromBinance(cryptoPrices, asset.Symbol)
		default:
			fmt.Printf("Undefined %s asset type. Use any from the list: %s, %s, %s",
				asset.RecordType, UsStockType, MoexStockType, CryptoType)
			continue
		}
		value := price * float32(asset.Amount)
		totalBalance = totalBalance + value

		fmt.Printf("Symbol: %s\n", asset.Symbol)
		fmt.Printf(" Price: $%f\n", price)
		fmt.Printf(" Amount: %d\n", asset.Amount)
		fmt.Printf(" Value: $%f\n", value)
	}

	fmt.Printf("\nBalance: $%f, RUB %f", totalBalance, totalBalance*rubPrice)
}
