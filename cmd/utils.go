package main

import (
	"regexp"
	"sort"
)

func pairExist(pair string, pairs []string) bool {
	if i := pairSearch(pairs, pair); i < len(pairs) && i >= 0 {
		if pairs[i] == pair {
			return true
		}
	}

	return false
}

func compileRegexp() map[string]*regexp.Regexp {
	return map[string]*regexp.Regexp{
		"start":      regexp.MustCompile(`^\/(start)`),
		"alert":      regexp.MustCompile(`^\/(a|A)(lert)\s[A-Za-z]+\s[0-9]+\.[0-9]+$`),
		"disconnect": regexp.MustCompile(`^\/*(disconnect)\s*[A-Za-z]*$`),
		"splitter":   regexp.MustCompile(`\s`),
	}
}

func checkDisconnectMsg(msg string, pairs []string) int {
	switch msg {
	case "all":
		return -1
	default:
		return pairSearch(pairs, msg)
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
