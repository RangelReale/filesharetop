package fstopsite

import (
	"bytes"
	"code.google.com/p/plotinum/plot"
	"code.google.com/p/plotinum/plotter"
	"code.google.com/p/plotinum/vg"
	"code.google.com/p/plotinum/vg/vgimg"
	"fmt"
	"github.com/RangelReale/filesharetop/importer"
	"github.com/RangelReale/filesharetop/info"
	"github.com/RangelReale/filesharetop/lib"
	"github.com/pmylund/go-cache"
	"html/template"
	"image/color"
	"mime"
	"net/http"
	"net/url"
	"path"
	"strconv"
	"strings"
	"time"
)

func RunServer(config *Config) {
	// manually load fonts from resource
	for _, fontasset := range AssetNames() {
		if strings.HasPrefix(fontasset, "res/fonts/") {
			fontname := strings.TrimSuffix(path.Base(fontasset), path.Ext(fontasset))
			fontbytes, err := Asset(fontasset)
			if err != nil {
				panic(err)
			}
			fontreader := bytes.NewReader(fontbytes)
			vg.LoadFont(fontname, fontreader)
		}
	}

	// create memory caches
	homeCache := cache.New(10*time.Minute, 5*time.Minute)
	categoryCache := cache.New(30*time.Minute, 30*time.Minute)
	detailCache := cache.New(15*time.Minute, 10*time.Minute)

	initialCheck := func(w http.ResponseWriter, r *http.Request) bool {
		// clear cache if requested
		nocache := r.Form.Get("nocache")
		if nocache == "1" {
			homeCache.Flush()
			categoryCache.Flush()
			detailCache.Flush()
		}

		return true
	}

	http.HandleFunc("/favicon.ico", func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "Not found", 404)
	})
	http.HandleFunc("/favicon.png", func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "Not found", 404)
	})

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		//config.Logger.Printf("Connection: %s\n", r.URL.String())

		var err error

		r.ParseForm()

		if !initialCheck(w, r) {
			return
		}

		csession := config.Session.Clone()
		defer csession.Close()

		cat := r.Form.Get("category")
		chart := r.Form.Get("chart")
		p_page := r.Form.Get("pg")
		page := 1
		if p_page != "" {
			t_page, err := strconv.ParseInt(p_page, 10, 32)
			if err != nil {
				http.Error(w, err.Error(), 500)
				return
			}
			page = int(t_page)
		}
		if page < 1 {
			page = 1
		}

		i := fstopinfo.NewInfo(config.Logger, csession)
		i.Database = config.Database

		// load categories
		var c fstopinfo.FSCategoryList

		catcache, catfound := categoryCache.Get("category")
		if catfound {
			c = catcache.(fstopinfo.FSCategoryList)
		}

		if c == nil {
			c, err = i.Categories()
			if err != nil {
				http.Error(w, err.Error(), 500)
				return
			}
			if c != nil {
				categoryCache.Set("category", c, 0)
			}
		}

		if c == nil {
			http.Error(w, "Could not load categories", 500)
			return
		}

		// load home
		var d fstopimp.FSTopStatsList

		var dcache interface{}
		var dfound bool
		var dname string
		if cat == "" {
			dname = "index"
		} else {
			// check if category exists
			if !c.Exists(cat) {
				http.Error(w, "Category not found", 404)
				return
			}
			dname = cat
		}
		dcache, dfound = homeCache.Get(dname)
		if dfound {
			d = dcache.(fstopimp.FSTopStatsList)
		}

		// data not on cache, load
		if d == nil {
			if cat == "" {
				d, err = i.Top(config.TopId)
			} else {
				d, err = i.TopCategory(config.TopId, cat)
			}
			if err != nil {
				http.Error(w, err.Error(), 500)
				return
			}
			if len(d) > 0 {
				homeCache.Set(dname, d, 0)
			}
		}

		pagecount := d.PageCount(config.PageSize)
		d = d.Paged(page, config.PageSize)

		w.Header().Add("Content-Type", "text/html; charset=utf-8")

		var body *bytes.Buffer = new(bytes.Buffer)

		tmpldata := map[string]interface{}{
			"Title":      config.Title,
			"page":       page,
			"categories": c,
		}

		fmt.Fprintln(body, "<table class=\"main\">")

		fmt.Fprintln(body, "<tr><th>Title</th><th width=\"8%\">Added</th><th width=\"5%\">Score</th><th width=\"5%\">Count</th><th width=\"5%\">Comm.</th></tr>")

		for _, ii := range d {
			fmt.Fprintf(body, "<tr><td><a href=\"%s\">%s</a> <a href=\"/view?id=%s\">[data]</a> <a href=\"/chart?id=%s\">[chart]</a></td>"+
				"<td align=\"center\">%s</td>"+
				"<td align=\"right\">%d</td><td align=\"center\">%d</td><td align=\"center\">%d</td>\n",
				ii.Link, ii.Title, ii.Id, ii.Id, FormatAddDate(ii.Last.AddDate), ii.Score, ii.Count, ii.Last.Comments)

			if chart == "" {
				fmt.Fprintf(body, "<td><img style=\"height: 107px;\" src=\"/chart?id=%s&size=short\"></td>",
					ii.Id)
			}

			fmt.Fprintf(body, "</tr>\n")

			if chart == "1" {
				fmt.Fprintf(body, "<tr><td colspan=\"5\" align=\"center\"><img src=\"/chart?id=%s&size=small\"/></td></tr>\n", ii.Id)
			}
		}
		fmt.Fprintln(body, "</table>")

		pageparams := url.Values{}
		if cat != "" {
			pageparams.Add("category", cat)
		}
		if page > 1 {
			pageparams.Set("pg", "1")
			tmpldata["page_first"] = fmt.Sprintf("/?%s", pageparams.Encode())
		}
		if page > 1 {
			pageparams.Set("pg", strconv.Itoa(page-1))
			tmpldata["page_prev"] = fmt.Sprintf("/?%s", pageparams.Encode())
		}
		if len(d) > 0 && page < pagecount {
			pageparams.Set("pg", strconv.Itoa(page+1))
			tmpldata["page_next"] = fmt.Sprintf("/?%s", pageparams.Encode())
		}
		if page != pagecount {
			pageparams.Set("pg", strconv.Itoa(pagecount))
			tmpldata["page_last"] = fmt.Sprintf("/?%s", pageparams.Encode())
		}

		tmpl := LoadTemplates("index")
		tmpldata["Body"] = template.HTML(body.String())

		err = tmpl.ExecuteTemplate(w, "index", tmpldata)
		if err != nil {
			http.Error(w, "Error processing template", 500)
			return
		}

	})

	http.HandleFunc("/view", func(w http.ResponseWriter, r *http.Request) {
		r.ParseForm()

		if !initialCheck(w, r) {
			return
		}

		csession := config.Session.Clone()
		defer csession.Close()

		w.Header().Add("Content-Type", "text/html; charset=utf-8")

		id := r.Form.Get("id")
		if id == "" {
			http.Error(w, "ID not sent", 500)
			return
		}

		i := fstopinfo.NewInfo(config.Logger, csession)
		i.Database = config.Database

		var d []*fstopinfo.FSInfoHistory
		var err error

		dcache, found := detailCache.Get(id)
		if found {
			d = dcache.([]*fstopinfo.FSInfoHistory)
		}

		if d == nil {
			d, err = i.History(id, config.HistoryDays)
			if err != nil {
				http.Error(w, err.Error(), 500)
				return
			}
			if d != nil {
				detailCache.Set(id, d, 0)
			}
		}

		if d == nil {
			http.Error(w, "Not found", 500)
			return
		}

		var body *bytes.Buffer = new(bytes.Buffer)

		fmt.Fprintln(body, "<table class=\"main\" border=\"1\">")

		fmt.Fprintln(body, "<tr><th>Date</th><th>Hour</th><th>Seeders</th><th>Leechers</th><th>Complete</th><th>Comments</th></tr>")

		var last *fstoplib.Item
		first := true

		for _, ii := range d {
			if ii.Item != nil {
				if first {
					fmt.Fprintf(body, "<tr><td colspan=\"6\">%s <a href=\"%s\">[goto]</a></td></tr>", strings.TrimSpace(ii.Item.Title), ii.Item.Link)
					first = false
				}

				if last != nil {
					//fmt.Printf("%s - [%d] [%d] [%d]\n", item.Title, pi.Seeders-item.Last.Seeders,
					//pi.Leechers-item.Last.Leechers, pi.Complete-item.Last.Complete)

					seeders := int64(ii.Item.Seeders - last.Seeders)
					leechers := int64(ii.Item.Leechers - last.Leechers)
					complete := int64(ii.Item.Complete - last.Complete)
					comments := int64(ii.Item.Comments - last.Comments)

					fmt.Fprintf(body, "<tr><td>%s</td><td>%d</td><td>%d (%d)</td><td>%d (%d)</td><td>%d (%d)</td><td>%d (%d)</td></tr>",
						ii.Date, ii.Hour,
						ii.Item.Seeders, seeders,
						ii.Item.Leechers, leechers,
						ii.Item.Complete, complete,
						ii.Item.Comments, comments)
				}

				last = ii.Item
			}
		}

		fmt.Fprintln(body, "</table>")

		tmpl := LoadTemplates("base")
		tmpldata := map[string]interface{}{
			"Title": config.Title,
			"Body":  template.HTML(body.String()),
		}
		err = tmpl.ExecuteTemplate(w, "base", tmpldata)
		if err != nil {
			http.Error(w, "Error processing template", 500)
			return
		}
	})

	http.HandleFunc("/chart", func(w http.ResponseWriter, r *http.Request) {
		r.ParseForm()

		if !initialCheck(w, r) {
			return
		}

		csession := config.Session.Clone()
		defer csession.Close()

		id := r.Form.Get("id")
		if id == "" {
			w.Header().Add("Content-Type", "text/html; charset=utf-8")
			http.Error(w, "ID not sent", 500)
			return
		}
		size := r.Form.Get("size")

		i := fstopinfo.NewInfo(config.Logger, csession)
		i.Database = config.Database

		var d []*fstopinfo.FSInfoHistory
		var err error

		dcache, found := detailCache.Get(id)
		if found {
			d = dcache.([]*fstopinfo.FSInfoHistory)
		}

		if d == nil {
			d, err = i.History(id, config.HistoryDays)
			if err != nil {
				w.Header().Add("Content-Type", "text/html; charset=utf-8")
				http.Error(w, err.Error(), 500)
				return
			}
			if d != nil {
				detailCache.Set(id, d, 0)
			}
		}

		if d == nil {
			w.Header().Add("Content-Type", "text/html; charset=utf-8")
			http.Error(w, "Not found", 500)
			return
		}

		p, err := plot.New()
		if err != nil {
			panic(err)
		}

		p.Title.Text = "Chart"
		p.X.Label.Text = "Time"
		p.Y.Label.Text = "Amount"

		c_seeders := make(plotter.XYs, 0)
		c_leechers := make(plotter.XYs, 0)
		c_complete := make(plotter.XYs, 0)
		c_comments := make(plotter.XYs, 0)

		var last *fstoplib.Item
		first := true
		cttotal := int32(0)

		for _, ii := range d {
			if ii.Item != nil {
				cttotal++

				if first {
					p.Title.Text = strings.TrimSpace(ii.Item.Title)
					first = false
				}

				if last != nil {
					c_seeders = append(c_seeders, struct{ X, Y float64 }{float64(cttotal), float64(ii.Item.Seeders)})
					c_leechers = append(c_leechers, struct{ X, Y float64 }{float64(cttotal), float64(ii.Item.Leechers)})
					c_complete = append(c_complete, struct{ X, Y float64 }{float64(cttotal), float64(ii.Item.Complete - last.Complete)})
					c_comments = append(c_comments, struct{ X, Y float64 }{float64(cttotal), float64(ii.Item.Comments)})
				}

				last = ii.Item
			}
		}

		w.Header().Add("Content-Type", "image/png")

		pl_seeders, err := plotter.NewLine(c_seeders)
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		pl_seeders.LineStyle.Width = vg.Length(1)
		pl_seeders.LineStyle.Color = color.RGBA{R: 255, A: 255}
		p.Add(pl_seeders)
		p.Legend.Add("Seeders", pl_seeders)

		pl_leechers, err := plotter.NewLine(c_leechers)
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		pl_leechers.LineStyle.Width = vg.Length(1)
		pl_leechers.LineStyle.Color = color.RGBA{G: 255, A: 255}
		p.Add(pl_leechers)
		p.Legend.Add("Leechers", pl_leechers)

		pl_complete, err := plotter.NewLine(c_complete)
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		pl_complete.LineStyle.Width = vg.Length(1)
		pl_complete.LineStyle.Color = color.RGBA{B: 255, A: 255}
		p.Add(pl_complete)
		p.Legend.Add("@Complete", pl_complete)

		pl_comments, err := plotter.NewLine(c_comments)
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		pl_comments.LineStyle.Width = vg.Length(1)
		pl_comments.LineStyle.Color = color.RGBA{R: 255, B: 255, A: 255}
		p.Add(pl_comments)
		p.Legend.Add("Comments", pl_comments)

		width := vg.Length(640)
		height := vg.Length(480)
		if size == "small" {
			width = vg.Length(640)
			height = vg.Length(160)
		} else if size == "short" {
			width = vg.Length(200)
			height = vg.Length(80)
			p.Title.Text = ""
			p.X.Label.Text = ""
			p.Y.Label.Text = ""
		}

		c := vgimg.PngCanvas{vgimg.New(width, height)}
		p.Draw(plot.MakeDrawArea(c))
		c.WriteTo(w)
	})

	http.HandleFunc("/res/", func(w http.ResponseWriter, r *http.Request) {
		p := strings.Replace(path.Clean(r.URL.Path), "/res/", "res/files/", 1)

		d, err := Asset(p)
		if err != nil {
			http.Error(w, "Error loading resource", 500)
			return
		}

		w.Header().Add("Content-Type", mime.TypeByExtension(path.Ext(p)))
		w.Write(d)

		//fmt.Fprintf(w, "RES %s\n", p)
	})

	http.ListenAndServe(fmt.Sprintf("localhost:%d", config.Port), nil)

}
