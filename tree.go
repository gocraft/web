package web

import (
	"regexp"
	"sort"
	"strings"
)

type pathNode struct {

	// Given the next segment s, if edges[s] exists, then we'll look there first.
	edges map[string]*pathNode

	// If set, failure to match on edges will match on wildcard
	wildcard *pathNode

	// If set, and we have nothing left to match, then we match on this node
	leaves pathLeafPtrSlice

	// If true, this pathNode has a pathparam that matches the rest of the path
	matchesFullPath bool
}

// pathLeaf represents a leaf path segment that corresponds to a single route.
// For the route /admin/forums/:forum_id:\d.*/suggestions/:suggestion_id:\d.*
// We'd have wildcards = ["forum_id", "suggestion_id"]
//         and regexps = [/\d.*/, /\d.*/]
// For the route /admin/forums/:forum_id/suggestions/:suggestion_id:\d.*
// We'd have wildcards = ["forum_id", "suggestion_id"]
//         and regexps = [nil, /\d.*/]
// For the route /admin/forums/:forum_id/suggestions/:suggestion_id
// We'd have wildcards = ["forum_id", "suggestion_id"]
//         and regexps = nil
type pathLeaf struct {
	// names of wildcards that lead to this leaf. eg, ["category_id"] for the wildcard ":category_id"
	wildcards []string

	// regexps corresponding to wildcards. If a segment has regexp contraint, its entry will be nil.
	// If the route has no regexp contraints on any segments, then regexps will be nil.
	regexps []*regexp.Regexp

	// Pointer back to the route
	route *route

	// If true, this leaf has a pathparam that matches the rest of the path
	matchesFullPath bool

	// If true, there are no regexp groups in this leaf
	hasGroupMatches bool
}

func newPathNode() *pathNode {
	return &pathNode{edges: make(map[string]*pathNode)}
}

func (pn *pathNode) add(path string, route *route) {
	pn.addInternal(splitPath(path), route, nil, nil)
}

func (pn *pathNode) addInternal(segments []string, route *route, wildcards []string, regexps []*regexp.Regexp) {
	if len(segments) == 0 {
		allNilRegexps := true
		hasGroupMatches := false
		for _, r := range regexps {
			if r != nil {
				allNilRegexps = false
				if r.NumSubexp() > 0 {
					hasGroupMatches = true
				}
			}
		}
		if allNilRegexps {
			regexps = nil
		}

		matchesFullPath := false
		if len(wildcards) > 0 {
			matchesFullPath = wildcards[len(wildcards)-1] == "*"
		}

		pn.leaves = append(pn.leaves, &pathLeaf{route: route, wildcards: wildcards, regexps: regexps, matchesFullPath: matchesFullPath, hasGroupMatches: hasGroupMatches})
		sort.Stable(pn.leaves)
	} else { // len(segments) >= 1
		seg := segments[0]
		wc, wcName, wcRegexpStr := isWildcard(seg)
		if wc {

			if pn.wildcard == nil {
				pn.wildcard = newPathNode()
			}

			if !pn.wildcard.matchesFullPath {
				pn.wildcard.matchesFullPath = wcName == "*"
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

// A slice of path leafs that supports sorting via the sort.Interface interface.
type pathLeafPtrSlice []*pathLeaf

func (this pathLeafPtrSlice) Len() int {
	return len(this)
}

func (this pathLeafPtrSlice) Less(i, j int) bool {
	// If there are more regexp patterns to match, then we're less. We want the
	// ones with regexps to be checked first.
	return len(this[i].regexps) > len(this[j].regexps)
}

func (this pathLeafPtrSlice) Swap(i, j int) {
	this[i], this[j] = this[j], this[i]
}

func (pn *pathNode) Match(path string) (leaf *pathLeaf, wildcards map[string]string) {

	// Bail on invalid paths.
	if len(path) == 0 || path[0] != '/' {
		return nil, nil
	}

	return pn.match(splitPath(path), nil)
}

// Segments is like ["admin", "users"] representing "/admin/users"
// wildcardValues are the actual values accumulated when we match on a wildcard.
func (pn *pathNode) match(segments []string, wildcardValues []string) (leaf *pathLeaf, wildcardMap map[string]string) {
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

	if leaf == nil && pn.matchesFullPath {
		for _, leaf := range pn.leaves {
			if leaf.matchesFullPath && leaf.match(wildcardValues) {
				if len(wildcardValues) > 0 {
					wcVals := []string{wildcardValues[len(wildcardValues)-1], seg}
					for _, s := range segments {
						wcVals = append(wcVals, s)
					}
					wildcardValues[len(wildcardValues)-1] = strings.Join(wcVals, "/")
				}
				return leaf, makeWildcardMap(leaf, wildcardValues)
			}
		}
		return nil, nil
	}

	return leaf, wildcardMap
}

func (leaf *pathLeaf) match(wildcardValues []string) bool {
	if leaf.regexps == nil {
		return true
	}

	// Invariant:
	if len(leaf.regexps) != len(wildcardValues) {
		panic("bug: invariant violated")
	}

	if leaf.hasGroupMatches {
		// Find all the match groups
		l := len(wildcardValues)
		matchedGroups := make([]string, l)
		for i, r := range leaf.regexps {
			if r != nil {
				// If there are no groups, check a simple match.
				if r.NumSubexp() == 0 {
					if !r.MatchString(wildcardValues[i]) {
						return false
					} else {
						matchedGroups[i] = wildcardValues[i]
					}
				} else {
					// Otherwise, there is a group. If we match, use the first group.
					match := r.FindStringSubmatch(wildcardValues[i])
					if match != nil {
						matchedGroups[i] = match[1]
					} else {
						return false
					}
				}
			}
		}

		// If we made it this far, set our values in.
		for n := 0; n < l; n++ {
			wildcardValues[n] = matchedGroups[n]
		}
	} else {
		for i, r := range leaf.regexps {
			if r != nil {
				if !r.MatchString(wildcardValues[i]) {
					return false
				}
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
		}

		return true, substrs[0], substrs[1]
	}

	return false, "", ""
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

func makeWildcardMap(leaf *pathLeaf, wildcards []string) map[string]string {
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
