package analytics

import (
	"sort"
	"strings"

	"typer/internal/model"
	"typer/internal/scoring"
)

type CharCount struct {
	Key   string
	Count int
}

type Summary struct {
	Sessions          int
	AvgGrossWPM       float64
	AvgNetWPM         float64
	AvgAccuracy       float64
	AvgErrors         float64
	ConsistencyTrend  float64
	TopErrorChars     []CharCount
	TopMissingChars   []CharCount
	TopUnexpectedChar []CharCount
}

func BuildSummary(sessions []model.SessionResult, topN int) Summary {
	if topN <= 0 {
		topN = 5
	}
	out := Summary{Sessions: len(sessions)}
	if len(sessions) == 0 {
		return out
	}

	netSamples := make([]float64, 0, len(sessions))
	errorMap := map[string]int{}
	missingMap := map[string]int{}
	extraMap := map[string]int{}

	for _, s := range sessions {
		out.AvgGrossWPM += s.Metrics.GrossWPM
		out.AvgNetWPM += s.Metrics.NetWPM
		out.AvgAccuracy += s.Metrics.Accuracy
		out.AvgErrors += float64(s.Metrics.Errors)
		netSamples = append(netSamples, s.Metrics.NetWPM)
		collectCharErrors([]rune(s.Prompt.Content), []rune(s.TypedText), errorMap, missingMap, extraMap)
	}

	n := float64(len(sessions))
	out.AvgGrossWPM = scoring.Round2(out.AvgGrossWPM / n)
	out.AvgNetWPM = scoring.Round2(out.AvgNetWPM / n)
	out.AvgAccuracy = scoring.Round2(out.AvgAccuracy / n)
	out.AvgErrors = scoring.Round2(out.AvgErrors / n)
	out.ConsistencyTrend = scoring.ConsistencyFromSamples(netSamples)
	out.TopErrorChars = topCounts(errorMap, topN)
	out.TopMissingChars = topCounts(missingMap, topN)
	out.TopUnexpectedChar = topCounts(extraMap, topN)
	return out
}

func collectCharErrors(target, typed []rune, wrong, missing, extra map[string]int) {
	n := min(len(target), len(typed))
	for i := 0; i < n; i++ {
		if target[i] != typed[i] {
			wrong[normalizeChar(target[i])]++
		}
	}
	if len(target) > n {
		for _, ch := range target[n:] {
			missing[normalizeChar(ch)]++
		}
	}
	if len(typed) > n {
		for _, ch := range typed[n:] {
			extra[normalizeChar(ch)]++
		}
	}
}

func normalizeChar(ch rune) string {
	switch ch {
	case ' ':
		return "<space>"
	case '\t':
		return "<tab>"
	case '\n':
		return "<newline>"
	default:
		return strings.ToLower(string(ch))
	}
}

func topCounts(src map[string]int, n int) []CharCount {
	items := make([]CharCount, 0, len(src))
	for k, v := range src {
		items = append(items, CharCount{Key: k, Count: v})
	}
	sort.Slice(items, func(i, j int) bool {
		if items[i].Count == items[j].Count {
			return items[i].Key < items[j].Key
		}
		return items[i].Count > items[j].Count
	})
	if len(items) > n {
		items = items[:n]
	}
	return items
}
