package poloniex

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/Charrette/poloniex/helper"
	"github.com/sirupsen/logrus"
)

type client struct {
	key    string
	secret string
}

// New instantiates a Poloniex client as a Poloniex interface.
func New(key, secret string) Poloniex {
	if key == "" || secret == "" {
		logrus.Warn("unable to retrieve Poloniex API credentials from environment. Only public calls will work.")
	}

	return &client{
		key:    key,
		secret: secret,
	}
}

func (c *client) GetTickers() ([]*Ticker, error) {
	tickersMap := make(map[string]*Ticker)

	if err := c.publicCall("returnTicker", &tickersMap); err != nil {
		return nil, err
	}

	tickers := []*Ticker{}
	for k, v := range tickersMap {
		v.Currency = k
		tickers = append(tickers, v)
	}

	return tickers, nil
}

func (c *client) Get24hVolume() (*Volume24h, error) {
	dest := make(map[string]*json.RawMessage)

	if err := c.publicCall("return24hVolume", &dest); err != nil {
		return nil, err
	}

	volume := &Volume24h{
		PrimaryCurrenciesTotals: make(map[string]string),
		Markets:                 make(map[string]map[string]string),
	}

	for k, v := range dest {
		marketVolume := map[string]string{}

		// First try to unmarshal result into a map of string.
		// Meaning it's a market volume.
		err := json.Unmarshal(*v, &marketVolume)
		if err != nil {
			var primaryCurrencyTotal string

			// If first unmarshal failed, then it might be the total of a primary currency.
			// So we try to unmarshal it into a string.
			e := json.Unmarshal(*v, &primaryCurrencyTotal)
			if e != nil {
				return nil, err
			}

			volume.PrimaryCurrenciesTotals[k] = primaryCurrencyTotal

			continue
		}

		volume.Markets[k] = marketVolume
	}

	return volume, nil
}

// As the response from Poloniex for returnOrderBook is pretty messy,
// I use this struct only to unmarshal the result,
// then a conversion is made to return a clean OrderBook structure.
type orderBookFromJSON struct {
	Asks     [][]interface{} `json:"asks"`
	Bids     [][]interface{} `json:"bids"`
	IsFrozen string          `json:"isFrozen"`
	Seq      int64           `json:"seq"`
}

func (c *client) GetOrderBook(currencyPair string, depth uint) (*OrderBook, error) {
	if currencyPair == "all" {
		return nil, errors.New("use GetAllOrderBook to get all order book")
	}

	o := &orderBookFromJSON{}

	params := []queryParam{
		queryParam{key: "currencyPair", value: currencyPair},
		queryParam{key: "depth", value: fmt.Sprintf("%v", depth)},
	}

	if err := c.publicCall("returnOrderBook", o, params...); err != nil {
		return nil, err
	}

	orderBook := c.convertOrderBook(o)
	orderBook.Pair = currencyPair

	return orderBook, nil
}

func (c *client) GetAllOrderBooks(depth uint) ([]*OrderBook, error) {
	o := make(map[string]*orderBookFromJSON)

	params := []queryParam{
		queryParam{key: "currencyPair", value: "all"},
		queryParam{key: "depth", value: fmt.Sprintf("%v", depth)},
	}

	if err := c.publicCall("returnOrderBook", &o, params...); err != nil {
		return nil, err
	}

	orderBooks := []*OrderBook{}
	for k, v := range o {
		orderBook := c.convertOrderBook(v)
		orderBook.Pair = k

		orderBooks = append(orderBooks, orderBook)
	}

	return orderBooks, nil
}

func (c *client) convertOrderBook(o *orderBookFromJSON) *OrderBook {
	orderBook := &OrderBook{
		IsFrozen: o.IsFrozen,
		Seq:      o.Seq,
	}

	for _, a := range o.Asks {
		if len(a) < 2 {
			continue
		}

		// The first value is a string instead of being a float,
		// god knows why?
		value, err := strconv.ParseFloat(fmt.Sprintf("%v", a[0]), 64)
		if err != nil {
			continue
		}

		// But the second value is a float :)
		amount, ok := a[1].(float64)
		if !ok {
			continue
		}

		orderBook.Asks = append(orderBook.Asks, &Order{Value: value, Amount: amount})
	}

	for _, b := range o.Bids {
		if len(b) < 2 {
			continue
		}

		// The first value is a string instead of being a float,
		// god knows why?
		value, err := strconv.ParseFloat(fmt.Sprintf("%v", b[0]), 64)
		if err != nil {
			continue
		}

		// But the second value is a float :)
		amount, ok := b[1].(float64)
		if !ok {
			continue
		}

		orderBook.Bids = append(orderBook.Bids, &Order{Value: value, Amount: amount})
	}

	return orderBook
}

// As the response from Poloniex for returnTradeHistory contains some strings that should be floats,
// I use this struct only to unmarshal the result,
// then a conversion is made to return a clean TradeHistory structure with floats where needed.
type tradeHistoryFromJSON struct {
	GlobalTradeID int64  `json:"globalTradeID"`
	TradeID       int64  `json:"tradeID"`
	Date          string `json:"date"`
	Type          string `json:"type"`
	Rate          string `json:"rate"`
	Amount        string `json:"amount"`
	Total         string `json:"total"`
}

func (c *client) GetTradeHistory(currencyPair string, start, end uint64) ([]*TradeHistory, error) {
	tradeHistoryFromJSON := []*tradeHistoryFromJSON{}

	params := []queryParam{
		queryParam{key: "currencyPair", value: currencyPair},
	}
	if start != 0 && end != 0 {
		params = append(params, queryParam{key: "start", value: fmt.Sprintf("%v", start)})
		params = append(params, queryParam{key: "end", value: fmt.Sprintf("%v", end)})
	}

	if err := c.publicCall("returnTradeHistory", &tradeHistoryFromJSON, params...); err != nil {
		return nil, err
	}

	tradeHistory := []*TradeHistory{}
	for _, t := range tradeHistoryFromJSON {
		rate, err := strconv.ParseFloat(t.Rate, 64)
		if err != nil {
			continue
		}

		amount, err := strconv.ParseFloat(t.Amount, 64)
		if err != nil {
			continue
		}

		total, err := strconv.ParseFloat(t.Total, 64)
		if err != nil {
			continue
		}

		tradeHistory = append(tradeHistory, &TradeHistory{
			GlobalTradeID: t.GlobalTradeID,
			TradeID:       t.TradeID,
			Date:          t.Date,
			Type:          t.Type,
			Rate:          rate,
			Amount:        amount,
			Total:         total,
		})
	}

	return tradeHistory, nil
}

func (c *client) GetChartData(currencyPair string, start, end uint64, period ChartDataPeriod) ([]*ChartData, error) {
	chartData := []*ChartData{}

	params := []queryParam{
		queryParam{key: "currencyPair", value: currencyPair},
		queryParam{key: "start", value: fmt.Sprintf("%v", start)},
		queryParam{key: "end", value: fmt.Sprintf("%v", end)},
		queryParam{key: "period", value: fmt.Sprintf("%v", period)},
	}

	if err := c.publicCall("returnChartData", &chartData, params...); err != nil {
		return nil, err
	}

	return chartData, nil
}

func (c *client) GetCurrencies() ([]*Currency, error) {
	currenciesMap := make(map[string]*Currency)

	if err := c.publicCall("returnCurrencies", &currenciesMap); err != nil {
		return nil, err
	}

	currencies := []*Currency{}
	for k, v := range currenciesMap {
		v.Name = k
		currencies = append(currencies, v)
	}

	return currencies, nil
}

// As the response from Poloniex for returnLoanOrders contains some strings that should be floats,
// I use this struct only to unmarshal the result,
// then a conversion is made to return a clean LoanOrders structure with floats where needed.
type loanFromJSON struct {
	Rate     string `json:"rate"`
	Amount   string `json:"amount"`
	RangeMin int64  `json:"rangeMin"`
	RangeMax int64  `json:"rangeMax"`
}

// As the response from Poloniex for returnLoanOrders contains some strings that should be floats,
// I use this struct only to unmarshal the result,
// then a conversion is made to return a clean LoanOrders structure with floats where needed.
type loanOrdersFromJSON struct {
	Offers  []*loanFromJSON `json:"offers"`
	Demands []*loanFromJSON `json:"demands"`
}

func (c *client) GetLoanOrders(currency string) (*LoanOrders, error) {
	loanOrdersFromJSON := &loanOrdersFromJSON{}

	if err := c.publicCall("returnLoanOrders", loanOrdersFromJSON, queryParam{key: "currency", value: currency}); err != nil {
		return nil, err
	}

	loanOrders := &LoanOrders{}

	for _, o := range loanOrdersFromJSON.Offers {
		rate, err := strconv.ParseFloat(o.Rate, 64)
		if err != nil {
			continue
		}

		amount, err := strconv.ParseFloat(o.Amount, 64)
		if err != nil {
			continue
		}

		loanOrders.Offers = append(loanOrders.Offers, &Loan{
			Rate:     rate,
			Amount:   amount,
			RangeMin: o.RangeMin,
			RangeMax: o.RangeMax,
		})
	}

	for _, d := range loanOrdersFromJSON.Demands {
		rate, err := strconv.ParseFloat(d.Rate, 64)
		if err != nil {
			continue
		}

		amount, err := strconv.ParseFloat(d.Amount, 64)
		if err != nil {
			continue
		}

		loanOrders.Demands = append(loanOrders.Demands, &Loan{
			Rate:     rate,
			Amount:   amount,
			RangeMin: d.RangeMin,
			RangeMax: d.RangeMax,
		})
	}

	return loanOrders, nil
}

type queryParam struct {
	key   string
	value string
}

func (c *client) publicCall(command string, dest interface{}, queryParams ...queryParam) error {
	req, err := http.NewRequest("GET", PublicAPI, nil)
	if err != nil {
		logrus.WithError(err).Info("unable to create GET request")

		return err
	}

	query := req.URL.Query()
	query.Set("command", string(command))

	for _, q := range queryParams {
		query.Set(q.key, q.value)
	}

	req.URL.RawQuery = query.Encode()

	return c.processRequest(req, dest)
}

func (c *client) GetBalances() ([]*Balance, error) {
	balancesFromJSON := make(map[string]string)

	if err := c.tradeCall("returnBalances", &balancesFromJSON); err != nil {
		return nil, err
	}

	balances := []*Balance{}
	for k, v := range balancesFromJSON {
		amount, err := strconv.ParseFloat(v, 64)
		if err != nil {
			continue
		}

		balances = append(balances, &Balance{
			Currency: k,
			Amount:   amount,
		})
	}

	return balances, nil
}

// As the response from Poloniex for returnCompleteBalances contains some strings that should be floats,
// I use this struct only to unmarshal the result,
// then a conversion is made to return a clean Balance structure with floats where needed.
type balanceFromJSON struct {
	Available string `json:"available"`
	OnOrders  string `json:"onOrders"`
	BTCValue  string `json:"btcValue"`
}

func (c *client) GetCompleteBalances(account BalanceAccount) ([]*CompleteBalance, error) {
	balancesFromJSON := make(map[string]*balanceFromJSON)

	params := []postParam{}
	if account != "" {
		params = append(params, postParam{
			key:   "account",
			value: fmt.Sprintf("%v", account),
		})
	}

	if err := c.tradeCall("returnCompleteBalances", &balancesFromJSON, params...); err != nil {
		return nil, err
	}

	completeBalances := []*CompleteBalance{}
	for k, v := range balancesFromJSON {
		a, err := strconv.ParseFloat(v.Available, 64)
		if err != nil {
			continue
		}

		o, err := strconv.ParseFloat(v.OnOrders, 64)
		if err != nil {
			continue
		}

		b, err := strconv.ParseFloat(v.BTCValue, 64)
		if err != nil {
			continue
		}

		completeBalances = append(completeBalances, &CompleteBalance{
			Currency:  k,
			Available: a,
			OnOrders:  o,
			BTCValue:  b,
		})
	}

	return completeBalances, nil
}

type postParam struct {
	key   string
	value string
}

func (c *client) tradeCall(command string, dest interface{}, postParams ...postParam) error {
	form := url.Values{}
	form.Set("command", string(command))
	form.Set("nonce", fmt.Sprintf("%d", time.Now().UnixNano()))

	for _, p := range postParams {
		form.Set(p.key, p.value)
	}

	body := form.Encode()
	req, err := http.NewRequest("POST", TradeAPI, strings.NewReader(body))
	if err != nil {
		logrus.WithError(err).Info("unable to create POST request")

		return err
	}

	signedBody, err := helper.HmacSha512(c.secret, body)
	if err != nil {
		logrus.WithError(err).Error("unable to create hash from secret and request body")

		return err
	}

	// TODO see if content type is right.
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Key", c.key)
	req.Header.Set("Sign", signedBody)

	return c.processRequest(req, dest)
}

func (c *client) processRequest(r *http.Request, dest interface{}) error {
	resp, err := http.DefaultClient.Do(r)
	if err != nil {
		logrus.WithError(err).Info("unable to process request")

		return err
	}
	defer resp.Body.Close()

	if err := json.NewDecoder(resp.Body).Decode(dest); err != nil {
		// TODO rethink that.
		// When Poloniex API returns an error, the JSON is a map of string with an "error" key.
		// We'll try to decode it as a map of string then.
		// _, e := c.handleReturnedError(resp.Body)
		// if e != nil {
		// 	logrus.WithError(err).Error("unable to decode JSON")

		// 	return e
		// }

		// TODO we should return errorMap here, + an error?
		logrus.WithError(err).Error("error decoder")
		return errors.New("find the poloniex error in the returned map")
	}

	return nil
}

func (c *client) handleReturnedError(body io.ReadCloser) (map[string]string, error) {
	errorMap := map[string]string{}
	fmt.Println(body)
	if err := json.NewDecoder(body).Decode(errorMap); err != nil {
		logrus.WithError(err).Info("unable to decode body as an error")

		return nil, err
	}

	return errorMap, nil
}
