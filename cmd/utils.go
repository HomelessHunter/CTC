package main

import (
	"regexp"
	"sort"
	"strings"
)

func pairExist(pair string, pairs []string) bool {
	for _, v := range pairs {
		if strings.Contains(v, pair) {
			return true
		}
	}
	return false
}

func compileRegexp() map[string]*regexp.Regexp {
	return map[string]*regexp.Regexp{
		"alert":      regexp.MustCompile(`^\/(a|A)(lert)\s[A-Za-z]+\s[0-9]+\.[0-9]+$`),
		"disconnect": regexp.MustCompile(`^\/*(disconnect)\s*[A-Za-z]*$`),
		"splitter":   regexp.MustCompile(`\s`),
	}
}

func checkDisconnectMsg(msg string, disconnectCh chan<- int, pairs []string) {
	switch msg {
	case "all":
		disconnectCh <- -1
	default:
		disconnectCh <- pairSearch(pairs, msg)
	}
}

func pairSearch(pairs []string, pair string) int {
	p := sort.StringSlice(pairs)
	p.Sort()
	return p.Search(pair)
}

func removePair(pairs []string, index int) []string {
	return append(pairs[:index], pairs[index+1:]...)
}
