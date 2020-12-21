package main

import (
  "bytes"
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
var blue = color.New(color.FgBlue).Add(color.Bold)
var green = color.New(color.FgGreen).Add(color.Bold)
var red = color.New(color.FgRed).Add(color.Bold)

func main() {
  filenamePtr := flag.String("f", "", "filename")
  detailsPtr := flag.Bool("d", false, "details")
  outPtr := flag.String("o", "", "output dir")
  flag.Parse()

  start := time.Now()
  filename := findConfigFile(*filenamePtr)
  var out bytes.Buffer

  if *outPtr != "" {
    blue.DisableColor()
    green.DisableColor()
    red.DisableColor()
  }

  out.WriteString("finance_stats:\n")
  out.WriteString("  file:           " + filename + "\n")
  out.WriteString("  date:           " + start.Format(time.RFC3339) + "\n")

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

  if *detailsPtr {
    printInvestmentDetails(&out, "stocks", stockStats)
    printInvestmentDetails(&out, "assets", assetsStats)
    printInvestmentDetails(&out, "crypto", cryptoStats)
  }

  fmt.Fprintf(&out, "savings:          %.2f\n", savings)
  printInvestmentStats(&out, "stocks", stockStats, currencyFactor)
  printInvestmentStats(&out, "assets", assetsStats, currencyFactor)
  printInvestmentStats(&out, "crypto", cryptoStats, currencyFactor)

  blue.Fprintf(&out, "total:            %.2f\n", savings + (investmentsSum * currencyFactor))
  fmt.Fprintf(&out, "income:           %.2f\n", income)
  fmt.Fprintf(&out, "expenses:         %.2f\n", expenses)
  fmt.Fprintf(&out, "budget:           %.2f\n", income - expenses)

  fmt.Println(out.String())

  if *outPtr != "" {
    writeFile(out, *outPtr, start)
  }
}

func writeFile(out bytes.Buffer, dir string, date time.Time) {
  filename := dir + "/" + date.Format(time.RFC3339) + ".yaml"
  file, err := os.Create(filename)

  if err != nil {
    fmt.Println("Could not create file:", filename)
    os.Exit(1)
  }

  file.Write(out.Bytes())
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

func printInvestmentDetails(out *bytes.Buffer, name string, stats InvestmentStats) {
  if len(stats.Details) > 0 {
    fmt.Fprintln(out, name + "_details:")

    for symbol, investments := range stats.Details {
      for _, detail := range investments {
        fmt.Fprintln(out, "  - symbol:      ", symbol)
        fmt.Fprintln(out, "    units:       ", detail.Units)
        fmt.Fprintf(out, "    in:           %.2f\n", detail.In)
        fmt.Fprintf(out, "    current:      %.2f\n", detail.Sum)
        if detail.In > detail.Sum {
          red.Fprintf(out, "    loss:         %.2f\n", detail.In - detail.Sum)
        } else {
          green.Fprintf(out, "    profit:       %.2f\n", detail.Sum - detail.In)
        }
        fmt.Fprintf(out, "    unit_price:   %.2f\n", detail.Quote.Price)
        fmt.Fprintf(out, "    today_change: %.2f\n", detail.Quote.Pct)
      }
    }
  }
}

func printInvestmentStats(out *bytes.Buffer,
                          name string,
                          stats InvestmentStats,
                          currencyFactor float64) {
  if stats.Sum != 0 {
    diff := stats.Sum - stats.SumIn
    c := green
    if diff < 0 {
      c = red
    }

    fmt.Fprintln(out, name + ":")
    fmt.Fprintf(out, "  current:        %.2f\n", stats.Sum * currencyFactor)
    c.Fprintf(out, "  diff:           %.2f\n", diff * currencyFactor)
    fmt.Fprintf(out, "  losses:         %.2f\n", stats.Loss * currencyFactor)
  }
}
