package main

import (
  "os"
  "log"
  "fmt"
  "time"
  "io/ioutil"
  "github.com/urfave/cli/v2"
  "github.com/guptarohit/asciigraph"
  "github.com/olekukonko/tablewriter"
)

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

type Options struct {
  File string
  Watch bool
  NoSummary bool
  NoGraph bool
  Graph string
}

func CmdSum() *cli.Command {
  return &cli.Command{
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
        Name: "watch",
        Aliases: []string{"w"},
        Value: false,
        Usage: "watch mode",
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
        Watch: c.Bool("watch"),
        NoSummary: c.Bool("no-summary"),
        Graph: "total",
      }

      sum(options)
      return nil
    },
  }
}

func doSum(options Options) {
  start := time.Now()
  filename, err := findConfigFile(options.File)

  if err != nil {
    log.Fatal(err)
  }

  c := &Conf{}
  err = readYaml(filename, c)
  if err != nil {
    log.Fatal(err)
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

  if !options.NoSummary {
    printSumTable(out, options)
    fmt.Println("")
  }

  history := loadHistory(filename)
  if !options.NoSummary {
    writeFile(out, filename, start)
  }

  if !options.NoGraph && len(history) > 0 {
    data := []float64{}

    for _, stat := range append(history, out) {
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

func sum(options Options) {
  if options.Watch {
    ticker := time.NewTicker(10 * time.Second)
    for; true; <-ticker.C {
      fmt.Print("\033[H\033[2J")
      doSum(options)
    }
  }

  doSum(options)
}

func printSumTable(out Out, options Options) {
  data := [][]string{}

  data = [][]string{
    []string{"Savings", fmt.Sprintf("%.2f", out.Savings)},
  }

  if out.Stocks.Sum > 0 {
    data = append(data, [][]string{
      []string{"Stocks Sum", fmt.Sprintf("%.2f", out.Stocks.Sum)},
      []string{"Stocks In", fmt.Sprintf("%.2f", out.Stocks.In)},
      []string{"Stocks Diff", fmt.Sprintf("%.2f", out.Stocks.Diff)},
      []string{"Stocks Loss", fmt.Sprintf("%.2f", out.Stocks.Loss)},
    }...)
  }

  if out.Assets.Sum > 0 {
    data = append(data, [][]string{
      []string{"Assets Sum", fmt.Sprintf("%.2f", out.Assets.Sum)},
      []string{"Assets In", fmt.Sprintf("%.2f", out.Assets.In)},
      []string{"Assets Diff", fmt.Sprintf("%.2f", out.Assets.Diff)},
      []string{"Assets Loss", fmt.Sprintf("%.2f", out.Assets.Loss)},
    }...)
  }

  if out.Crypto.Sum > 0 {
    data = append(data, [][]string{
      []string{"Crypto Sum", fmt.Sprintf("%.2f", out.Crypto.Sum)},
      []string{"Crypto In", fmt.Sprintf("%.2f", out.Crypto.In)},
      []string{"Crypto Diff", fmt.Sprintf("%.2f", out.Crypto.Diff)},
      []string{"Crypto Loss", fmt.Sprintf("%.2f", out.Crypto.Loss)},
    }...)
  }

  data = append(data, [][]string{
    []string{"Investments Sum", fmt.Sprintf("%.2f", out.InvestmentsSum)},
    []string{"Total", fmt.Sprintf("%.2f", out.Total)},
  }...)

  table := tablewriter.NewWriter(os.Stdout)

  for _, v := range data {
    table.Append(v)
  }

  table.Render()
}

func writeFile(out Out, configFile string, date time.Time) {
  dir, err := getOutDir(configFile)
  if err != nil {
    log.Fatal(err)
  }

  target := dir + "/" + date.Format(time.RFC3339) + ".yaml"
  file, err := os.Create(target)

  if err != nil {
    log.Fatal("Could not create file:", dir)
  }

  file.Write(yamlToBytes(&out))
  file.Close()
}

func loadHistory(configFile string) []Out {
  outs := make([]Out, 0)
  dir, err := getOutDir(configFile)

  if err != nil {
    return outs
  }

  files, err := ioutil.ReadDir(dir)
  if err != nil {
    return outs
  }

  for _, f := range files {
    path := dir + "/" + f.Name()
    out := &Out{}
    err := readYaml(path, out)
    if err != nil {
      log.Println(err)
      continue
    }

    outs = append(outs, *out)
  }

  return outs
}
