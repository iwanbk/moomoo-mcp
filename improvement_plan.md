# Improvement Plan: Token-Efficient Output

Context: tools that can return many rows (`get_historical_klines`,
`get_history_orders`, `get_history_deals`) currently serialize each row as a
full JSON object, repeating the same field names on every row. This is the
main token cost driver for MCP responses in this project.

## Priority summary

| # | Item | Status |
|---|------|--------|
| 1 | Columnar output format + number rounding | Done |
| 2 | `limit` push-down (`maxAckKLNum`) + fix `NextReqKey` pagination | Do |
| 3 | Field selection (`fields`) | Done |
| — | Plain CSV/TSV instead of JSON | Deferred — re-evaluate only if #1 misses the token budget |
| — | Intraday timestamp compaction | Optional — fold into #1 if intraday klines prove common |

---

## 1. Columnar output format + number rounding

### 1a. Columnar format

**Current format** (array-of-objects, field names repeated on every row):

```json
[
  {
    "time": "2024-01-01",
    "open": 123.45,
    "high": 124.0,
    "low": 122.5,
    "close": 123.8,
    "volume": 1000000,
    "turnover": 123456789.12,
    "change_rate": 0.023
  },
  {
    "time": "2024-01-02",
    "open": 123.8,
    "high": 125.1,
    "low": 123.0,
    "close": 124.9,
    "volume": 980000,
    "turnover": 121000000.5,
    "change_rate": 0.008
  }
]
```

**Proposed format** — replace array-of-objects with a single header + array-of-arrays:

```json
{
  "columns": ["time", "open", "high", "low", "close", "volume", "turnover", "change_rate"],
  "rows": [
    ["2024-01-01", 123.45, 124.0, 122.5, 123.8, 1000000, 123456789.12, 0.023],
    ["2024-01-02", 123.8, 125.1, 123.0, 124.9, 980000, 121000000.5, 0.008]
  ]
}
```

Field names are sent once instead of once per row.

- Applies to: `get_historical_klines` (`Kline`), `get_history_orders` /
  `get_history_deals` (`Order`, `Deal`).
- **Estimated improvement:** ~30-50% fewer output tokens for responses with
  many rows (savings grow with row count; negligible for 1-2 rows).
- Trade-off: raw output is less self-describing/readable per row than
  array-of-objects, though still easy for the model to parse given the
  header.

### 1b. Number rounding

Orthogonal, cheap win that applies to **all** tools (snapshots, quotes, order
book, klines), not just the tabular ones. Values like `turnover:
123456789.12` and `change_rate: 0.023` carry more digits than are useful and
each digit is a token.

- Round `change_rate` (and similar ratios) to ~4 decimal places.
- Round prices to sensible tick precision.
- Round/scale large `turnover` values (e.g. drop cents).

- **Estimated improvement:** a few percent across every response, compounding
  with 1a on tabular responses.
- Trade-off: minor loss of precision; pick rounding that stays below what a
  caller would actually act on.

### Intraday timestamp compaction (optional)

For `1min`/`5min`/etc., many consecutive rows share the same date (390 `1min`
rows/day all have the same `yyyy-MM-dd`). Splitting into a `date` (emitted
once) + `time` (per row, just `HH:mm:ss`) avoids repeating the date string on
every row. Fold this into 1a only if intraday kline requests prove common;
otherwise skip — it adds a few percent for intraday only.

---

## 2. `limit` push-down + pagination fix

This item is part optimization, part **correctness fix**. Before writing code,
resolve the design decision below.

### 2a. Pagination is currently broken (correctness)

`RequestHistoryKL` returns a `NextReqKey` (`Qot_RequestHistoryKL.pb.go`, S2C
field 3): when a single request cannot return all data, the next request must
carry this key to continue. The current `GetKlines`
([internal/moomoo/market.go:164-187](internal/moomoo/market.go#L164-L187))
reads only the first `GetKlList()` page and **ignores `NextReqKey`** — so a
broad range today is **silently truncated to one page**, not returned in full.

The "unbounded query returns thousands of rows" premise that motivates a
`limit` is therefore only true *after* pagination is implemented. Decide one
of:

- **(A) Paginate to completeness**, then rely on `limit` (below) as the guard
  against huge responses. Preferred if callers expect full ranges.
- **(B) Document single-page behavior** and expose a way to continue.

`get_history_orders` / `get_history_deals` likely have the same pagination
shape — verify and apply the same decision.

### 2b. `limit` pushed down to OpenD

Add an optional `limit` (max row count) argument to `get_historical_klines`,
`get_history_orders`, and `get_history_deals`, with a sane default (e.g. 200)
applied when the caller omits it.

Push the limit **down to the server**, don't fetch everything then slice in
Go. `RequestHistoryKL` has a native `maxAckKLNum` field
(`Qot_RequestHistoryKL.pb.go` C2S field 6, "max K-lines to return"), settable
via `adapt.With("maxAckKLNum", int32(n))` on `RequestHistoryKLWithContext`.
This lets OpenD cap the result — less data over the wire and correct
semantics.

- **To verify:** with a begin/end range + `maxAckKLNum`, does the API return
  the first N rows or the last N? For klines, callers usually want the
  **most recent** rows.

**Estimated row counts for `get_historical_klines`**, assuming ~390 trading
minutes/day (US market; other markets have slightly fewer minutes/day, e.g.
HK ~330), ~21 trading days/month, ~252/year:

| kl_type | per day | 1 week | 1 month | 3 months | 1 year |
|---------|--------:|-------:|--------:|---------:|-------:|
| 1min    | 390     | 1,950  | 8,190   | 24,570   | 98,280 |
| 5min    | 78      | 390    | 1,638   | 4,914    | 19,656 |
| 15min   | 26      | 130    | 546     | 1,638    | 6,552  |
| 30min   | 13      | 65     | 273     | 819      | 3,276  |
| 60min   | 7       | 35     | 147     | 441      | 1,764  |
| day     | 1       | 5      | 21      | 63       | 252    |
| week    | —       | ~0.2   | ~4      | 13       | 52     |
| month   | —       | —      | ~1      | ~3       | 12     |

Note: intraday types (`1min`-`60min`) grow fast — a `1min` request over even
one month can already return 8k+ rows, which is the scenario a `limit`
default is meant to guard against.

- **Estimated improvement:** caps worst-case response size for broad queries
  (e.g. `1min` klines over months, or unbounded history windows) — prevents
  unintentional multi-thousand-row responses that would otherwise dominate
  token usage.
- Trade-off: callers who genuinely need the full range must pass a higher
  explicit `limit` (or paginate), adding one extra round trip in that case.

---

## 3. Field selection (`fields`)

Let the caller pass a `fields` argument to control which columns are emitted.
This is independent of #1 (it changes *which* columns are sent, not the
*shape* of each row) and independent of the deferred CSV idea.

No backward-compatibility constraint applies (early stage), so the **default
should be a lean column set, not "all fields"** — this way the token saving
applies automatically, without callers having to opt in. Emit only what most
queries need by default and let `fields` opt into the rest.

- Proposed kline default: `time, open, high, low, close, volume`. Drop
  `turnover` and `change_rate` from the default — rarely used and derivable
  from the other columns (2 of 8 columns gone by default).
- Callers who want the dropped columns pass them explicitly via `fields`.
- Apply the same "lean default + opt-in" approach to `get_history_orders` /
  `get_history_deals`.

- Applies to the same tabular tools as #1.
- **Estimated improvement:** proportional to columns dropped — the lean kline
  default (6 of 8 columns) is roughly a further ~25% on top of #1 for those
  responses, applied automatically.
- Trade-off: a caller who needs a dropped column must name it in `fields`;
  the default must be chosen well so that's rare.

---

## Deferred: plain CSV/TSV text instead of JSON

The columnar JSON in #1 still carries JSON syntax overhead per row: quotes
around the date string, a comma between every value, and `[`/`]` wrapping
each row. Dropping JSON entirely for tabular tools and returning a plain
CSV/TSV text block removes that overhead on top of the columnar win:

```
time,open,high,low,close,volume,turnover,change_rate
2024-01-01,123.45,124.0,122.5,123.8,1000000,123456789.12,0.023
2024-01-02,123.8,125.1,123.0,124.9,980000,121000000.5,0.008
```

- **Estimated improvement:** roughly another ~15-25% on top of #1's
  30-50%, so combined vs. the original array-of-objects format: roughly
  **45-65% fewer output tokens** for large tabular responses.
- Trade-off: loses JSON's type fidelity (everything is text, caller must
  parse numbers) and stacks another format decision on top of #1 — more
  implementation and testing surface for a marginal additional gain.

**Decision: deferred.** Do #1 + #3 first; re-evaluate CSV only if the token
budget is still exceeded afterward.
