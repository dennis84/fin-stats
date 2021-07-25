# Fin Stats

Command line interface for daily finance statistics

- Sum finances of bank accounts, stocks, assets and crypto for a better overview
- Quote statistics
- Portfolio statistics
- WSB trending

## Build

```bash
go build
```

## Sum

Config:

```yaml
# Output currency (EUR or USD)
currency: USD

savings:
  sparkasse: 10000
  visa: 1200

# List of investments, fetched from yahoo finance.
investments:
  assets:
    GC=F: [{units: 2, in: 1200}]
  stocks:
    TSLA: [{units: 1, in: 400}, {units: 1, in: 500}]
    AAPL: [{units: 2, in: 140}]
  crypto:
    BTC-USD: [{units: 1, in: 24000}]

income:
  salary: 2500
  paperboy: 300

expenses:
  rent: 1200
  car: 50
  internet: 40
  electricity: 30
```

Run:

```bash
# help
fin-stats -h
# use ~/finances.yaml
fin-stats sum
# pass filename
fin-stats sum -f ~/finances.yaml
# no graph
fin-stats sum --no-graph
# no summary
fin-stats sum --no-summary
# watch mode
fin-stats sum --watch
```

Output:

```bash
+-----------------+----------+
| Savings         | 11200.00 |
| Stocks Sum      |  1542.67 |
| Stocks In       |  1160.00 |
| Stocks Diff     |   382.67 |
| Stocks Loss     |     0.00 |
| Assets Sum      |  3729.80 |
| Assets In       |  2400.00 |
| Assets Diff     |  1329.80 |
| Assets Loss     |     0.00 |
| Investments Sum |  5272.47 |
| Total           | 16472.47 |
+-----------------+----------+

Total:
 16527 ┼───╮
 16518 ┤   │
 16508 ┤   │
 16499 ┤   │
 16489 ┤   │
 16480 ┤   │
 16470 ┤   │                               ╭
 16461 ┤   │                       ╭───────╯
 16451 ┤   ╰───────────────────────╯
```

## Quote

Print quote stats

```bash
fin-stats quote --watch aapl
```

Output:

```bash
symbol: AAPL
price: 132.18
pct: 3.0804002

 132 ┤                               ╭────
 132 ┤                         ╭╮  ╭─╯
 132 ┤                         │╰─╮│
 132 ┤                      ╭╮ │  ││
 132 ┤   ╭───────╮╭─╮╭╮     ││ │  ││
 132 ┤╭─╮│       ││ │││  ╭╮ ││ │  ╰╯
 132 ┼╯ ╰╯       ││ ╰╯╰──╯╰─╯│ │
 132 ┤           ││          ╰─╯
 132 ┤           ╰╯
```

## Trending

Print trending from wsb

```bash
fin-stats trending --watch -n 10
```

Output:

```
+--------+---------------------------+--------+-------+
| SYMBOL |           NAME            | PRICE  |  PCT  |
+--------+---------------------------+--------+-------+
| NVDA   | NVIDIA                    | 195.58 | -0.18 |
| CLOV   | Clover Health Investments |   8.05 | -6.61 |
| AMC    | AMC Entertainment         |  36.99 | -0.67 |
| GME    | GameStop                  | 180.36 | 0.84  |
| WISH   | ContextLogic              |   9.18 | -3.37 |
+--------+---------------------------+--------+-------+
```

## Portfolio

Print portfolio stats

```bash
fin-stats portfolio --watch
```

Output:

```
+--------+---------+---------+---------+-------------+-----------+
| SYMBOL |   SUM   |   IN    |  DIFF   | QUOTE PRICE | QUOTE PCT |
+--------+---------+---------+---------+-------------+-----------+
| TSLA   |  635.50 |  400.00 |  235.50 |      635.50 |     -2.21 |
| TSLA   |  635.50 |  500.00 |  135.50 |      635.50 |     -2.21 |
| GC=F   | 3734.00 | 2400.00 | 1334.00 |     1867.00 |     -0.84 |
| AAPL   |  263.20 |  260.00 |    3.20 |      131.60 |      2.63 |
+--------+---------+---------+---------+-------------+-----------+
```
