package main

import (
	"strings"
	"unicode"
)

// Tokenize splits text into words
func Tokenize(text string) []string {
	fields := strings.Fields(text)
	tokens := make([]string, 0, len(fields))
	
	for _, field := range fields {
		cleaned := strings.TrimFunc(field, func(r rune) bool {
			return !unicode.IsLetter(r) && !unicode.IsNumber(r)
		})
		if len(cleaned) > 0 {
			tokens = append(tokens, strings.ToLower(cleaned))
		}
	}
	
	return tokens
}

// CalculateSimilarity computes word overlap similarity between two texts
func CalculateSimilarity(text1, text2 string) float64 {
	tokens1 := Tokenize(text1)
	tokens2 := Tokenize(text2)
	
	if len(tokens1) == 0 || len(tokens2) == 0 {
		return 0.0
	}
	
	// Create sets
	set1 := make(map[string]bool)
	for _, t := range tokens1 {
		set1[t] = true
	}
	
	set2 := make(map[string]bool)
	for _, t := range tokens2 {
		set2[t] = true
	}
	
	// Calculate intersection
	intersection := 0
	for t := range set1 {
		if set2[t] {
			intersection++
		}
	}
	
	// Calculate union
	union := len(set1) + len(set2) - intersection
	
	if union == 0 {
		return 0.0
	}
	
	return float64(intersection) / float64(union)
}
