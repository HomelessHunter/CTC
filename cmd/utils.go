package main

import (
	"errors"
	"fmt"
	"regexp"

	db "github.com/HomelessHunter/CTC/db/models"
)

func pairExist(market, pair string, alerts []db.Alert) bool {
	if len(alerts) == 0 {
		return false
	}
	alert, err := db.NewAlert(db.WithMarket(market), db.WithPair(pair))
	if err != nil {
		fmt.Println(err)
		return false
	}
	_, err = alert.SortNFind(alerts)
	if err != nil {
		return false
	}
	return true
}

func compileRegexp() map[string]*regexp.Regexp {
	return map[string]*regexp.Regexp{
		"start":      regexp.MustCompile(`^\/(start)$`),
		"help":       regexp.MustCompile(`^\/help$`),
		"alert":      regexp.MustCompile(`^\/(a|A)(lert)\s[A-Za-z]+\s[0-9]+\.*[0-9]*$`),
		"price":      regexp.MustCompile(`^\/(p|P)rice\s[a-zA-Z]+$`),
		"disconnect": regexp.MustCompile(`^\/*(disconnect)\s*[A-Za-z]*\s*[A-Za-z]*$`),
		"splitter":   regexp.MustCompile(`\s`),
	}
}

func findDisconnectAlert(market, pair string, alerts []db.Alert) (int, error) {
	if len(alerts) == 0 {
		return -2, errors.New("pairs shouldn't be empty")
	}

	if pair == "all" || market == "all" {
		return -1, nil
	}
	alert, err := db.NewAlert(db.WithMarket(market), db.WithPair(pair))
	if err != nil {
		return -2, err
	}
	return alert.SortNFind(alerts)
}

// func pairSearch(pairs []string, pair string) int {
// 	p := sort.StringSlice(pairs)
// 	p.Sort()
// 	return p.Search(pair)
// }

// func removePair(pairs []string, index int) []string {
// 	return append(pairs[:index], pairs[index+1:]...)
// }
