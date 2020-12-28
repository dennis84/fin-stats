package main

import (
  "math"
  "os"
  "fmt"
  "time"
  "strings"
  "github.com/urfave/cli/v2"
  "github.com/guptarohit/asciigraph"
  "github.com/olekukonko/tablewriter"
)

var graphData = []float64{}

func CmdQuote() *cli.Command {
  return &cli.Command{
    Name: "quote",
    Usage: "Print quote",
    Flags: []cli.Flag {
      &cli.BoolFlag{
        Name: "watch",
        Aliases: []string{"w"},
        Value: false,
        Usage: "watch mode",
      },
    },
    Action:  func(c *cli.Context) error {
      symbols := []string{}
      if c.NArg() > 0 {
        symbols = strings.Split(c.Args().Get(0), ",")
      }

      quoteInfo(symbols, c.Bool("watch"))
      return nil
    },
  }
}

func printQuotes(quotes []Quote) {
  table := tablewriter.NewWriter(os.Stdout)
  headers := []string{"Symbol", "Price", "Pct", "State", "Name", "Trading Hours"}
  data := [][]string{}

  for _, q := range quotes {
    color := tablewriter.FgGreenColor
    event := ""

    if (q.State == "CLOSED" || q.State == "PREPRE") && q.MarketInfo.DurationUntilOpenPre != nil {
      event = fmt.Sprintf("Pre market opens in %s", formatDuration(*q.MarketInfo.DurationUntilOpenPre))
    } else if (q.State == "CLOSED" || q.State == "PREPRE") && q.MarketInfo.DurationUntilOpen != nil {
      event = fmt.Sprintf("Market opens in %s", formatDuration(*q.MarketInfo.DurationUntilOpen))
    } else if q.State == "PRE" && q.MarketInfo.DurationUntilOpen != nil {
      event = fmt.Sprintf("Market opens in %s", formatDuration(*q.MarketInfo.DurationUntilOpen))
    } else if q.State == "OPEN" && q.MarketInfo.DurationUntilClose != nil {
      event = fmt.Sprintf("Market closes in %s", formatDuration(*q.MarketInfo.DurationUntilClose))
    } else if q.State == "POST" && q.MarketInfo.DurationUntilClosePost != nil {
      event = fmt.Sprintf("Post market closes in %s", formatDuration(*q.MarketInfo.DurationUntilClosePost))
    }

    row := []string{
      q.Symbol,
      fmt.Sprintf("%.2f", q.Price),
      fmt.Sprintf("%.2f", q.Pct),
      q.State,
      q.Name,
      event,
    }

    if q.Pct < 0 {
      color = tablewriter.FgRedColor
    }

    table.Rich(row, []tablewriter.Colors{
      tablewriter.Colors{},
      tablewriter.Colors{},
      tablewriter.Colors{tablewriter.Bold, color},
    })
  }


  table.SetHeader(headers)

  for _, v := range data {
    table.Append(v)
  }

  table.Render()
}

func printQuoteGraph(q Quote) {
  graphData = append(graphData, q.Price)
  graph := asciigraph.Plot(graphData, asciigraph.Height(16))
  if len(graphData) > 80 {
    graph = asciigraph.Plot(graphData, asciigraph.Height(8), asciigraph.Width(80))
  }

  fmt.Println("")
  fmt.Println(graph)
}

func quoteInfo(symbols []string, watch bool) {
  if watch {
    ticker := time.NewTicker(2 * time.Second)
    for; true; <-ticker.C {
      quotes := []Quote{}
      for _, symbol := range symbols {
        q := getQuote(symbol, true)
        quotes = append(quotes, q)
      }

      fmt.Print("\033[H\033[2J")
      printQuotes(quotes)
      if len(quotes) == 1 && quotes[0].State != "CLOSED" {
        printQuoteGraph(quotes[0])
      }
    }
  }

  quotes := []Quote{}
  for _, symbol := range symbols {
    q := getQuote(symbol, true)
    quotes = append(quotes, q)
  }

  printQuotes(quotes)
}

func formatDuration(d time.Duration) string {
  mins := math.Mod(d.Minutes(), 60)
  secs := math.Mod(d.Seconds(), 60)
  return fmt.Sprintf(
    "%d hours %d min %d sec",
    int64(d.Hours()),
    int64(mins),
    int64(secs),
  )
}
