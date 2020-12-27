package main

import (
  "os"
  "path/filepath"
  "os/user"
  "fmt"
  "time"
  "net/http"
  "io/ioutil"
  "gopkg.in/yaml.v2"
  "github.com/piquette/finance-go/quote"
  "github.com/urfave/cli/v2"
)

type Order struct {
  Units float64
  In float64
  Currency string
}

type Conf struct {
  Currency string
  Savings map[string]float64
  Investments struct {
    Assets map[string][]Order
    Stocks map[string][]Order
    Crypto map[string][]Order
  }
  Income map[string]float64
  Expenses map[string]float64
}

type InvestmentStats struct {
  Sum float64
  In float64
  Diff float64
  Loss float64
  Details map[string][]InvestmentDetail `yaml:"-"`
}

type InvestmentDetail struct {
  Units float64
  Sum float64
  In float64
  Diff float64
  Quote Quote
}

type Quote struct {
  Price float64
  Pct float64
  Symbol string
  State string
  MarketInfo MarketInfo
}

var client = &http.Client{Timeout: 10 * time.Second}

func main() {
  app := &cli.App{
    Name: "fin-stats",
    Usage: "",
    Commands: []*cli.Command{
      CmdSum(),
      CmdQuote(),
      CmdPortfolio(),
      CmdMentions(),
    },
  }

  err := app.Run(os.Args)
  if err != nil {
    fmt.Println(err)
    os.Exit(1)
  }
}

func getQuote(symbol string, fail bool) Quote {
  q, err := quote.Get(symbol)
  if err != nil {
    fmt.Println(err)
  }

  price := q.RegularMarketPrice
  pct := q.RegularMarketChangePercent
  if q.MarketState == "PRE" && q.PreMarketPrice > 0 {
    price = q.PreMarketPrice
    pct = q.PreMarketChangePercent
  } else if q.MarketState == "POST" && q.PostMarketPrice > 0 {
    price = q.PostMarketPrice
    pct = q.PostMarketChangePercent
  }

  if price == 0 && fail {
    fmt.Println("Quote price is 0, used wrong symbol?", symbol)
    os.Exit(1)
  }

  quote := Quote{
    price,
    pct,
    q.Symbol,
    string(q.MarketState),
    getMarketInfo(*q),
  }
  return quote
}

func getCurrency(name string) float64 {
  if name == "EUR" {
    return getQuote("EUR=X", true).Price
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
      quote := getQuote(symbol, true)
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
        In: totalIn,
        Sum: total,
        Diff: total - totalIn,
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

func findConfigFile(file string) string {
  usr, err := user.Current()
  if err != nil {
    fmt.Println(err)
  }

  if file != "" {
    _, err = os.Stat(file)
    if err != nil {
      fmt.Println("File does not exists:", file)
      os.Exit(1)
    }

    return file
  }

  filename := "finances.yaml"
  _, err = os.Stat(usr.HomeDir + "/" + filename)
  if err == nil {
    return usr.HomeDir + "/" + filename
  }

  return filename
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
