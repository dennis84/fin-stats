package main

import (
  "os"
  "log"
  "fmt"
  "time"
  "encoding/json"
  "github.com/urfave/cli/v2"
  "github.com/olekukonko/tablewriter"
)

type Mention struct {
  Symbol string
  Mentions string
  Quote Quote `yaml:"quote,inline"`
}

type MentionsTable struct {
  DataValues []Mention `json:"data_values"`
}

type MentionsOut struct {
  Mentions map[string]Mention
}

func CmdMentions() *cli.Command {
  return &cli.Command{
    Name: "mentions",
    Usage: "Top mentions on wsb",
    Flags: []cli.Flag {
      &cli.BoolFlag{
        Name: "watch",
        Aliases: []string{"w"},
        Value: false,
        Usage: "watch mode",
      },
      &cli.IntFlag{
        Name: "number",
        Aliases: []string{"n"},
        Value: 10,
        Usage: "max number of mentions",
      },
    },
    Action:  func(c *cli.Context) error {
      mentions(c.Bool("watch"), c.Int("number"))
      return nil
    },
  }
}

func mentions(watch bool, max int) {
  if watch {
    ticker := time.NewTicker(60 * time.Second)
    for; true; <-ticker.C {
      fmt.Print("\033[H\033[2J")
      printMentions(max)
    }
  }

  printMentions(max)
}

func printMentions(max int) {
  url := "https://wsbsynth.com/ajax/get_table.php"
  r, err := client.Get(url)

  if err != nil {
    log.Fatal("API request failed: ", err)
  }

  defer r.Body.Close()
  m := &MentionsTable{}
  err = json.NewDecoder(r.Body).Decode(m)

  if err != nil {
    log.Fatal("Could not decode API response: ", err)
  }

  data := [][]string{}
  table := tablewriter.NewWriter(os.Stdout)
  table.SetHeader([]string{"Symbol", "Mentions", "Price", "Pct"})

  for _, mention := range m.DataValues[0:max] {
    q, _ := getQuote(mention.Symbol, false)
    row := []string{
      mention.Symbol,
      mention.Mentions,
      fmt.Sprintf("%.2f", q.Price),
      fmt.Sprintf("%.2f", q.Pct),
    }

    color := tablewriter.FgGreenColor
    if q.Pct < 0 {
      color = tablewriter.FgRedColor
    }

    table.Rich(row, []tablewriter.Colors{
      tablewriter.Colors{},
      tablewriter.Colors{},
      tablewriter.Colors{},
      tablewriter.Colors{tablewriter.Bold, color},
    })
  }

  for _, v := range data {
    table.Append(v)
  }

  fmt.Println("")
  table.Render()
}
