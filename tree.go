package web

import (
  "strings"
)

type PathNode struct {
  edges    map[string]*PathNode
  
  leaf *PathLeaf // If set, and we have nothing left to match, then we match on this node
  
  wildcard *PathNode    // If set, failure to match on edges will match on wildcard
  
  // Router/Namespace
}

type PathLeaf struct {
  wildcards []string // names of wildcards that lead to this leaf. eg, ["category_id"] for the wildcard ":category_id"
  route *Route
}

func newPathNode() *PathNode {
  return &PathNode{edges: make(map[string]*PathNode)}
}

func (pn *PathNode) add(path string, route *Route) {
  pn.addInternal(splitPath(path), route, nil)
}

func (pn *PathNode) addInternal(segments []string, route *Route, wildcards []string) {
  if len(segments) == 0 {
    if pn.leaf == nil {
      pn.leaf = &PathLeaf{route: route, wildcards: wildcards}
    } else {
      panic("there's already a handler at this node")
    }
  } else { // len(segments) >= 1
    seg := segments[0]
    wc, wcName := isWildcard(seg)
    if wc {
      if pn.wildcard == nil {
        pn.wildcard = newPathNode()
      }
      pn.wildcard.addInternal(segments[1:], route, append(wildcards, wcName))
    } else {
      subPn, ok := pn.edges[seg]
      if !ok {
        subPn = newPathNode()
        pn.edges[seg] = subPn
      }
      subPn.addInternal(segments[1:], route, wildcards)
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
// wildcards are the actual values accumulated when we match on a wildcard.
func (pn *PathNode) match(segments []string, wildcardValues []string) (leaf *PathLeaf, wildcardMap map[string]string) {
  // Handle leaf nodes:
  if len(segments) == 0 {
    return pn.leaf, makeWildcardMap(pn.leaf, wildcardValues)
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

// key is a non-empty path segment like "admin" or ":category_id"
// Returns true if it's a wildcard, and if it is, also returns it's name. Eg, (true, "category_id")
func isWildcard(key string) (bool, string) {
  if key[0] == ':' {
    return true, key[1:]
  } else {
    return false, ""
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


