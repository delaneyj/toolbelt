package datalog

import (
	"iter"
	"strings"
)

type Triple = [3]string
type Pattern = [3]string

func NewTriple(subject, predicate, object string) Triple {
	return Triple{subject, predicate, object}
}

type State map[string]string

func isVariable(s string) bool {
	return strings.HasPrefix(s, "?")
}

func deepCopyState(state State) State {
	newState := make(State, len(state)+1)
	for key, value := range state {
		newState[key] = value
	}
	return newState
}

func matchVariable(variable, triplePart string, state State) State {
	bound, ok := state[variable]
	if ok {
		return matchPart(bound, triplePart, state)
	}
	newState := deepCopyState(state)
	newState[variable] = triplePart
	return newState
}

func matchPart(patternPart, triplePart string, state State) State {
	if state == nil {
		return nil
	}
	if isVariable(patternPart) {
		return matchVariable(patternPart, triplePart, state)
	}
	if patternPart == triplePart {
		return state
	}
	return nil
}

func MatchPattern(pattern Pattern, triple Triple, state State) State {
	newState := deepCopyState(state)

	for idx, patternPart := range pattern {
		triplePart := triple[idx]
		newState = matchPart(patternPart, triplePart, newState)
		if newState == nil {
			return nil
		}
	}

	return newState
}

func (db *DB) QuerySingle(state State, pattern Pattern) (valid []State) {
	for triple := range relevantTriples(db, pattern) {
		newState := MatchPattern(pattern, triple, state)
		if newState != nil {
			valid = append(valid, newState)
		}
	}
	return valid
}

func (db *DB) QueryWhere(where ...Pattern) []State {
	states := []State{{}}
	for _, pattern := range where {
		revised := make([]State, 0, len(states))
		for _, state := range states {
			revised = append(revised, db.QuerySingle(state, pattern)...)
		}
		states = revised
	}
	return states
}

func (db *DB) Query(find []string, where ...Pattern) [][]string {
	states := db.QueryWhere(where...)

	results := make([][]string, len(states))
	for i, state := range states {
		results[i] = actualize(state, find...)
	}
	return results
}

func actualize(state State, find ...string) []string {
	results := make([]string, len(find))
	for i, findPart := range find {
		r := findPart
		if isVariable(findPart) {
			r = state[findPart]
		}
		results[i] = r
	}
	return results
}

type DB struct {
	triples     []Triple
	entityIndex map[string][]Triple
	attrIndex   map[string][]Triple
	valueIndex  map[string][]Triple
}

func CreateDB(triples ...Triple) *DB {
	return &DB{
		triples:     triples,
		entityIndex: indexBy(triples, 0),
		attrIndex:   indexBy(triples, 1),
		valueIndex:  indexBy(triples, 2),
	}
}

func indexBy(triples []Triple, idx int) map[string][]Triple {
	index := map[string][]Triple{}
	for _, triple := range triples {
		key := triple[idx]
		index[key] = append(index[key], triple)
	}
	return index
}

func relevantTriples(db *DB, pattern Pattern) iter.Seq[Triple] {
	return func(yield func(Triple) bool) {
		id, attr, value := pattern[0], pattern[1], pattern[2]
		if !isVariable(id) {
			for _, triple := range db.entityIndex[id] {
				if !yield(triple) {
					return
				}
			}
			return
		}
		if !isVariable(attr) {
			for _, triple := range db.attrIndex[attr] {
				if !yield(triple) {
					return
				}
			}
			return
		}
		if !isVariable(value) {
			for _, triple := range db.valueIndex[value] {
				if !yield(triple) {
					return
				}
			}
			return
		}

		for _, triple := range db.triples {
			if !yield(triple) {
				return
			}
		}
	}
}
