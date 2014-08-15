package fstoplib

type Item struct {
	Id    string `bson:"id"`
	Title string `bson:"title"`
	// Complete link including host
	Link     string `bson:"link"`
	Category string `bson:"category"`
	// YYYY-MM-DD
	AddDate string `bson:"adddate"`
	// Amount of times found
	Count int32 `bson:"count"`

	// The following values should be -1 if not available
	Seeders     int32 `bson:"seeders"`
	SeedersPos  int32 `bson:"seederspos"`
	Leechers    int32 `bson:"leechers"`
	LeechersPos int32 `bson:"leecherspos"`
	Complete    int32 `bson:"complete"`
	CompletePos int32 `bson:"completepos"`
	Comments    int32 `bson:"comments"`
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
