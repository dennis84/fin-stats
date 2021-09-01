package main

import (
	"fmt"
	"github.com/olekukonko/tablewriter"
	"github.com/urfave/cli/v2"
	"log"
	"os"
	"sort"
	"time"
)

func cmdPortfolio() *cli.Command {
	return &cli.Command{
		Name:  "portfolio",
		Usage: "Print investment stats",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "file",
				Aliases: []string{"f"},
				Value:   "",
				Usage:   "finance config",
			},
			&cli.BoolFlag{
				Name:    "watch",
				Aliases: []string{"w"},
				Value:   false,
				Usage:   "watch mode",
			},
		},
		Action: func(c *cli.Context) error {
			portfolio(c.String("file"), c.Bool("watch"))
			return nil
		},
	}
}

func printInvestmentDetailsTable(details []InvestmentDetail) {
	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"Symbol", "Sum", "In", "Diff", "Quote Price", "Quote Pct"})

	for _, detail := range details {
		row := []string{
			detail.Quote.Symbol,
			fmt.Sprintf("%.2f", detail.Sum),
			fmt.Sprintf("%.2f", detail.In),
			fmt.Sprintf("%.2f", detail.Diff),
			fmt.Sprintf("%.2f", detail.Quote.Price),
			fmt.Sprintf("%.2f", detail.Quote.Pct),
		}

		color := tablewriter.FgGreenColor
		if detail.Quote.Pct < 0 {
			color = tablewriter.FgRedColor
		}

		table.Rich(row, []tablewriter.Colors{
			{},
			{},
			{},
			{},
			{},
			{tablewriter.Bold, color},
		})
	}

	table.Render()
}

func printPortiflio(file string, clear bool) {
	filename, err := findConfigFile(file)

	if err != nil {
		log.Fatal(err)
	}

	c := &Conf{}
	err = readYaml(filename, c)
	if err != nil {
		log.Fatal("Could not read config file: ", err)
	}

	details := []InvestmentDetail{}
	stockStats := getInvestmentsStats(c.Investments.Stocks)
	assetsStats := getInvestmentsStats(c.Investments.Assets)
	cryptoStats := getInvestmentsStats(c.Investments.Crypto)

	for _, values := range stockStats.Details {
		details = append(details, values...)
	}

	for _, values := range assetsStats.Details {
		details = append(details, values...)
	}

	for _, values := range cryptoStats.Details {
		details = append(details, values...)
	}

	sort.Slice(details, func(i, j int) bool {
		return details[i].Quote.Symbol > details[j].Quote.Symbol
	})

	if clear {
		fmt.Print("\033[H\033[2J")
	}

	printInvestmentDetailsTable(details)
}

func portfolio(file string, watch bool) {
	if watch {
		ticker := time.NewTicker(10 * time.Second)
		for ; true; <-ticker.C {
			printPortiflio(file, true)
		}
	}

	printPortiflio(file, false)
}
