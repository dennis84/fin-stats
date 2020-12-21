package main

import (
  "os"
  "os/user"
  "fmt"
  "time"
  "io/ioutil"
  "flag"
  "gopkg.in/yaml.v2"
  "github.com/piquette/finance-go/quote"
  "github.com/fatih/color"
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
  SumIn float64
  Loss float64
  Details map[string][]InvestmentDetail
}

type InvestmentDetail struct {
  Units float64
  Sum float64
  In float64
  Quote Quote
}

type Quote struct {
  Price float64
  Pct float64
}

var quoteCache = make(map[string]Quote)
var highlight = color.New(color.FgBlue).Add(color.Bold)

func main() {
  filenamePtr := flag.String("f", "", "filename")
  detailsPtr := flag.Bool("d", false, "details")
  flag.Parse()

  start := time.Now()
  filename := findConfigFile(*filenamePtr)
  fmt.Println("finance_stats:")
  fmt.Println("  file:          ", filename)
  fmt.Println("  date:          ", start.Format(time.RFC3339))

  c, err := readConf(filename)
  if err != nil {
    fmt.Println("Err", err)
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

  if *detailsPtr {
    printInvestmentDetails("stocks", stockStats)
    printInvestmentDetails("assets", assetsStats)
    printInvestmentDetails("crypto", cryptoStats)
  }

  fmt.Printf("savings:          %.2f\n", savings)
  printInvestmentStats("stocks", stockStats, currencyFactor)
  printInvestmentStats("assets", assetsStats, currencyFactor)
  printInvestmentStats("crypto", cryptoStats, currencyFactor)

  highlight.Printf("total:            %.2f\n", savings + (investmentsSum * currencyFactor))
  fmt.Printf("income:           %.2f\n", income)
  fmt.Printf("expenses:         %.2f\n", expenses)
  fmt.Printf("budget:           %.2f\n", income - expenses)
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

func printInvestmentDetails(name string, stats InvestmentStats) {
  if len(stats.Details) > 0 {
    fmt.Println(name + "_details:")

    for symbol, investments := range stats.Details {
      for _, detail := range investments {
        fmt.Println("  - symbol:      ", symbol)
        fmt.Println("    units:       ", detail.Units)
        fmt.Printf("    in:           %.2f\n", detail.In)
        highlight.Printf("    current:      %.2f\n", detail.Sum)
        if detail.In > detail.Sum {
          color.Set(color.FgRed)
          fmt.Printf("    loss:         %.2f\n", detail.In - detail.Sum)
          color.Unset()
        } else {
          color.Set(color.FgGreen)
          fmt.Printf("    profit:       %.2f\n", detail.Sum - detail.In)
          color.Unset()
        }
        fmt.Printf("    unit_price:   %.2f\n", detail.Quote.Price)
        fmt.Printf("    today_change: %.2f\n", detail.Quote.Pct)
      }
    }
  }
}

func printInvestmentStats(name string, stats InvestmentStats, currencyFactor float64) {
  if stats.Sum != 0 {
    diff := stats.Sum - stats.SumIn
    fmt.Println(name + ":")
    fmt.Printf("  current:        %.2f\n", stats.Sum * currencyFactor)
    if diff >= 0 {
      color.Set(color.FgGreen)
    } else {
      color.Set(color.FgRed)
    }
    fmt.Printf("  diff:           %.2f\n", diff * currencyFactor)
    if stats.Loss > 0 {
      color.Set(color.FgRed)
    } else {
      color.Set(color.FgGreen)
    }
    fmt.Printf("  losses:         %.2f\n", stats.Loss * currencyFactor)
    color.Unset()
  }
}
