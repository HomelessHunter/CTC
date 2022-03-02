package main

import (
	"errors"
	"fmt"
	"regexp"

	db "github.com/HomelessHunter/CTC/db/models"
)

// func pairExist(pair string, pairs []string) bool {
// 	if i := pairSearch(pairs, pair); i < len(pairs) && i >= 0 {
// 		if pairs[i] == pair {
// 			return true
// 		}
// 	}

// 	return false
// }

func pairExist(market, pair string, alerts []db.Alert) bool {
	if len(alerts) == 0 {
		return false
	}
	alert := db.Alert{Market: market, Pair: pair}
	_, err := alert.Find(alerts)
	if err != nil {
		fmt.Printf("cannot find alert: %s", err)
		return false
	}
	return true
}

func compileRegexp() map[string]*regexp.Regexp {
	return map[string]*regexp.Regexp{
		"start":      regexp.MustCompile(`^\/(start)`),
		"alert":      regexp.MustCompile(`^\/(a|A)(lert)\s[A-Za-z]+\s[0-9]+\.[0-9]+$`),
		"disconnect": regexp.MustCompile(`^\/*(disconnect)\s*[A-Za-z]*$`),
		"splitter":   regexp.MustCompile(`\s`),
	}
}

func findDisconnectAlert(market, pair string, alerts []db.Alert) (int, error) {
	if len(alerts) == 0 {
		return -2, errors.New("pairs shouldn't be empty")
	}
	switch pair {
	case "all":
		return -1, nil
	default:
		alert := db.Alert{Market: market, Pair: pair}
		return alert.Find(alerts)
	}
}

// func pairSearch(pairs []string, pair string) int {
// 	p := sort.StringSlice(pairs)
// 	p.Sort()
// 	return p.Search(pair)
// }

// func removePair(pairs []string, index int) []string {
// 	return append(pairs[:index], pairs[index+1:]...)
// }
