package poloniex

// Poloniex endpoints.
const (
	URL       = "https://poloniex.com"
	TradeAPI  = URL + "/tradingApi"
	PublicAPI = URL + "/public"
)

// Poloniex is the public interface to interact with Poloniex API.
type Poloniex interface {
	//////////////////
	// Public calls //
	//////////////////

	// Returns the ticker for all markets.
	// Ex: map["BTC_EXP"] == {ID: 123, Last:"456", etc.}
	GetTickers() ([]*Ticker, error)

	// Returns the 24-hour volume for all markets, plus totals for primary currencies.
	Get24hVolume() (*Volume24h, error)

	// Returns the order book for a given market, as well as a sequence number for use with the Push API,
	// and an indicator specifying whether the market is frozen.
	GetOrderBook(currencyPair string, depth uint) (*OrderBook, error)

	// Returns the order book of all markets, as well as a sequence number for use with the Push API,
	// and an indicator specifying whether the market is frozen.
	GetAllOrderBooks(depth uint) ([]*OrderBook, error)

	// Returns the past 200 trades for a given market, or up to 50,000 trades between a range specified in UNIX timestamps,
	// by the "start" and "end" GET parameters.
	// If start or end are set to 0, it returns the past 200 trades.
	GetTradeHistory(currencyPair string, start, end uint64) ([]*TradeHistory, error)

	// Returns candlestick chart data. Required GET parameters are "currencyPair",
	// "period" (candlestick period in seconds; valid values are 300, 900, 1800, 7200, 14400, and 86400),
	// "start", and "end". "Start" and "end" are given in UNIX timestamp format and used to specify the date range for the data returned.
	GetChartData(currencyPair string, start, end uint64, period ChartDataPeriod) ([]*ChartData, error)

	// Returns information about currencies.
	GetCurrencies() ([]*Currency, error)

	// Returns the list of loan offers and demands for a given currency, specified by the "currency" GET parameter.
	GetLoanOrders(currency string) (*LoanOrders, error)

	///////////////////
	// Private calls //
	///////////////////

	// Returns all of your available balances.
	GetBalances() ([]*Balance, error)

	// Returns all of your balances, including available balance, balance on orders, and the estimated BTC value of your balance.
	GetCompleteBalances(account BalanceAccount) ([]*CompleteBalance, error)
}

type Ticker struct {
	Currency      string
	ID            int64  `json:"id"`
	Last          string `json:"last"`
	LowestAsk     string `json:"lowestAsk"`
	HighestBid    string `json:"highestBid"`
	PercentChange string `json:"percentChange"`
	BaseVolume    string `json:"baseVolume"`
	QuoteVolume   string `json:"quoteVolume"`
	IsFrozen      string `json:"isFrozen"`
	High24hr      string `json:"high24hr"`
	Low24hr       string `json:"low24hr"`
}

// Volume24h represents the response of the Return24hVolume API call.
type Volume24h struct {
	// Contains total volume for some primary currencies.
	// Example: map["totalBTC"] = "7364.48394883"
	// Possible keys:
	// - totalBTC
	// - totalETH
	// - totalUSDT
	// - totalXMR
	// - totalXUSD
	PrimaryCurrenciesTotals map[string]string

	// This map contains volumes by market.
	// Example: For the market BTC_LTC, it gives:
	// map["BTC_LTC"]["BTC"] == "2.23248854".
	// map["BTC_LTC"]["LTC"] == "87.10381314".
	Markets map[string]map[string]string
}

type Order struct {
	Value  float64
	Amount float64
}

type OrderBook struct {
	Pair     string
	Asks     []*Order
	Bids     []*Order
	IsFrozen string
	Seq      int64
}

type TradeHistory struct {
	GlobalTradeID int64
	TradeID       int64
	Date          string
	Type          string
	Rate          float64
	Amount        float64
	Total         float64
}

type ChartDataPeriod int64

const (
	Period300   = 300
	Period900   = 900
	Period1800  = 1800
	Period7200  = 7200
	Period14400 = 14400
	Period86400 = 86400
)

type ChartData struct {
	Date            int64   `json:"date"`
	High            float64 `json:"high"`
	Low             float64 `json:"low"`
	Open            float64 `json:"open"`
	Close           float64 `json:"close"`
	Volume          float64 `json:"volume"`
	QuoteVolume     float64 `json:"quoteVolume"`
	WeightedAverage float64 `json:"weightedAverage"`
}

type Currency struct {
	ID             int64  `json:"id"`
	Name           string `json:"name"`
	TxFee          string `json:"txFee"`
	MinConf        int    `json:"minConf"`
	DepositAddress string `json:"depositAddress"`
	Disabled       int    `json:"disabled"`
	Delisted       int    `json:"delisted"`
	Frozen         int    `json:"frozen"`
}

type Loan struct {
	Rate     float64
	Amount   float64
	RangeMin int64
	RangeMax int64
}

type LoanOrders struct {
	Offers  []*Loan
	Demands []*Loan
}

type Balance struct {
	Currency string
	Amount   float64
}

type BalanceAccount string

const (
	AllAccounts         BalanceAccount = "all"
	ExchangeAccountOnly BalanceAccount = ""
)

type CompleteBalance struct {
	Currency  string
	Available float64
	OnOrders  float64
	BTCValue  float64
}

// TradeCommand is an alias to string representing private calls to poloniex API.
// An authentication is required in order for these calls to work.
// type TradeCommand string

// // Possible TradeCommand values.
// const (
// 	ReturnBalances                 TradeCommand = "returnBalances"
// 	ReturnCompleteBalances         TradeCommand = "returnCompleteBalances"
// 	ReturnDepositAddresses         TradeCommand = "returnDepositAddresses"
// 	GenerateNewAddress             TradeCommand = "generateNewAddress"
// 	ReturnDepositsWithdrawals      TradeCommand = "returnDepositsWithdrawals"
// 	ReturnOpenOrders               TradeCommand = "returnOpenOrders"
// 	ReturnPrivateTradeHistory      TradeCommand = "returnTradeHistory"
// 	ReturnOrderTrades              TradeCommand = "returnOrderTrades"
// 	Buy                            TradeCommand = "buy"
// 	Sell                           TradeCommand = "sell"
// 	CancelOrder                    TradeCommand = "cancelOrder"
// 	MoveOrder                      TradeCommand = "moveOrder"
// 	Withdraw                       TradeCommand = "withdraw"
// 	ReturnFeelInfo                 TradeCommand = "returnFeelInfo"
// 	ReturnAvailableAccountBalances TradeCommand = "returnAvailableAccountBalances"
// 	ReturnTradableBalances         TradeCommand = "returnTradableBalancies"
// 	TransferBalance                TradeCommand = "transferBalance"
// 	ReturMarginAccountSummary      TradeCommand = "returnMarginAccountSummary"
// 	MarginBuy                      TradeCommand = "marginBuy"
// 	MarginSell                     TradeCommand = "marginSell"
// 	GetMarginPosition              TradeCommand = "getMarginPosition"
// 	CloseMarginPosition            TradeCommand = "closeMarginPosition"
// 	CreateLoanOffer                TradeCommand = "createLoanOffer"
// 	CancelLoanOffer                TradeCommand = "cancelLoanOffer"
// 	ReturnOpenLoanOffers           TradeCommand = "returnOpenLoanOffers"
// 	ReturnActiveLoans              TradeCommand = "returnActiveLoans"
// 	ReturnLendingHistory           TradeCommand = "returnLendingHistory"
// 	ToggleAutoRenew                TradeCommand = "toggleAutoRenew"
// )
