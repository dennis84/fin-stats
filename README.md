# Fin Stats

Sum finances of bank accounts, stocks, assets and crypto for a better overview.

## Usage

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
fin-stats
# pass filename
fin-stats -file ~/finances.yaml
# no details
fin-stats -no-details
# no summary
fin-stats -no-summary
# no graph
fin-stats -no-graph
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

stock_details:
  AAPL:
  - units: 2
    sum: 263.62
    in: 260
    diff: 3.6200000000000045
    quote:
      price: 131.81
      pct: 2.7918599
  TSLA:
  - units: 1
    sum: 656.99
    in: 400
    diff: 256.99
    quote:
      price: 656.99
      pct: 1.0971602
  - units: 1
    sum: 656.99
    in: 500
    diff: 156.99
    quote:
      price: 656.99
      pct: 1.0971602

asset_details:
  GC=F:
  - units: 2
    sum: 3750.8
    in: 2400
    diff: 1350.8000000000002
    quote:
      price: 1875.4
      pct: -0.39303294


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
