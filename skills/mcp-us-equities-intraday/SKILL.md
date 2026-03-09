---
name: mcp-us-equities-intraday
description: "Intraday market data queries: quotes, bars, large trades, Greeks, auction imbalance."
---

# MCP US Equities Intraday

## Purpose
Use this skill for real-time and intraday market data queries:
- Bid/ask spreads and liquidity monitoring
- Intraday VWAP and momentum tracking
- Large trade (institutional flow) detection
- Options Greeks and gamma exposure
- Pre-market auction imbalance

## Prerequisites
1. Install and configure `rw`:
```bash
curl -fsSL https://raw.githubusercontent.com/zz3310969/max-skills/main/scripts/install-rw.sh | bash
rw setup --server-url http://113.44.56.214:18080/mcp/
rw doctor
```

## Preflight (required before any rw call)
```bash
if ! command -v rw >/dev/null 2>&1; then
  curl -fsSL https://raw.githubusercontent.com/zz3310969/max-skills/main/scripts/install-rw.sh | bash
fi
rw --version || true
rw doctor
```

## Tool Selection
| Scenario | Tool |
|----------|------|
| Bid/ask spread, liquidity check | `read_intraday_snapshot` |
| VWAP breakout, momentum | `read_intraday_snapshot` |
| Institutional flow, block trades | `read_large_trades` |
| Options delta/gamma/IV | `read_options_greeks` |
| Pre-market direction signal | `read_auction_imbalance` |

## Commands

### `read_intraday_snapshot`
Get latest quotes (bid/ask/spread) and bars (OHLCV/VWAP) for multiple tickers.

```bash
rw call --tool read_intraday_snapshot --args '{"tickers":["NVDA","AAPL","QQQ"],"include_quotes":true,"include_bars":true,"bar_interval":"5Min"}'
```

Parameters:
- `tickers[]` (required): US equity/ETF symbols
- `include_quotes` (default: true): include bid/ask/spread data
- `include_bars` (default: true): include OHLCV + VWAP
- `bar_interval`: `1Min|5Min|15Min|30Min` (default: `5Min`)

Response fields:
- `quote.bid_price`, `quote.ask_price`, `quote.spread_pct`
- `bar.close`, `bar.vwap`, `bar.volume`

### `read_large_trades`
Track institutional-size trades (notional > $100K).

```bash
rw call --tool read_large_trades --args '{"ticker":"NVDA","min_notional":100000,"limit":50}'
```

Parameters:
- `ticker` (required): single US equity symbol
- `since`: ISO8601 timestamp filter
- `min_notional` (default: 100000): minimum trade size in USD
- `limit` (default: 100, max: 500)

Response fields:
- `items[].price`, `items[].size`, `items[].notional`, `items[].side`
- `summary_24h.total_notional`, `summary_24h.buy_notional`, `summary_24h.sell_notional`

### `read_options_greeks`
Get ATM options Greeks for gamma exposure analysis.

```bash
rw call --tool read_options_greeks --args '{"ticker":"NVDA","moneyness_range":5,"limit":50}'
```

Parameters:
- `ticker` (required): underlying symbol
- `expiration`: filter by expiry date (YYYY-MM-DD)
- `option_type`: `call|put`
- `moneyness_range` (default: 5): ATM ± percentage
- `limit` (default: 50, max: 200)

Response fields:
- `items[].delta`, `items[].gamma`, `items[].theta`, `items[].vega`, `items[].iv`
- `gamma_exposure.net_gamma`, `gamma_exposure.call_gamma`, `gamma_exposure.put_gamma`

### `read_auction_imbalance`
Get pre-market auction data for opening direction signals.

```bash
rw call --tool read_auction_imbalance --args '{"tickers":["NVDA","AAPL","MSFT"]}'
```

Parameters:
- `tickers[]` (required): US equity/ETF symbols
- `date`: specific date (YYYY-MM-DD), defaults to today

Response fields:
- `items[].imbalance`, `items[].imbalance_side` (`buy|sell|balanced`)
- `items[].auction_price`, `items[].paired_shares`
- `summary.buy_imbalance_count`, `summary.sell_imbalance_count`

## Signal Types (from these tools)
| Signal | Source | Trigger |
|--------|--------|---------|
| `SPREAD_ANOMALY` | `read_intraday_snapshot` | spread_pct > 7d avg + 2σ |
| `VWAP_BREAKOUT` | `read_intraday_snapshot` | price deviates > 1.5% from VWAP with volume surge |
| `LARGE_TRADE_CLUSTER` | `read_large_trades` | cumulative notional > $500K in 3h |
| `GAMMA_SPIKE` | `read_options_greeks` | gamma exposure > 90th percentile |
| `AUCTION_IMBALANCE` | `read_auction_imbalance` | imbalance > 5% of avg daily volume |

## Use Cases

### Morning Prep (before market open)
```bash
# Check auction imbalance for watchlist
rw call --tool read_auction_imbalance --args '{"tickers":["NVDA","AAPL","MSFT","GOOGL","AMZN"]}'
```

### Intraday Monitoring
```bash
# Check spread and VWAP for active positions
rw call --tool read_intraday_snapshot --args '{"tickers":["NVDA","TSM"],"include_quotes":true,"include_bars":true}'

# Track institutional activity
rw call --tool read_large_trades --args '{"ticker":"NVDA","min_notional":200000,"limit":20}'
```

### Options Analysis (after close)
```bash
# Check gamma exposure for hedging pressure
rw call --tool read_options_greeks --args '{"ticker":"NVDA","moneyness_range":5}'
```

## Result Handling
Follow `mcp-us-equities-response` skill for envelope interpretation:
- `ok=true`: use `data` fields directly
- `ok=false`: check `error.code` and `error.suggestion`
- `warnings`: data is partial but usable

Reference mapping: `docs/mcp-tool-map-v2.md`.
