package main

import (
  "os"
  "fmt"
  "time"
  "github.com/urfave/cli/v2"
  "github.com/guptarohit/asciigraph"
  "github.com/olekukonko/tablewriter"
)

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
      symbol := ""
      if c.NArg() > 0 {
        symbol = c.Args().Get(0)
      }

      quoteInfo(symbol, c.Bool("watch"))
      return nil
    },
  }
}

func printQuote(q Quote) {
  data := [][]string{
    []string{"Symbol", q.Symbol},
    []string{"Price", fmt.Sprintf("%.2f", q.Price)},
    []string{"Pct", fmt.Sprintf("%.2f", q.Pct)},
  }

  table := tablewriter.NewWriter(os.Stdout)

  for _, v := range data {
    table.Append(v)
  }

  table.Render()
}

func quoteInfo(symbol string, watch bool) {
  if watch {
    quotes := []Quote{}
    ticker := time.NewTicker(2 * time.Second)
    for; true; <-ticker.C {
      q := getQuote(symbol, true)
      quotes = append(quotes, q)

      fmt.Print("\033[H\033[2J")
      printQuote(q)

      data := []float64{}

      for _, quote := range quotes {
        data = append(data, quote.Price)
      }

      graph := asciigraph.Plot(data, asciigraph.Height(16))
      if len(data) > 80 {
        graph = asciigraph.Plot(data, asciigraph.Height(8), asciigraph.Width(80))
      }

      fmt.Println("")
      fmt.Println(graph)
    }
  }

  q := getQuote(symbol, true)
  printQuote(q)
}

