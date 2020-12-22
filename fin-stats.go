package main

import (
  "bytes"
  "os"
  "path/filepath"
  "os/user"
  "fmt"
  "time"
  "io/ioutil"
  "gopkg.in/yaml.v2"
  "github.com/piquette/finance-go/quote"
  "github.com/guptarohit/asciigraph"
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

type Options struct {
  File string `flag:"file f"`
  NoDetails bool
  NoSummary bool
  NoGraph bool
  Graph string
}

var quoteCache = make(map[string]Quote)

func main() {
  app := &cli.App{
    Name: "fin-stats",
    Usage: "",
    Commands: []*cli.Command{
      {
        Name: "sum",
        Usage: "Print stats",
        Flags: []cli.Flag {
          &cli.StringFlag{
            Name: "file",
            Aliases: []string{"f"},
            Value: "",
            Usage: "finance config",
          },
          &cli.BoolFlag{
            Name: "no-details",
            Value: false,
            Usage: "hide investment details",
          },
          &cli.BoolFlag{
            Name: "no-summary",
            Value: false,
            Usage: "hide summary details",
          },
          &cli.BoolFlag{
            Name: "no-graph",
            Value: false,
            Usage: "hide graph",
          },
          &cli.StringFlag{
            Name: "graph",
            Value: "total",
            Usage: "Graph value",
          },
        },
        Action:  func(c *cli.Context) error {
          options := Options{
            File: c.String("file"),
            NoDetails: c.Bool("no-details"),
            NoSummary: c.Bool("no-summary"),
            NoGraph: c.Bool("no-graph"),
            Graph: "total",
          }

          sum(options)
          return nil
        },
      },
      {
        Name: "quote",
        Usage: "Print quote",
        Action:  func(c *cli.Context) error {
          symbol := ""
          if c.NArg() > 0 {
            symbol = c.Args().Get(0)
          }

          quoteInfo(symbol)
          return nil
        },
      },
    },
  }

  err := app.Run(os.Args)
  if err != nil {
    fmt.Println(err)
    os.Exit(1)
  }
}

func quoteInfo(symbol string) {
  q := getQuote(symbol)
  d := toBytes(&q)
  fmt.Println("symbol:", symbol)
  fmt.Println(string(d))
}

func sum(options Options) {
  start := time.Now()
  filename := findConfigFile(options.File)
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

  prettyPrint(&buf, out, options)

  history := loadHistory(filename)
  writeFile(buf, filename, start)

  if !options.NoGraph && len(history) > 0 {
    data := []float64{}

    for _, stat := range history {
      data = append(data, stat.Total)
    }

    graph := asciigraph.Plot(data, asciigraph.Height(8))
    if len(data) > 80 {
      graph = asciigraph.Plot(data, asciigraph.Height(8), asciigraph.Width(80))
    }

    fmt.Println("Total:")
    fmt.Println(graph)
  }
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

  if price == 0 {
    fmt.Println("Quote price is 0, used wrong symbol?", symbol)
    os.Exit(1)
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

  return InvestmentStats{sum, sumIn, sum - sumIn, loss, details}
}

func prettyPrint(buf *bytes.Buffer, out Out, options Options) {
  d := toBytes(&out)

  if !options.NoSummary {
    fmt.Fprintln(buf, string(d))
  }

  if !options.NoDetails {
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

func loadHistory(filename string) []Out {
  dir, err := filepath.Abs(filepath.Dir(filename))
  if err != nil {
    fmt.Println(err)
  }

  outs := make([]Out, 0)
  dataDir := dir + "/finances"
  _, err = os.Stat(dataDir)
  if err != nil {
    return outs
  }

  files, err := ioutil.ReadDir(dataDir)
  if err != nil {
    return outs
  }

  for _, f := range files {
    path := dataDir + "/" + f.Name()
    out, err := readOut(path)
    if err != nil {
      fmt.Println("Could not read file from history:", path, err)
      continue
    }

    outs = append(outs, *out)
  }

  return outs
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

func readConf(filename string) (*Conf, error) {
  buf, err := ioutil.ReadFile(filename)
  if err != nil {
    return nil, err
  }

  c := &Conf{}
  err = yaml.Unmarshal(buf, c)
  if err != nil {
    return nil, fmt.Errorf("in file %q: %v", filename, err)
  }

  return c, nil
}

func readOut(filename string) (*Out, error) {
  buf, err := ioutil.ReadFile(filename)
  if err != nil {
    return nil, err
  }

  c := &Out{}
  err = yaml.Unmarshal(buf, c)
  if err != nil {
    return nil, fmt.Errorf("in file %q: %v", filename, err)
  }

  return c, nil
}

func getHomeDir() string {
  usr, err := user.Current()
  if err != nil {
    fmt.Println(err)
  }

  return usr.HomeDir
}
