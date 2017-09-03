package main

import (
	"github.com/Charrette/poloniex"
	"github.com/davecgh/go-spew/spew"
	"github.com/sirupsen/logrus"
)

func main() {

	p := poloniex.New()

	// volume, err := p.Get24hVolume()
	// if err != nil {
	// 	fmt.Println(err)
	// 	return
	// }

	// spew.Dump(volume)

	// tickers, err := p.GetTickers()
	// if err != nil {
	// 	logrus.Fatal(err)
	// }

	// spew.Dump(tickers)

	// orderBook, err := p.GetAllOrderBooks(1)
	// if err != nil {
	// 	logrus.Fatal(err)
	// }

	// spew.Dump(orderBook)

	// tradeHistory, err := p.GetTradeHistory("BTC_NXT", 1410158341, 1410499372)
	// if err != nil {
	// 	logrus.Fatal(err)
	// }

	// spew.Dump(tradeHistory)

	// chartData, err := p.GetChartData("BTC_NXT", 1405699200, 1406699200, poloniex.Period14400)
	// if err != nil {
	// 	logrus.Fatal(err)
	// }

	// spew.Dump(chartData)

	// currencies, err := p.GetCurrencies()
	// if err != nil {
	// 	logrus.Fatal(err)
	// }

	// spew.Dump(currencies)

	// loanOrders, err := p.GetLoanOrders("BTC")
	// if err != nil {
	// 	logrus.Fatal(err)
	// }

	// spew.Dump(loanOrders)

	// balances, err := p.GetBalances()
	// if err != nil {
	// 	logrus.Fatal(err)
	// }

	// spew.Dump(balances)

	completeBalances, err := p.GetCompleteBalances(poloniex.AllAccounts)
	if err != nil {
		logrus.Fatal(err)
	}

	spew.Dump(completeBalances)
	// if err := p.PublicCall(poloniex.Return24hVolume); err != nil {
	// 	fmt.Println(err)

	// 	return
	// }

	// p := poloniex.New()

	// if err := p.TradeCall(poloniex.ReturnDepositAddresses); err != nil {
	// 	fmt.Println(err)

	// 	return
	// }
}
