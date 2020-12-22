# Fin Stats

Sum finances of bank accounts, stocks, assets and crypto for a better overview.

## Build

```bash
go build fin-stats.go
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
```

Output:

```bash
date: 2020-12-22T10:59:57.529261789+01:00
savings: 11200
stocks:
  sum: 1577.6
  in: 1160
  diff: 417.5999999999999
  loss: 0
assets:
  sum: 3750.8
  in: 2400
  diff: 1350.8000000000002
  loss: 0
investments_sum: 5328.4
total: 16528.4
income: 2500
expenses: 1320
budget: 1180

Total:
 16534 ┼──────────────╮        ╭
 15868 ┤              │        │
 15201 ┤              │        │
 14534 ┤              │        │
 13867 ┤              │        │
 13200 ┤              │        │
 12534 ┤              │     ╭──╯
 11867 ┤              │     │
 11200 ┤              ╰─────╯
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

## Mentions

Print mentions from wsb

```bash
fin-stats mentions --watch -n 10
```

Output:

```
+--------+----------+--------+-------+
| SYMBOL | MENTIONS | PRICE  |  PCT  |
+--------+----------+--------+-------+
| AAPL   |     2038 | 132.15 |  3.06 |
| PLTR   |     1769 |  28.14 | -1.30 |
| GME    |     1367 |  19.39 | 24.86 |
| TSLA   |     1310 | 622.61 | -4.19 |
| SPY    |      671 | 367.51 | -0.10 |
| QS     |      518 | 121.06 | 27.57 |
| NIO    |      449 |  47.08 | -3.82 |
| VLDR   |      238 |  25.97 |  5.23 |
| PTON   |      237 | 161.78 | 12.04 |
| AMD    |      187 |  91.70 | -1.64 |
+--------+----------+--------+-------+
```

## Portfolio

Print portfolio stats

```bash
fin-stats portfolio
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
