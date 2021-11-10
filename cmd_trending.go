package main

import (
	"encoding/json"
	"fmt"
	"github.com/olekukonko/tablewriter"
	"github.com/urfave/cli/v2"
	"log"
	"os"
	"time"
)

func cmdTrending() *cli.Command {
	return &cli.Command{
		Name:  "trending",
		Usage: "Top trending on wsb",
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:    "watch",
				Aliases: []string{"w"},
				Value:   false,
				Usage:   "watch mode",
			},
			&cli.IntFlag{
				Name:    "number",
				Aliases: []string{"n"},
				Value:   10,
				Usage:   "max number of trending tickers",
			},
		},
		Action: func(c *cli.Context) error {
			number := c.Int("number")
			if number > 20 {
				number = 20
			}
			trending(c.Bool("watch"), number)
			return nil
		},
	}
}

func trending(watch bool, max int) {
	if watch {
		ticker := time.NewTicker(60 * time.Second)
		for ; true; <-ticker.C {
			fmt.Print("\033[H\033[2J")
			printTrending(max)
		}
	}

	printTrending(max)
}

func printTrending(max int) {
	url := "https://api.wsb-tracker.com/data/symbols-overview"
	r, err := client.Get(url)

	if err != nil {
		log.Fatal("API request failed: ", err)
	}

	defer r.Body.Close()
	list := [][]string{}

	err = json.NewDecoder(r.Body).Decode(&list)

	if err != nil {
		log.Fatal("Could not decode API response: ", err)
	}

	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"Symbol", "Name", "Price", "Pct"})

	for _, tuple := range list[0:max] {
		q, _ := getQuote(tuple[1], false)
		row := []string{
			tuple[1],
			tuple[0],
			formatPrice(q.Price),
			fmt.Sprintf("%.2f", q.Pct),
		}

		color := tablewriter.FgGreenColor
		if q.Pct < 0 {
			color = tablewriter.FgRedColor
		}

		table.Rich(row, []tablewriter.Colors{
			{},
			{},
			{},
			{tablewriter.Bold, color},
		})
	}

	fmt.Println("")
	table.Render()
}
