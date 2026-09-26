# filesharetop

Shared core for a family of "top torrents" sites. It ranks torrents by how fast their activity grows over time, not by their raw seeder counts.

This repo is a library only. Each tracker has its own plugin repo, which provides a scraper and the two runnable binaries:

| Plugin | Tracker | Site port |
|---|---|---|
| [asiatorrents-fstop](https://github.com/RangelReale/asiatorrents-fstop) | AsiaTorrents | 13111 |
| [nyaa-fstop](https://github.com/RangelReale/nyaa-fstop) | Nyaa | 13112 |
| [avistaz-fstop](https://github.com/RangelReale/avistaz-fstop) | AvistaZ | 13113 |
| [kat-fstop](https://github.com/RangelReale/kat-fstop) | KickassTorrents | 13114 |

## How it works

1. **Import.** A plugin's `<site>-importer` runs periodically (for example, hourly from cron). It scrapes the tracker's top lists and saves a snapshot to MongoDB (collection `history`, one document per UTC date and hour). It also saves the tracker's category mapping.
2. **Consolidate.** The importer then scores each torrent over a 48-hour window (collection `current`) and a 7-day window (collection `current_weekly`):
   - Between consecutive snapshots, rises in seeders, leechers, completed downloads and comments add points; drops subtract fewer points.
   - A torrent's score is reduced in proportion to the imports it was missing from.
3. **Serve.** The plugin's `<site>-site` binary runs the web UI. It shows ranked lists with category filtering, paging, per-torrent history and PNG charts.

## Packages

| Directory | Package | Purpose |
|---|---|---|
| `lib/` | `fstoplib` | `Fetcher` interface and `Item` type implemented by plugins |
| `importer/` | `fstopimp` | Snapshot storage and score consolidation |
| `info/` | `fstopinfo` | Read queries for the website |
| `site/` | `fstopsite` | HTTP server, templates, charts |

## Writing a plugin

Implement `fstoplib.Fetcher`:

```go
type Fetcher interface {
	ID() string
	SetLogger(l *log.Logger)
	Fetch() (map[string]*Item, error)          // keyed by torrent ID
	CategoryMap() (*CategoryMap, error)        // display category -> site categories
}
```

Create items with `fstoplib.NewItem()`, so stats the tracker doesn't report stay at `-1`.

A minimal importer looks like this:

```go
imp := fstopimp.NewImporter(logger, session)
imp.Database = "fstop_mysite"
imp.Import(fetcher)
imp.Consolidate("", 48)
imp.Consolidate("weekly", 168)
```

A minimal site looks like this:

```go
config := fstopsite.NewConfig(port)
config.Title, config.Logger, config.Session = "My Top", logger, session
config.Database, config.TopId = "fstop_mysite", "weekly"
fstopsite.RunServer(config)
```

## Requirements

- Go, with the code laid out in a GOPATH workspace. This is pre-modules code with no `go.mod`.
- MongoDB.
- Dependencies:
  - `gopkg.in/mgo.v2`
  - `github.com/pmylund/go-cache`
  - `code.google.com/p/plotinum`. That host is gone; the project continues as `gonum.org/v1/plot`.

## Development

Templates, CSS and fonts under `site/res/` are embedded into `site/res.go` with [go-bindata](https://github.com/jteeuwen/go-bindata). Regenerate after editing:

```sh
cd site && ./buildres.sh
```
