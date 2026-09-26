# Code Review: filesharetop

**Date:** 2026-09-26
**Scope:** All hand-written Go under `lib/`, `importer/`, `info/` and `site/`, plus the templates in `site/res/`. The generated file `site/res.go` was not reviewed.

**Method:** Every finding comes from reading the code, not running it. The build needs a GOPATH layout, a live MongoDB and plotinum (`code.google.com/p/plotinum`), which is no longer available upstream.

## Summary

| # | Severity | Issue | Location |
|---|---|---|---|
| 1 | High | The time-window query compares an int to a string, so its first `$or` branch never matches | `importer/importer.go:104`, `info/info.go:92` |
| 2 | High | Stored XSS: scraped titles and links are rendered unescaped | `site/server.go:182-186`, `site/server.go:285` |
| 3 | Medium | A failed import can leave the categories empty, and the site then caches that for 30 minutes | `importer/importer.go:67-75`, `site/server.go:110` |
| 4 | Medium | `Consolidate` swaps collections in a way readers can see, so a partial result can be cached | `importer/importer.go:151-164`, `site/server.go:152` |
| 5 | Medium | The `-1` "not available" value inflates scores | `importer/scorecalc.go:18-21` |
| 6 | Medium | Anyone can flush every cache with `?nocache=1` | `site/server.go:44-54` |
| 7+ | Low | Various, listed below | |

---

## High

### 1. The time-window query never matches its first `$or` branch

**Where:** `importer/importer.go:104` and `info/info.go:92`

`FSTopRecord.Hour` is stored as an `int` (`importer/data.go:10`), but the query compares it to a string:

```go
"hour": bson.M{"$gte": strconv.Itoa(dtlim.Hour())},
```

MongoDB only compares values of the same type, so an int field never satisfies `$gte "5"`. Only the second branch (`date > startDate`) can ever match. This has been the case since the initial import.

**Effect:** every window loses the tail of its start day, which is up to 24 hours:

| Window | Intended | Actual |
|---|---|---|
| `current` | 48h | 24–48h |
| `current_weekly` | 168h | 144–168h |
| `/view` and `/chart` history | 168h | 144–168h |

As a result, scores come from less data than intended.

**Fix:** pass `dtlim.Hour()` as an int. Better still, filter on `import_time: {$gte: dtlim}`, which is a single condition that can use an index.

### 2. Stored XSS on the list and detail pages

**Where:** `site/server.go:182-186` (list page) and `site/server.go:285` (`/view`)

`ii.Title`, `ii.Link` and `ii.Id` go into the HTML through `fmt.Fprintf`. The result is then passed to the template as `template.HTML`, which switches off `html/template` escaping:

- **Titles:** whoever uploads a torrent to the tracker chooses its title, so any uploader can inject script into this site.
- **Links:** `Link` can also be a `javascript:` URL.
- **IDs:** `Id` is inserted into `/chart?id=` and `/view?id=` URLs without query escaping.

**Fix:** move the table markup into `index.tpl` and `base.tpl`, and range over the data there so `html/template` escapes each value for its context. At minimum, apply `template.HTMLEscapeString` to text and `url.QueryEscape` to query values, and reject any link whose scheme isn't `http` or `https`.

---

## Medium

### 3. A failed import can leave the categories empty

**Where:** `importer/importer.go:67-75` and `site/server.go:104-113`

`Import` runs `ccat.RemoveAll(nil)` *before* it calls `fetcher.CategoryMap()`:

- If `CategoryMap()` fails, the `category` collection stays empty until the next hourly import.
- Even when it succeeds, there is a short window where the collection is empty or only partly filled.

The site then makes this worse. `Info.Categories()` always returns a non-nil slice (built with `make`), so the `if c != nil` check at `server.go:110` never stops an empty list from being cached. Once an empty list is cached, every `?category=` page returns 404 for 30 minutes.

**Fix:**
- Call `CategoryMap()` before removing anything.
- Only cache categories when `len(c) > 0`.

### 4. `Consolidate` swaps collections in a way readers can see

**Where:** `importer/importer.go:151-164`

`Consolidate` empties `current_*` with `RemoveAll` and then inserts documents one at a time. A site request during that window reads an empty or partial collection:

- An empty result isn't cached (`server.go:152` checks `len(d) > 0`).
- A partial result is cached, and served for 10 minutes.

**Fix:** write to a temporary collection (for example `current_weekly_tmp`), then run `renameCollection` with `dropTarget: true`. Bulk-inserting into the temporary collection would also be faster.

### 5. The `-1` "not available" value inflates scores

**Where:** `importer/scorecalc.go:18-21`

`Item` stat fields use `-1` to mean "not available", but `CalcScore` subtracts them as if they were real values. If a field is `-1` in one sighting and real in the next, the difference counts as genuine activity:

| Field | Change | Score effect |
|---|---|---|
| Comments | −1 → 5 | +60 |
| Seeders | −1 → 1000 | +5005 |

A drop in the other direction is penalized much less, because the weights for negative changes are smaller. So a fetcher that occasionally fails to parse a field inflates scores.

**Fix:** skip a field's contribution when either value is `< 0`. `/view` (`server.go:294-297`) and `/chart` show the same distorted changes and need the same check.

### 6. Anyone can flush every cache with `?nocache=1`

**Where:** `site/server.go:44-54`

Any visitor can flush all three in-memory caches. Charts are rendered on demand, so one such request per second is enough to keep MongoDB and the chart renderer under full load.

**Fix:** only honour `nocache` from localhost, or require a secret value set in `Config`.

---

## Low

### Importer

- **Title, link and category never update.** They're taken from an item's *first* sighting in the window (`importer.go:124-130`). A torrent that is renamed or recategorized keeps its old values for the whole window. Taking these fields from `Last` would fix this.
- **Import can panic on a nil category map.** `*fcat` is dereferenced without a check (`importer.go:77`), so a fetcher that returns `nil, nil` crashes the import.
- **Index errors are ignored.** `EnsureIndexKey` errors are discarded here and throughout `info/info.go`.

### Site: behavior

- **The server can fail silently.** The error from `http.ListenAndServe` is ignored (`server.go:496`). If the port is already in use, `RunServer` just returns.
- **`HistoryDays` is really a number of hours.** It is set to 168 and passed to `History` as `hours` (`config.go:15`, `server.go:258`, `server.go:366`). It should be renamed to `HistoryHours`.
- **Bad page values give the wrong status and can panic:**
  - An unparsable `pg` returns 500 instead of 400 (`server.go:84`).
  - `PageSize == 0` panics in `FSTopStatsList.PageCount` with a divide by zero.
  - On 32-bit builds, a huge `pg` makes `(page-1)*pagesize` overflow in `Paged`, and the negative start index panics.
- **An empty list gets a broken "Last" link.** With `pagecount == 0`, `page != pagecount` is true, so the page links to `pg=0` (`server.go:212`).
- **Missing assets return the wrong status.** `/res/` returns 500 instead of 404 for a missing file, and sends an empty `Content-Type` for unknown extensions (`server.go:484-490`). Path traversal is not a risk, because `Asset` only reads the embedded map.

### Site: rendering

- **`/view` hides the first sighting.** The first sighting is only used as the baseline for the next row and is never printed. The title row also has `colspan="7"`, but the table has 6 columns (`server.go:285`).
- **The chart X axis counts sightings, not time.** Gaps where the item was missing disappear.
- **The chart mixes totals and changes.** Complete is plotted as a per-hour change while the other lines are totals (`server.go:413`). The `@Complete` legend label hints this is deliberate, but the chart doesn't explain it.
- **The chart plots `-1` values as data.**
- **An item seen once gets an empty chart.** Its `XYs` are empty, so depending on plotinum's behavior the chart draws with infinite axis ranges or panics.
- **Some chart errors panic.** A `plot.New()` error calls `panic` inside the handler (`server.go:385`).
- **Chart output errors are ignored.** The return value of `c.WriteTo(w)` isn't checked (`server.go:478`).

### Site: performance

- **Templates are re-parsed on every request.** `LoadTemplates` runs per request (`server.go:217`, `server.go:324`) and uses `template.Must`, so a broken asset panics on every request. Parse the templates once in `RunServer`.

### Tests

- **The test is stale.** `info/info_test.go` calls `i.Top()` with no argument, but `Top` now takes an `id`. It also needs a live MongoDB. There are no tests for `DefaultScoreCalculator`, `FinishScore` or `FSTopStatsList.Paged`/`PageCount`. These are pure functions and would be easy to cover.

---

## Suggested order

1. **Fix #1 and #2 first.** Both are small, contained changes with a large effect.
2. **Then #3 and #4 together.** They are the same "empty or partial collection gets cached" problem.
3. **Then #5.** Add unit tests for `CalcScore` alongside it.
4. **Bump the version.** #1 and #5 change scores, so bump `importer/version.go`. The fixes for #2, #3 and #6 touch the site, so bump `site/version.go` too.
