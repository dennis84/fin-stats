package main

import (
  "time"
  "github.com/piquette/finance-go"
)

type MarketConfig struct {
  OpenPreAt string
  OpenAt string
  CloseAt string
  ClosePostAt string
}

type MarketInfo struct {
  DurationUntilOpen *time.Duration
  DurationUntilOpenPre *time.Duration
  DurationUntilClose *time.Duration
  DurationUntilClosePost *time.Duration
}

var markets = map[string]MarketConfig{
  "de_market": MarketConfig{"08:00 AM", "09:00 AM", "05:30 PM", "08:00 PM"},
  "us_market": MarketConfig{"04:00 AM", "09:30 AM", "04:00 PM", "08:00 PM"},
  "hk_market": MarketConfig{"", "09:30 AM", "04:00 PM", ""},
  "dk_market": MarketConfig{"", "09:00 AM", "05:00 PM", ""},
  "gb_market": MarketConfig{"", "09:00 AM", "05:00 PM", ""},
  "fr_market": MarketConfig{"", "09:00 AM", "05:30 PM", ""},
  "cn_market": MarketConfig{"", "09:15 AM", "03:00 PM", ""},
  "au_market": MarketConfig{"", "09:30 AM", "04:00 PM", ""},
  "jp_market": MarketConfig{"", "09:00 AM", "03:00 PM", ""},
}

func getMarketInfo(q finance.Quote) MarketInfo {
  now := time.Now()
  info := MarketInfo{}

  if conf, ok := markets[q.MarketID]; ok {
    tz := q.ExchangeTimezoneName

    if conf.OpenPreAt != "" {
      duration := getDateAt(now, conf.OpenPreAt, tz).Sub(now)
      info.DurationUntilOpenPre = &duration
    }

    if conf.OpenAt != "" {
      duration := getDateAt(now, conf.OpenAt, tz).Sub(now)
      info.DurationUntilOpen = &duration
    }

    if conf.CloseAt != "" {
      duration := getDateAt(now, conf.CloseAt, tz).Sub(now)
      info.DurationUntilClose = &duration
    }

    if conf.ClosePostAt != "" {
      duration := getDateAt(now, conf.ClosePostAt, tz).Sub(now)
      info.DurationUntilClosePost = &duration
    }
  }

  return info
}

func getDateAt(ref time.Time, hour string, timezone string) time.Time {
  loc, _ := time.LoadLocation(timezone)
  parsed, _ := time.ParseInLocation("03:04 PM", hour, loc)
  ref = getNextBusinessDay(ref)

  t := time.Date(
    ref.Year(),
    ref.Month(),
    ref.Day(),
    parsed.Hour(),
    parsed.Minute(),
    parsed.Second(),
    parsed.Nanosecond(),
    loc,
  )

  return t
}

func getNextBusinessDay(t time.Time) time.Time {
  for {
    if isBusinessDay(t) {
      return t
    }

    t = t.Add(time.Hour * 24)
  }
}

func isBusinessDay(t time.Time) bool {
  return t.Weekday() != time.Saturday && t.Weekday() != time.Sunday
}
