package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"regexp"
	"strings"
	"time"
)

func main() {
	runTime := time.Now()
	model := train("big.txt")
	correctionTime := time.Now()
	badSpelling := "beaty"
	correctedSpelling := correct(badSpelling, model)
	if correctedSpelling != badSpelling {
		fmt.Printf("%s - did you mean %s?\n", badSpelling, correctedSpelling)
	} else {
		fmt.Println(correctedSpelling)
	}
	log.Println("Time taken to correct:", time.Since(correctionTime))
	log.Println("Total time taken to load and correct:", time.Since(runTime))
}

// train - Load words in big.txt into a map called WORDS
// WORDS keeps a track of the words and the number of times they appear in the file.
func train(trainingData string) map[string]int {
	WORDS := make(map[string]int)
	pattern := regexp.MustCompile("[a-z]+")
	content, err := ioutil.ReadFile(trainingData)
	if err != nil {
		log.Fatalf("Could not find file %s. You can get a copy from http://norvig.com/big.txt", trainingData)
	}
	for _, word := range pattern.FindAllString(strings.ToLower(string(content)), -1) {
		WORDS[word]++
	}
	return WORDS
}

// correct - Most probable spelling correction for word.
func correct(word string, model map[string]int) string {
	if _, exists := model[word]; exists {
		return word
	}
	if correction := best(word, edits1, model); correction != "" {
		return correction
	}
	if correction := best(word, edits2, model); correction != "" {
		return correction
	}
	return word
}

// best - Generate possible spelling corrections for word from subset of `words` that appear in the dictionary of WORDS.
func best(word string, edits func(string, chan string), model map[string]int) string {
	channel := make(chan string, 1024*1024)
	go func() { edits(word, channel); channel <- "" }()
	maxFreq := 0
	correction := ""
	for word := range channel {
		if word == "" {
			break
		}
		if freq, exists := model[word]; exists && freq > maxFreq {
			maxFreq, correction = freq, word
		}
	}
	return correction
}

// edits1 - All edits that are one edit away from `word`.
func edits1(word string, channel chan string) {
	const alphabet = "abcdefghijklmnopqrstuvwxyz"
	type Pair struct{ leading, trailing string }
	var splits []Pair
	for i := 0; i < len(word)+1; i++ {
		splits = append(splits, Pair{word[:i], word[i:]})
	}

	for _, s := range splits {
		if len(s.trailing) > 0 {
			channel <- s.leading + s.trailing[1:]
		}
		if len(s.trailing) > 1 {
			channel <- s.leading + string(s.trailing[1]) + string(s.trailing[0]) + s.trailing[2:]
		}
		for _, character := range alphabet {
			if len(s.trailing) > 0 {
				channel <- s.leading + string(character) + s.trailing[1:]
			}
		}
		for _, character := range alphabet {
			channel <- s.leading + string(character) + s.trailing
		}
	}
}

// edits2 - All edits that are two edits away from `word`.
func edits2(word string, channel chan string) {
	channel1 := make(chan string, 1024*1024)
	go func() { edits1(word, channel1); channel1 <- "" }()
	for e1 := range channel1 {
		if e1 == "" {
			break
		}
		edits1(e1, channel)
	}
}
