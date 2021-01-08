package main

import (
  "log"
  "fmt"
  "time"
  "github.com/urfave/cli/v2"
  "github.com/guptarohit/asciigraph"
  "github.com/piquette/finance-go/chart"
  "github.com/piquette/finance-go/datetime"
)

func CmdGraph() *cli.Command {
  return &cli.Command{
    Name: "graph",
    Usage: "Print graph",
    Flags: []cli.Flag {
      &cli.StringFlag{
        Name: "period",
        Aliases: []string{"p"},
        Value: "",
        Usage: "Default is 1mo. Allowed: 1d,1wk,6mo,1yr,5yr",
      },
    },
    Action:  func(c *cli.Context) error {
      if c.NArg() > 0 {
        graph(c.Args().Get(0), c.String("period"))
        return nil
      }

      return fmt.Errorf("No symbol passed to command")
    },
  }
}

func graph(symbol string, period string) {
  var start int
  var interval datetime.Interval
  end := int(time.Now().Unix())

  switch period {
  case "1d":
    end = 0
    interval = datetime.FiveMins
  case "1wk":
    start = int(time.Now().Add(-7 * 24 * time.Hour).Unix())
    interval = datetime.OneHour
  case "2wk":
    start = int(time.Now().Add(-14 * 24 * time.Hour).Unix())
    interval = datetime.OneHour
  case "6mo":
    start = int(time.Now().Add(-6 * 30 * 24 * time.Hour).Unix())
    interval = datetime.FiveDay
  case "1yr":
    start = int(time.Now().Add(-12 * 30 * 24 * time.Hour).Unix())
    interval = datetime.FiveDay
  case "5yr":
    start = int(time.Now().Add(-5 * 12 * 30 * 24 * time.Hour).Unix())
    interval = datetime.OneMonth
  default:
    start = int(time.Now().Add(-30 * 24 * time.Hour).Unix())
    interval = datetime.OneDay
  }

  // fetch chart bars.
  params := &chart.Params{
    Symbol:   symbol,
    Interval: interval,
  }
  if start != 0 {
    params.Start = datetime.FromUnix(start)
  }
  if end != 0 {
    params.End = datetime.FromUnix(end)
  }

  iter := chart.Get(params)

  data := []float64{}
  for iter.Next() {
    b := iter.Bar()
    fl, _ := b.Close.Round(2).Float64()
    data = append(data, fl)
  }
  if iter.Err() != nil {
    log.Fatal(iter.Err())
  }

  // Append current (pre/post market) price
  q, err := getQuote(symbol, true)
  if err == nil {
    data = append(data, q.Price)
  }

  if len(data) == 0 {
    log.Fatal("Could not find chart data")
  }

  graph := asciigraph.Plot(data, asciigraph.Height(16))
  fmt.Println(graph)
}
