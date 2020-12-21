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
# use ~/finances.yaml
fin-stats
# pass filename
fin-stats -f ~/finances.yaml
# print investments details
fin-stats -f ~/finances.yaml -d
```

Output:
```bash
finance_stats:
  file:           ./finances.yaml
  date:           2020-12-21T15:12:53+01:00
stocks_details:
  - symbol:       TSLA
    units:        1
    in:           400.00
    current:      665.05
    profit:       265.05
    unit_price:   665.05
    today_change: -4.31
  - symbol:       TSLA
    units:        1
    in:           500.00
    current:      665.05
    profit:       165.05
    unit_price:   665.05
    today_change: -4.31
  - symbol:       AAPL
    units:        2
    in:           280.00
    current:      250.20
    loss:         29.80
    unit_price:   125.10
    today_change: -1.23
assets_details:
  - symbol:       GC=F
    units:        2
    in:           2400.00
    current:      3775.80
    profit:       1375.80
    unit_price:   1887.90
    today_change: -0.05
savings:          11200.00
stocks:
  current:        1580.30
  diff:           400.30
  losses:         29.80
assets:
  current:        3775.80
  diff:           1375.80
  losses:         0.00
total:            16556.10
income:           2500.00
expenses:         1320.00
budget:           1180.00
```
