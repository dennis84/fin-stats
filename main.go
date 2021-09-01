package main

import (
	"fmt"
	"github.com/piquette/finance-go/quote"
	"github.com/urfave/cli/v2"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/user"
	"path/filepath"
	"time"
)

// Order ...
type Order struct {
	Units    float64
	In       float64
	Currency string
}

// Conf ...
type Conf struct {
	Currency    string
	Savings     map[string]float64
	Investments struct {
		Assets map[string][]Order
		Stocks map[string][]Order
		Crypto map[string][]Order
	}
	Income   map[string]float64
	Expenses map[string]float64
}

// InvestmentStats ...
type InvestmentStats struct {
	Sum     float64
	In      float64
	Diff    float64
	Loss    float64
	Details map[string][]InvestmentDetail `yaml:"-"`
}

// InvestmentDetail ...
type InvestmentDetail struct {
	Units float64
	Sum   float64
	In    float64
	Diff  float64
	Quote Quote
}

// Quote ...
type Quote struct {
	Price      float64
	Pct        float64
	Symbol     string
	State      string
	Name       string
	MarketInfo MarketInfo
}

var client = &http.Client{Timeout: 10 * time.Second}

func main() {
	app := &cli.App{
		Name:  "fin-stats",
		Usage: "",
		Commands: []*cli.Command{
			cmdSum(),
			cmdQuote(),
			cmdGraph(),
			cmdPortfolio(),
			cmdTrending(),
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal("Error: ", err)
	}
}

func getQuote(symbol string, fail bool) (Quote, error) {
	result := Quote{}
	q, err := quote.Get(symbol)
	if err != nil {
		return result, err
	}

	if q == nil {
		return result, fmt.Errorf("Could not find quote by symbol %s", symbol)
	}

	result.Price = q.RegularMarketPrice
	result.Pct = q.RegularMarketChangePercent

	if q.MarketState == "PRE" && q.PreMarketPrice > 0 {
		result.Price = q.PreMarketPrice
		result.Pct = q.PreMarketChangePercent
	} else if q.MarketState == "POST" && q.PostMarketPrice > 0 {
		result.Price = q.PostMarketPrice
		result.Pct = q.PostMarketChangePercent
	}

	result.Symbol = q.Symbol
	result.State = string(q.MarketState)
	result.Name = q.ShortName
	result.MarketInfo = getMarketInfo(*q)

	return result, nil
}

func getCurrency(name string) float64 {
	if name == "EUR" {
		q, _ := getQuote("EUR=X", true)
		return q.Price
	}

	return 1.0
}

func getInvestmentsStats(investments map[string][]Order) InvestmentStats {
	sum := 0.0
	sumIn := 0.0
	loss := 0.0
	details := make(map[string][]InvestmentDetail)

	for symbol, orders := range investments {
		for _, order := range orders {
			quote, _ := getQuote(symbol, true)
			price := quote.Price
			in := order.In

			if order.Currency != "" && order.Currency != "USD" {
				factor := getCurrency(order.Currency)
				price = price / factor
				in = in / factor
			}

			total := price * order.Units
			totalIn := in * order.Units
			sum = sum + total
			sumIn = sumIn + totalIn
			detail := InvestmentDetail{
				Units: order.Units,
				In:    totalIn,
				Sum:   total,
				Diff:  total - totalIn,
				Quote: quote,
			}

			details[symbol] = append(details[symbol], detail)
			if totalIn > total {
				loss = loss + totalIn - total
			}
		}
	}

	return InvestmentStats{sum, sumIn, sum - sumIn, loss, details}
}

func getHomeDir() string {
	usr, err := user.Current()
	if err != nil {
		fmt.Println(err)
	}

	return usr.HomeDir
}

func getOutDir(configFile string) (string, error) {
	dir, err := filepath.Abs(filepath.Dir(configFile))
	if err != nil {
		return "", err
	}

	dataDir := dir + "/finances"
	_, err = os.Stat(dataDir)
	if err != nil {
		return "", fmt.Errorf("Output dir does not exist: %s", dataDir)
	}

	return dataDir, nil
}

func findConfigFile(file string) (string, error) {
	usr, err := user.Current()
	if err != nil {
		return "", err
	}

	if file != "" {
		_, err = os.Stat(file)
		if err != nil {
			return "", fmt.Errorf("File does not exists: %s", file)
		}

		return file, nil
	}

	filename := "finances.yaml"
	_, err = os.Stat(usr.HomeDir + "/" + filename)
	if err == nil {
		return usr.HomeDir + "/" + filename, nil
	}

	return filename, nil
}

func readYaml(filename string, in interface{}) error {
	buf, err := ioutil.ReadFile(filename)
	if err != nil {
		return err
	}

	err = yaml.Unmarshal(buf, in)
	if err != nil {
		return fmt.Errorf("in file %q: %v", filename, err)
	}

	return nil
}

func yamlToBytes(in interface{}) []byte {
	d, err := yaml.Marshal(&in)
	if err != nil {
		fmt.Println(err)
	}

	return d
}
