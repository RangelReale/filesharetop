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
	"html/template"
	"image/color"
	"mime"
	"net/http"
	"path"
	"strings"
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

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		csession := config.Session.Clone()
		defer csession.Close()

		r.ParseForm()

		cat := r.Form.Get("category")
		chart := r.Form.Get("chart")

		i := fstopinfo.NewInfo(config.Logger, csession)
		i.Database = config.Database

		var d []*fstopimp.FSTopStats
		var err error

		if cat == "" {
			d, err = i.Top(config.TopId)
		} else {
			d, err = i.TopCategory(config.TopId, cat)
		}
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}

		w.Header().Add("Content-Type", "text/html; charset=utf-8")

		var body *bytes.Buffer = new(bytes.Buffer)

		fmt.Fprintln(body, "<table class=\"main\">")

		fmt.Fprintln(body, "<tr><th>Title</th><th width=\"8%\">Added</th><th width=\"5%\">Score</th><th width=\"5%\">Count</th><th width=\"5%\">Comm.</th></tr>")

		for _, ii := range d {
			fmt.Fprintf(body, "<tr><td><a href=\"%s\">%s</a> <a href=\"/view?id=%s\">[data]</a> <a href=\"/chart?id=%s\">[chart]</a></td>"+
				"<td align=\"center\">%s</td>"+
				"<td align=\"right\">%d</td><td align=\"center\">%d</td><td align=\"center\">%d</td>\n",
				ii.Link, ii.Title, ii.Id, ii.Id, FormatAddDate(ii.Last.AddDate), ii.Score, ii.Count, ii.Last.Comments)

			if chart == "2" {
				fmt.Fprintf(body, "<td><img src=\"/chart?id=%s&size=short\"></td>", ii.Id)
			}

			fmt.Fprintf(body, "</tr>\n")

			if chart == "1" {
				fmt.Fprintf(body, "<tr><td colspan=\"5\" align=\"center\"><img src=\"/chart?id=%s&size=small\"/></td></tr>\n", ii.Id)
			}
		}
		fmt.Fprintln(body, "</table>")

		tmpl := LoadTemplates("index")
		tmpldata := map[string]interface{}{
			"Title": config.Title,
			"Body":  template.HTML(body.String()),
		}
		err = tmpl.ExecuteTemplate(w, "index", tmpldata)
		if err != nil {
			http.Error(w, "Error processing template", 500)
			return
		}

	})

	http.HandleFunc("/view", func(w http.ResponseWriter, r *http.Request) {
		csession := config.Session.Clone()
		defer csession.Close()

		r.ParseForm()

		w.Header().Add("Content-Type", "text/html; charset=utf-8")

		id := r.Form.Get("id")
		if id == "" {
			http.Error(w, "ID not sent", 500)
			return
		}

		i := fstopinfo.NewInfo(config.Logger, csession)
		i.Database = config.Database

		d, err := i.History(id, config.HistoryDays)
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
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

		tmpl := LoadTemplates("index")
		tmpldata := map[string]interface{}{
			"Title": config.Title,
			"Body":  template.HTML(body.String()),
		}
		err = tmpl.ExecuteTemplate(w, "index", tmpldata)
		if err != nil {
			http.Error(w, "Error processing template", 500)
			return
		}
	})

	http.HandleFunc("/chart", func(w http.ResponseWriter, r *http.Request) {
		csession := config.Session.Clone()
		defer csession.Close()

		r.ParseForm()

		id := r.Form.Get("id")
		if id == "" {
			w.Header().Add("Content-Type", "text/html; charset=utf-8")
			http.Error(w, "ID not sent", 500)
			return
		}
		size := r.Form.Get("size")

		i := fstopinfo.NewInfo(config.Logger, csession)
		i.Database = config.Database

		d, err := i.History(id, config.HistoryDays)
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
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
					c_complete = append(c_complete, struct{ X, Y float64 }{float64(cttotal), float64(ii.Item.Complete)})
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

		/*
			pl_complete, err := plotter.NewLine(c_complete)
			if err != nil {
				http.Error(w, err.Error(), 500)
				return
			}
			pl_complete.LineStyle.Width = vg.Length(1)
			pl_complete.LineStyle.Color = color.RGBA{B: 255, A: 255}
			p.Add(pl_complete)
			p.Legend.Add("Complete", pl_complete)
		*/

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
