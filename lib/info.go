package fstoplib

type Item struct {
	Id    string
	Title string
	// Complete link including host
	Link     string
	Category string
	// YYYY-MM-DD
	AddDate string
	// Amount of times found
	Count int32

	// The following values should be -1 if not available
	Seeders     int32
	SeedersPos  int32
	Leechers    int32
	LeechersPos int32
	Complete    int32
	CompletePos int32
	Comments    int32
}

type CategoryMap map[string][]string

func NewItem() *Item {
	return &Item{
		Count:       0,
		Seeders:     -1,
		SeedersPos:  -1,
		Leechers:    -1,
		LeechersPos: -1,
		Complete:    -1,
		CompletePos: -1,
		Comments:    -1,
	}
}
