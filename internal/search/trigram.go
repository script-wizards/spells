package search

import (
	"math"
	"sort"
	"strings"
)

type Index struct {
	trigrams map[string][]int
	terms    []string
}

type Match struct {
	Value string
	Score float64
}

func BuildIndex(terms []string) Index {
	idx := Index{
		trigrams: make(map[string][]int),
		terms:    make([]string, len(terms)),
	}

	copy(idx.terms, terms)

	for i, term := range terms {
		trigrams := extractTrigrams(strings.ToLower(term))
		for _, trigram := range trigrams {
			idx.trigrams[trigram] = append(idx.trigrams[trigram], i)
		}
	}

	return idx
}

func Query(idx Index, query string, limit int) []Match {
	if len(idx.terms) == 0 || query == "" {
		return nil
	}

	queryTrigrams := extractTrigrams(strings.ToLower(query))
	if len(queryTrigrams) == 0 {
		return nil
	}

	scores := make([]float64, len(idx.terms))
	queryVector := buildVector(queryTrigrams)

	for i, term := range idx.terms {
		termTrigrams := extractTrigrams(strings.ToLower(term))
		termVector := buildVector(termTrigrams)
		scores[i] = cosineSimilarity(queryVector, termVector)
	}

	matches := make([]Match, len(idx.terms))
	for i, score := range scores {
		matches[i] = Match{
			Value: idx.terms[i],
			Score: score,
		}
	}

	sort.Slice(matches, func(i, j int) bool {
		return matches[i].Score > matches[j].Score
	})

	if limit > 0 && limit < len(matches) {
		matches = matches[:limit]
	}

	return matches
}

func extractTrigrams(text string) []string {
	if len(text) < 3 {
		return []string{text}
	}

	trigrams := make([]string, 0, len(text)-2)
	for i := 0; i <= len(text)-3; i++ {
		trigrams = append(trigrams, text[i:i+3])
	}
	return trigrams
}

func buildVector(trigrams []string) map[string]int {
	vector := make(map[string]int)
	for _, trigram := range trigrams {
		vector[trigram]++
	}
	return vector
}

func cosineSimilarity(a, b map[string]int) float64 {
	if len(a) == 0 || len(b) == 0 {
		return 0.0
	}

	var dotProduct, normA, normB float64

	for trigram, countA := range a {
		if countB, exists := b[trigram]; exists {
			dotProduct += float64(countA * countB)
		}
		normA += float64(countA * countA)
	}

	for _, countB := range b {
		normB += float64(countB * countB)
	}

	if normA == 0 || normB == 0 {
		return 0.0
	}

	return dotProduct / (math.Sqrt(normA) * math.Sqrt(normB))
}
