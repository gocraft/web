package web

import (
	"regexp"
	"strings"
)

type PathNode struct {

	// Given the next segment s, if edges[s] exists, then we'll look there first.
	edges map[string]*PathNode

	// If set, failure to match on edges will match on wildcard
	wildcard *PathNode

	// If set, and we have nothing left to match, then we match on this node
	leaves []*PathLeaf
}

// For the route /admin/forums/:forum_id:\d.*/suggestions/:suggestion_id:\d.*
// We'd have wildcards = ["forum_id", "suggestion_id"]
//         and regexps = [/\d.*/, /\d.*/]
// For the route /admin/forums/:forum_id/suggestions/:suggestion_id:\d.*
// We'd have wildcards = ["forum_id", "suggestion_id"]
//         and regexps = [nil, /\d.*/]
// For the route /admin/forums/:forum_id/suggestions/:suggestion_id
// We'd have wildcards = ["forum_id", "suggestion_id"]
//         and regexps = nil
type PathLeaf struct {
	// names of wildcards that lead to this leaf. eg, ["category_id"] for the wildcard ":category_id"
	wildcards []string

	// regexps corresponding to wildcards. If a segment has regexp contraint, its entry will be nil.
	// If the route has no regexp contraints on any segments, then regexps will be nil.
	regexps []*regexp.Regexp

	// Pointer back to the route
	route *Route
}

func newPathNode() *PathNode {
	return &PathNode{edges: make(map[string]*PathNode)}
}

func (pn *PathNode) add(path string, route *Route) {
	pn.addInternal(splitPath(path), route, nil, nil)
}

func (pn *PathNode) addInternal(segments []string, route *Route, wildcards []string, regexps []*regexp.Regexp) {
	if len(segments) == 0 {
		allNilRegexps := true
		for _, r := range regexps {
			if r != nil {
				allNilRegexps = false
				break
			}
		}
		if allNilRegexps {
			regexps = nil
		}
		pn.leaves = append(pn.leaves, &PathLeaf{route: route, wildcards: wildcards, regexps: regexps})
		// TODO: ? detect if we have duplicate leaves. (eg, 2 routes that are exactly the same)
	} else { // len(segments) >= 1
		seg := segments[0]
		wc, wcName, wcRegexpStr := isWildcard(seg)
		if wc {
			if pn.wildcard == nil {
				pn.wildcard = newPathNode()
			}
			pn.wildcard.addInternal(segments[1:], route, append(wildcards, wcName), append(regexps, compileRegexp(wcRegexpStr)))
		} else {
			subPn, ok := pn.edges[seg]
			if !ok {
				subPn = newPathNode()
				pn.edges[seg] = subPn
			}
			subPn.addInternal(segments[1:], route, wildcards, regexps)
		}
	}
}

func (pn *PathNode) Match(path string) (leaf *PathLeaf, wildcards map[string]string) {

	// Bail on invalid paths.
	if len(path) == 0 || path[0] != '/' {
		return nil, nil
	}

	return pn.match(splitPath(path), nil)
}

// Segments is like ["admin", "users"] representing "/admin/users"
// wildcardValues are the actual values accumulated when we match on a wildcard.
func (pn *PathNode) match(segments []string, wildcardValues []string) (leaf *PathLeaf, wildcardMap map[string]string) {
	// Handle leaf nodes:
	if len(segments) == 0 {
		for _, leaf := range pn.leaves {
			if leaf.match(wildcardValues) {
				return leaf, makeWildcardMap(leaf, wildcardValues)
			}
		}
		return nil, nil
	}

	var seg string
	seg, segments = segments[0], segments[1:]

	subPn, ok := pn.edges[seg]
	if ok {
		leaf, wildcardMap = subPn.match(segments, wildcardValues)
	}

	if leaf == nil && pn.wildcard != nil {
		leaf, wildcardMap = pn.wildcard.match(segments, append(wildcardValues, seg))
	}

	return leaf, wildcardMap
}

func (leaf *PathLeaf) match(wildcardValues []string) bool {
	if leaf.regexps == nil {
		return true
	}

	// Invariant:
	if len(leaf.regexps) != len(wildcardValues) {
		panic("bug of some sort")
	}

	for i, r := range leaf.regexps {
		if r != nil {
			if !r.MatchString(wildcardValues[i]) {
				return false
			}
		}
	}
	return true
}

// key is a non-empty path segment like "admin" or ":category_id" or ":category_id:\d+"
// Returns true if it's a wildcard, and if it is, also returns it's name / regexp.
// Eg, (true, "category_id", "\d+")
func isWildcard(key string) (bool, string, string) {
	if key[0] == ':' {
		substrs := strings.SplitN(key[1:], ":", 2)
		if len(substrs) == 1 {
			return true, substrs[0], ""
		} else {
			return true, substrs[0], substrs[1]
		}
	} else {
		return false, "", ""
	}
}

// "/" -> []
// "/admin" -> ["admin"]
// "/admin/" -> ["admin"]
// "/admin/users" -> ["admin", "users"]
func splitPath(key string) []string {
	elements := strings.Split(key, "/")
	if elements[0] == "" {
		elements = elements[1:]
	}
	if elements[len(elements)-1] == "" {
		elements = elements[:len(elements)-1]
	}
	return elements
}

func makeWildcardMap(leaf *PathLeaf, wildcards []string) map[string]string {
	if leaf == nil {
		return nil
	}

	leafWildcards := leaf.wildcards

	if len(wildcards) == 0 || (len(leafWildcards) != len(wildcards)) {
		return nil
	}

	// At this point, we know that wildcards and leaf.wildcards match in length.
	assoc := make(map[string]string)
	for i, w := range wildcards {
		assoc[leafWildcards[i]] = w
	}

	return assoc
}

func compileRegexp(regStr string) *regexp.Regexp {
	if regStr == "" {
		return nil
	}

	return regexp.MustCompile("^" + regStr + "$")
}
