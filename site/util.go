package fstopsite

import (
	"fmt"
	"html/template"
	"time"
)

func LoadTemplates(names ...string) *template.Template {
	t := template.New("server")
	template.Must(t.New("header").Parse(AssetLoader("res/header.tpl")))
	template.Must(t.New("footer").Parse(AssetLoader("res/footer.tpl")))
	for _, name := range names {
		template.Must(t.New(name).Parse(AssetLoader(fmt.Sprintf("res/%s.tpl", name))))
	}
	return t
}

func FormatAddDate(date string) string {
	if date == "" {
		return date
	}

	ad, err := time.Parse("2006-01-02", date)
	if err != nil {
		return date
	}

	add := time.Now().Sub(ad)

	hours := add.Hours()
	days := float64(0)
	months := float64(0)
	years := float64(0)

	for hours >= 8765.81 {
		years++
		hours -= 8765.81
	}

	for hours >= 730.484 {
		months++
		hours -= 730.484
	}

	for hours >= 24 {
		days++
		hours -= 24
	}

	if years > 0 {
		return fmt.Sprintf("%d year(s)", int32(years))
	}
	if months > 0 {
		return fmt.Sprintf("%d month(s)", int32(months))
	}
	if days > 0 {
		return fmt.Sprintf("%d day(s)", int32(days))
	}
	return fmt.Sprintf("%d hour(s)", int32(hours))

	//return fmt.Sprintf("%dY %dM %dD %dH", int32(years), int32(months), int32(days), int32(hours))
}

func StyleAddDate(date string) string {
	if date == "" {
		return date
	}

	ad, err := time.Parse("2006-01-02", date)
	if err != nil {
		return date
	}

	add := time.Now().Sub(ad)
	if add.Hours() < 720 {
		return "color: #00c000;"
	}
	return ""
}
