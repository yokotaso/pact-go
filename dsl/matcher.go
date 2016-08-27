package dsl

// Example matching rule / generated doc
// {
//     "method": "POST",
//     "path": "/",
//     "query": "",
//     "headers": {"Content-Type": "application/json"},
//     "matchingRules": {
//       "$.body.animals": {"min": 1, "match": "type"},
//       "$.body.animals[*].*": {"match": "type"},
//       "$.body.animals[*].children": {"min": 1, "match": "type"},
//       "$.body.animals[*].children[*].*": {"match": "type"}
//     },
//     "body": {
//       "animals": [
//         {
//           "name" : "Fred",
//           "children": [
//             {
//               "age": 9
//             }
//           ]
//         }
//       ]
//     }
// 	}

// Matcher types supported by JVM:
//
// method	                    description
// string, stringValue				Match a string value (using string equality)
// number, numberValue				Match a number value (using Number.equals)*
// booleanValue								Match a boolean value (using equality)
// stringType									Will match all Strings
// numberType									Will match all numbers*
// integerType								Will match all numbers that are integers (both ints and longs)*
// decimalType								Will match all real numbers (floating point and decimal)*
// booleanType								Will match all boolean values (true and false)
// stringMatcher							Will match strings using the provided regular expression
// timestamp									Will match string containing timestamps. If a timestamp format is not given, will match an ISO timestamp format
// date												Will match string containing dates. If a date format is not given, will match an ISO date format
// time												Will match string containing times. If a time format is not given, will match an ISO time format
// ipAddress									Will match string containing IP4 formatted address.
// id													Will match all numbers by type
// hexValue										Will match all hexadecimal encoded strings
// uuid												Will match strings containing UUIDs

// RULES I'd like to follow:
// 0. Allow the option of string bodies for simple things
// 1. Have all of the matchers deal with interfaces{} for their values (or a Matcher/Builder type interface)
//    - Interfaces may turn out to be primitives like strings, ints etc. (valid JSON values I guess)
// 2. Make all matcher values serialise as map[string]interface{} to be able to easily convert to JSON,
//    and allows simpler interspersing of builder logic
//    - can we embed builders in maps??
// 3. Keep the matchers/builders simple, and orchestrate them from another class/func/place
//    Candidates are:
//    - Interaction
//    - Some new DslBuilder thingo
import (
	"fmt"
	"log"
	"strconv"
)

// matcherType is essentially a key value JSON pairs for serialisation
type matcherType map[string]interface{}

// Matching Rule
type matchingRuleType map[string]matcherType

// Matcher is responsible for generating Pact values and matching in the Pact file.
// It can be used as an alternative to plain string matches in the DSL.
//
// The Matcher part will be the right hand side of a matchingRule declaration
// The Value part will be used in the example body/header/request/etc.
type Matcher struct {
	// Matcher gets the matching strategy associated with the current Matcher.
	Matcher matcherType

	// Value to be serialised to JSON. This value is what is used in the example
	// for API responses.
	Value interface{}

	// Type of Matcher
	Type int
}

// Matcher Types
const (
	// LikeMatcher is the ID for the Like Matcher
	LikeMatcher = iota

	// TermMatcher is the ID for the Term Matcher
	TermMatcher

	// ArrayMinLikeMatcher is the ID for the ArrayMinLike Matcher
	ArrayMinLikeMatcher

	// ArrayMaxLikeMatcher is the ID for the ArrayMaxLikeMatcher Matcher
	ArrayMaxLikeMatcher
)

// jsonArray is the type for JSON arrays
type jsonArray map[string][]interface{}

// Creates sample array contents given an array Matcher.
func makeArrayContents(times int, key string, value interface{}) jsonArray {
	contents := make([]interface{}, times)
	for i := 0; i < times; i++ {
		contents[i] = value
	}
	return jsonArray{
		key: contents,
	}
}

// ArrayMinLike matches nested arrays in request bodies.
// Ensure that each item in the list matches the provided example and the list
// is no smaller than the provided min.
func ArrayMinLike(min int, value interface{}) Matcher {
	return Matcher{
		Matcher: map[string]interface{}{
			"min":   min,
			"match": "type",
		},
		Value: value, //makeArrayContents(min, key, value),
		Type:  ArrayMinLikeMatcher,
	}
}

// ArrayMaxLike matches nested arrays in request bodies.
// Ensure that each item in the list matches the provided example and the list
// is no greater than the provided max.
func ArrayMaxLike(max int, value interface{}) Matcher {
	return Matcher{
		Matcher: map[string]interface{}{
			"max":   max,
			"match": "type",
		},
		Value: value, //makeArrayContents(min, key, value),
		Type:  ArrayMinLikeMatcher,
	}
}

// Like specifies that the given content type should be matched based
// on type (int, string etc.) instead of a verbatim match.
func Like(content interface{}) Matcher {
	return Matcher{
		Matcher: map[string]interface{}{
			"match": "type",
		},
		Value: content,
		Type:  LikeMatcher,
	}
}

// PactTerm specifies that the given content type should be matched based
// on a regular expression.
func PactTerm(matcher string, content interface{}) Matcher {
	return Matcher{
		Matcher: map[string]interface{}{
			"match": "regex",
			"regex": matcher,
		},
		Value: content,
		Type:  TermMatcher,
	}
}

// PactDslBuilder contains the struct generates examples and matching rules
// given a structure containing matchers.
type PactDslBuilder struct {
	// Matching rules used by the verifier to confirm Provider confirms to Pact.
	MatchingRules matchingRuleType `json:"matchingRules"`

	// Generated test body for the consumer testing via the Mock Server.
	Body map[string]interface{} `json:"body"`

	path string
}

// BuildPact takes a map containing recursive Matchers and generates the rules
// to be serialised into the Pact file.
func BuildPact(root map[string]interface{}) PactDslBuilder {

	dsl := PactDslBuilder{}
	dsl.path = "$.body"
	// Recurse through the matcher, updating path as we go

	// 1.1 Recurse through Matcher, building generated body first
	// 1.2 Update PATH as we go -> deferred

	dsl.path, dsl.Body, dsl.MatchingRules = build("", root, make(map[string]interface{}), dsl.path, make(matchingRuleType))

	return dsl
}

const pathSep = "."
const allListItems = "[*]"
const startList = "["
const endList = "]"

// Recurse the Matcher tree and build up an example body and set of matchers for
// the Pact file. Ideally this stays as a pure function, but probably might need
// to store matchers externally.
//
// See PactBodyBuilder.groovy line 96 for inspiration/logic.
//
// Arguments:
// 	- key           => Current key in the body to set
// 	- value         => Value held in the next Matcher (which may be another Matcher)
// 	- body          => Current state of the body map
// 	- path          => Path to the current key
//  - matchingRules => Current set of matching rules
//
// TODO: Should return a DSL/Some object that encapsulates the path, matchers and body???
//
func build(key string, value interface{}, body map[string]interface{}, path string, matchingRules matchingRuleType) (string, map[string]interface{}, matchingRuleType) {
	log.Println("[DEBUG] dsl generator: recursing => key:", key, ", body:", body, ", value: ", value)

	switch t := value.(type) {

	case Matcher:
		switch t.Type {

		// ArrayLike Matchers
		case ArrayMinLikeMatcher, ArrayMaxLikeMatcher:
			times := 1

			if _, ok := t.Matcher["min"]; ok {
				times = t.Matcher["min"].(int)
			} else {
				times = t.Matcher["max"].(int)
			}

			arrayMap := make(map[string]interface{})
			minArray := make([]interface{}, times)

			build("0", t.Value, arrayMap, path+buildPath(key, allListItems), matchingRules)
			log.Println("[DEBUG] dsl generator: adding matcher (arrayLike) =>", path+buildPath(key, ""))
			matchingRules[path+buildPath(key, "")] = t.Matcher

			// TODO: Need to understand the .* notation before implementing it. Notably missing from Groovy DSL
			// log.Println("[DEBUG] dsl generator: Adding matcher (type)              =>", path+buildPath(key, allListItems)+".*")
			// matchingRules[path+buildPath(key, allListItems)+".*"] = t.Matcher

			for i := 0; i < times; i++ {
				minArray[i] = arrayMap["0"]
			}
			body[key] = minArray
			path = path + buildPath(key, "")

			// Simple Matchers (Terminal cases)
		case TermMatcher, LikeMatcher:
			body[key] = t.Value
			log.Println("[DEBUG] dsl generator: adding matcher (Term/Like)         =>", path+buildPath(key, ""))
			matchingRules[path+buildPath(key, "")] = t.Matcher
		default:
			log.Fatalf("unknown matcher: %d", t.Type)
		}

	// Slice/Array types
	case []interface{}:
		arrayValues := make([]interface{}, len(t))
		arrayMap := make(map[string]interface{})

		// This is a real hack. I don't like it
		// I also had to do it for the Array*LikeMatcher's, which I also don't like
		for i, el := range t {
			k := fmt.Sprintf("%d", i)
			build(k, el, arrayMap, path+buildPath(key, fmt.Sprintf("%s%d%s", startList, i, endList)), matchingRules)
			arrayValues[i] = arrayMap[k]
		}
		body[key] = arrayValues

	// Map -> Recurse keys (All objects start here!)
	case map[string]interface{}:
		entry := make(map[string]interface{})
		path = path + buildPath(key, "")

		for k, v := range t {
			log.Println("[DEBUG] dsl generator: \t=> map type. recursing into key =>", k)

			// Starting position
			if key == "" {
				_, body, matchingRules = build(k, v, copyMap(body), path, matchingRules)
			} else {
				_, body[key], matchingRules = build(k, v, entry, path, matchingRules)
			}
		}

	// Primitives (terminal cases)
	default:
		log.Println("[DEBUG] dsl generator: \t=> unknown type, probably just a primitive (string/int/etc.)", value)
		body[key] = value
	}

	log.Println("[DEBUG] dsl generator: returning body: ", body)

	return path, body, matchingRules
}

// TODO: allow regex in paths.
func buildPath(name string, children string) string {
	// We know if a key is an integer, it's not valid JSON and therefore is Probably
	// the shitty array hack from above. Skip creating a new path if the key is bungled
	// TODO: save the children?
	if _, err := strconv.Atoi(name); err != nil && name != "" {
		return pathSep + name + children
	}

	return ""
}
