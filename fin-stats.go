package main

import (
  "bytes"
  "os"
  "path/filepath"
  "os/user"
  "fmt"
  "time"
  "io/ioutil"
  "flag"
  "gopkg.in/yaml.v2"
  "github.com/piquette/finance-go/quote"
)

type Order struct {
  Units float64
  In float64
  Currency string
}

type conf struct {
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
}

type Out struct {
  Date time.Time
  Savings float64
  Stocks InvestmentStats `yaml:"stocks,omitempty"`
  Assets InvestmentStats `yaml:"assets,omitempty"`
  Crypto InvestmentStats `yaml:"crypto,omitempty"`
  InvestmentsSum float64 `yaml:"investments_sum"`
  Total float64

  Income float64
  Expenses float64
  Budget float64
}

type DetailsOut struct {
  Details map[string][]InvestmentDetail
}

var quoteCache = make(map[string]Quote)

func main() {
  filenamePtr := flag.String("f", "", "filename")
  detailsPtr := flag.Bool("d", false, "details")
  flag.Parse()

  start := time.Now()
  filename := findConfigFile(*filenamePtr)
  var buf bytes.Buffer

  c, err := readConf(filename)
  if err != nil {
    fmt.Println(err)
    os.Exit(1)
  }

  savings := 0.0
  income := 0.0
  expenses := 0.0
  currencyFactor := getCurrency(c.Currency)
  stockStats := getInvestmentsStats(c.Investments.Stocks)
  assetsStats := getInvestmentsStats(c.Investments.Assets)
  cryptoStats := getInvestmentsStats(c.Investments.Crypto)
  investmentsSum := assetsStats.Sum + stockStats.Sum + cryptoStats.Sum

  for _, value := range c.Savings {
    savings = savings + value
  }

  for _, value := range c.Income {
    income = income + value
  }

  for _, value := range c.Expenses {
    expenses = expenses + value
  }

  out := Out{
    Date: start,
    Savings: savings,
    Stocks: stockStats,
    Assets: assetsStats,
    Crypto: cryptoStats,
    InvestmentsSum: investmentsSum,
    Total: savings + (investmentsSum * currencyFactor),
    Income: income,
    Expenses: expenses,
    Budget: income - expenses,
  }

  prettyPrint(&buf, out, *detailsPtr)
  writeFile(buf, filename, start)
}

func writeFile(buf bytes.Buffer, filename string, date time.Time) {
  dir, err := filepath.Abs(filepath.Dir(filename))
  if err != nil {
    fmt.Println(err)
  }

  dataDir := dir + "/finances"
  _, err = os.Stat(dataDir)
  if err != nil {
    fmt.Println("Stats could not be saved in:", dataDir)
    os.Exit(0)
  }

  target := dataDir + "/" + date.Format(time.RFC3339) + ".yaml"
  file, err := os.Create(target)

  if err != nil {
    fmt.Println("Could not create file:", filename)
    os.Exit(1)
  }

  file.Write(buf.Bytes())
  file.Close()
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

func readConf(filename string) (*conf, error) {
  buf, err := ioutil.ReadFile(filename)
  if err != nil {
    return nil, err
  }

  c := &conf{}
  err = yaml.Unmarshal(buf, c)
  if err != nil {
    return nil, fmt.Errorf("in file %q: %v", filename, err)
  }

  return c, nil
}

func getQuote(symbol string) Quote {
  q, err := quote.Get(symbol)
  if err != nil {
    fmt.Println(err)
  }

  if value, ok := quoteCache[symbol]; ok {
    return value
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

  quote := Quote{price, pct}
  quoteCache[symbol] = quote
  return quote
}

func getCurrency(name string) float64 {
  if name == "EUR" {
    return getQuote("EUR=X").Price
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
      quote := getQuote(symbol)
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

  return InvestmentStats{sum, sumIn, loss, details}
}

func prettyPrint(buf *bytes.Buffer, out Out, details bool) {
  d := toBytes(&out)
  fmt.Fprintln(buf, string(d))

  if details {
    stocks := toBytes(&DetailsOut{out.Stocks.Details})
    assets := toBytes(&DetailsOut{out.Assets.Details})
    crypto := toBytes(&DetailsOut{out.Crypto.Details})
    if out.Stocks.Sum > 0 {
      fmt.Fprintln(buf, "stock_" + string(stocks))
    }
    if out.Assets.Sum > 0 {
      fmt.Fprintln(buf, "asset_" + string(assets))
    }
    if out.Crypto.Sum > 0 {
      fmt.Fprintln(buf, "crypto_" + string(crypto))
    }
  }

  fmt.Println(buf.String())
}

func toBytes(in interface{}) []byte {
  d, err := yaml.Marshal(&in)
  if err != nil {
    fmt.Println(err)
  }

  return d
}

func getHomeDir() string {
  usr, err := user.Current()
  if err != nil {
    fmt.Println(err)
  }

  return usr.HomeDir
}
