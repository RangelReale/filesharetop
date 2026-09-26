# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Overview

`filesharetop` is the shared core library for a family of "top torrents" sites. It has no `main` package of its own. Each tracker has its own repo (`kat-fstop`, `nyaa-fstop`, `avistaz-fstop`, `asiatorrents-fstop`, all under `github.com/RangelReale/`). Each of those supplies a scraper plus two small binaries (`<site>-importer` and `<site>-site`) that wire the scraper into the packages here.

This is legacy, pre-modules Go code. There is no `go.mod`, and imports rely on GOPATH layout (`$GOPATH/src/github.com/RangelReale/filesharetop`). Some dependencies are dead upstream: `code.google.com/p/plotinum` (now `gonum.org/v1/plot`) and `gopkg.in/mgo.v2` (unmaintained). Building with a modern toolchain means either setting up GOPATH with `GO111MODULE=off` or adding a `go.mod` and replacing those imports.

## Packages

Directory names and package names differ, so check the import alias when reading code:

| Dir | Package | Role |
|---|---|---|
| `lib/` | `fstoplib` | The contract with the site plugins: the `Fetcher` interface, `Item`, `CategoryMap` |
| `importer/` | `fstopimp` | Stores fetcher output in MongoDB and computes scores |
| `info/` | `fstopinfo` | Read-side queries used by the website |
| `site/` | `fstopsite` | HTTP server (`RunServer(*Config)`), templates, PNG charts |

## Data flow

1. A plugin's importer binary runs periodically (hourly, since records are keyed by date and hour). It calls `Importer.Import(fetcher)`, which:
   - upserts one `history` document keyed by `{date, hour}` (UTC) that holds the whole `map[id]*Item`;
   - replaces the `category` collection with `fetcher.CategoryMap()`. That maps a display category to the list of raw site category strings.
2. Then it calls `Importer.Consolidate(id, hours)`, usually twice: `("", 48)` → collection `current` and `("weekly", 168)` → `current_weekly`. See `BuildCurrentCollectionName`. Consolidate walks the history in that window. For each pair of consecutive sightings of an item, it sums `ScoreCalculator.CalcScore(current, previous)`, which weights deltas in seeders, leechers, completed and comments. `FinishScore` then scales the total down by the fraction of imports in which the item was missing.
3. The site reads through `fstopinfo.Info`: `Top(id)`, `TopCategory(id, category)`, `History(id, hours)` and `Categories()`. `site.Config.TopId` chooses which `current_*` collection is shown (plugins use `"weekly"`).

`Item` stat fields use `-1` to mean "not available". Always build items with `fstoplib.NewItem()`, not a struct literal. `AddDate` is a `YYYY-MM-DD` string.

The default database name is `filesharetop`, but every plugin overrides it (`fstop_<site>`). The MongoDB host is hard-coded in the plugin binaries (`localhost` / `127.0.0.1`).

## Site

- Routes live in `site/server.go`: `/` (list, with `category`, `pg` and `chart` params), `/view` (item detail), `/chart` (a PNG rendered with plotinum) and `/res/` (static assets). `?nocache=1` clears the in-memory `go-cache` caches.
- Templates, CSS and fonts in `site/res/` are embedded into `site/res.go` with `go-bindata`. **After editing anything under `site/res/`, regenerate with** `cd site && ./buildres.sh` (or `buildres.bat`); `go-bindata` must be installed. Don't hand-edit `res.go`.
- `LoadTemplates` always parses `header` and `footer` along with the named templates.

## Testing

`info/info_test.go` needs a live MongoDB on `localhost`. It is also stale: it calls `i.Top()` with no argument, but `Top` now takes an `id`. Run a single test with `go test ./info -run TestInfo`.

## Versioning

`importer/version.go` and `site/version.go` each have a `VERSION` constant, which the plugin binaries print with `-version`. Bump the relevant one when changing behavior.
