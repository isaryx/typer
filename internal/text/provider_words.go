package text

import (
	"context"
	"fmt"
	"math/rand/v2"
	"strings"
	"time"
	"unicode/utf8"

	"typer/internal/model"
)

const defaultWordCount = 15

type practiceBucket int

const (
	bucketShort practiceBucket = iota
	bucketCore
	bucketMid
	bucketLong
	practiceBucketCount
)

var practiceBucketRatios = [practiceBucketCount]float64{0.20, 0.40, 0.27, 0.13}

// nearestBucketSearchOrder lists buckets to try when redistributing slots from an empty bucket.
var nearestBucketSearchOrder = [practiceBucketCount][practiceBucketCount - 1]practiceBucket{
	bucketShort: {bucketCore, bucketMid, bucketLong},
	bucketCore:  {bucketShort, bucketMid, bucketLong},
	bucketMid:   {bucketCore, bucketLong, bucketShort},
	bucketLong:  {bucketMid, bucketCore, bucketShort},
}

type bucketState struct {
	wordIdx []int
	pos     int
}

type WordsProvider struct {
	words    []string
	sourceID string
	buckets  [practiceBucketCount]bucketState
}

func NewWordsProvider(wordsFile string) (*WordsProvider, error) {
	corpus, err := LoadWordCorpus(wordsFile)
	if err != nil {
		return nil, err
	}
	return newWordsProviderFromCorpus(corpus)
}

func newWordsProviderFromCorpus(corpus WordCorpus) (*WordsProvider, error) {
	if len(corpus.Words) == 0 {
		return nil, fmt.Errorf("word list is empty")
	}
	p := &WordsProvider{
		words:    corpus.Words,
		sourceID: corpus.SourceID,
	}
	for i, w := range corpus.Words {
		b, ok := practiceBucketForWord(w)
		if !ok {
			continue
		}
		p.buckets[b].wordIdx = append(p.buckets[b].wordIdx, i)
	}
	for b := range practiceBucketCount {
		if len(p.buckets[b].wordIdx) > 0 {
			return p, nil
		}
	}
	return nil, fmt.Errorf("word list has no words in lengths 3–12")
}

func practiceBucketForWord(w string) (practiceBucket, bool) {
	switch n := utf8.RuneCountInString(w); {
	case n >= 3 && n <= 4:
		return bucketShort, true
	case n >= 5 && n <= 6:
		return bucketCore, true
	case n >= 7 && n <= 8:
		return bucketMid, true
	case n >= 9 && n <= 12:
		return bucketLong, true
	default:
		return 0, false
	}
}

func practiceSlotCounts(n int) [practiceBucketCount]int {
	var counts [practiceBucketCount]int
	sum := 0
	for b := range practiceBucketCount {
		counts[b] = int(float64(n)*practiceBucketRatios[b] + 0.5)
		sum += counts[b]
	}
	counts[bucketCore] += n - sum
	return counts
}

func (p *WordsProvider) effectiveSlotCounts(n int) [practiceBucketCount]int {
	counts := practiceSlotCounts(n)
	for b := range practiceBucketCount {
		if counts[b] == 0 || len(p.buckets[b].wordIdx) > 0 {
			continue
		}
		slots := counts[b]
		counts[b] = 0
		recipient, ok := p.nearestNonEmptyBucket(b)
		if !ok {
			continue
		}
		counts[recipient] += slots
	}
	return counts
}

func (p *WordsProvider) nearestNonEmptyBucket(from practiceBucket) (practiceBucket, bool) {
	for _, b := range nearestBucketSearchOrder[from] {
		if len(p.buckets[b].wordIdx) > 0 {
			return b, true
		}
	}
	return 0, false
}

func (p *WordsProvider) Name() string {
	return p.sourceID
}

func (p *WordsProvider) refillBucket(b practiceBucket) {
	state := &p.buckets[b]
	if len(state.wordIdx) == 0 {
		return
	}
	rand.Shuffle(len(state.wordIdx), func(i, j int) {
		state.wordIdx[i], state.wordIdx[j] = state.wordIdx[j], state.wordIdx[i]
	})
	state.pos = 0
}

func (p *WordsProvider) drawFromBucket(b practiceBucket) (string, bool) {
	state := &p.buckets[b]
	if len(state.wordIdx) == 0 {
		return "", false
	}
	if state.pos >= len(state.wordIdx) {
		p.refillBucket(b)
	}
	word := p.words[state.wordIdx[state.pos]]
	state.pos++
	return word, true
}

func (p *WordsProvider) Next(_ context.Context, c Constraints) (model.Prompt, error) {
	n := c.Words
	if n <= 0 {
		n = defaultWordCount
	}

	counts := p.effectiveSlotCounts(n)
	picked := make([]string, 0, n)
	for b := range practiceBucketCount {
		for range counts[b] {
			w, ok := p.drawFromBucket(b)
			if !ok {
				return model.Prompt{}, fmt.Errorf("no words available in bucket %d", b)
			}
			picked = append(picked, w)
		}
	}

	rand.Shuffle(len(picked), func(i, j int) {
		picked[i], picked[j] = picked[j], picked[i]
	})

	return model.Prompt{
		ID:      time.Now().UTC().Format(time.RFC3339Nano),
		Content: strings.Join(picked, " "),
		Source:  p.Name(),
		Mode:    model.ModeWords,
	}, nil
}
