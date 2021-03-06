package fstopimp

import (
	"github.com/RangelReale/filesharetop/lib"
)

type ScoreCalculator interface {
	CalcScore(date string, hour int, current *fstoplib.Item, previous *fstoplib.Item) int32
	FinishScore(id string, score int32, missed int32, total int32) int32
}

type DefaultScoreCalculator struct {
}

func (c *DefaultScoreCalculator) CalcScore(date string, hour int, current *fstoplib.Item, previous *fstoplib.Item) int32 {
	score := int32(0)

	seeders := int32(current.Seeders - previous.Seeders)
	leechers := int32(current.Leechers - previous.Leechers)
	complete := int32(current.Complete - previous.Complete)
	comments := int32(current.Comments - previous.Comments)

	if seeders >= 0 {
		score += seeders * 5
	} else {
		score += seeders * 2
	}

	if leechers >= 0 {
		score += leechers * 3
	} else {
		score += leechers * 1
	}

	if complete >= 0 {
		score += complete * 3
	} else {
		score += complete * 1
	}

	if comments > 0 {
		score += comments * 10
	}

	return score
}

func (c *DefaultScoreCalculator) FinishScore(id string, score int32, missed int32, total int32) int32 {
	if missed > 0 && total > 0 {
		newscore := float64(score) * (float64(total-missed) / float64(total))
		score = int32(newscore)
	}
	return score
}
