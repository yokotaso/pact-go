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
import (
	"fmt"
	"log"
)

// matcherType is essentially a key value JSON pairs for serialisation
type matcherType map[string]interface{}

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

// ArrayMaxlike matches nested arrays in request bodies.
// Ensure that each item in the list matches the provided example and the list
// is no greater than the provided max.
func ArrayMaxlike(max int, value interface{}) Matcher {
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
	MatchingRules map[string]string `json:"matchingRules"`

	// Generated test body for the consumer testing via the Mock Server.
	Body map[string]interface{} `json:"body"`

	path string
}

// BuildPact takes a map containing recursive Matchers and generates the rules
// to be serialised into the Pact file.
func BuildPact(root map[string]interface{}) PactDslBuilder {

	dsl := PactDslBuilder{}
	dsl.path = "$.body"
	body := make(map[string]interface{})
	// Recurse through the matcher, updating path as we go

	// 1.1 Recurse through Matcher, building generated body first
	// 1.2 Update PATH as we go -> deferred

	dsl.path, dsl.Body = build("", root, body, dsl.path)

	return dsl
}

const pathSep = "."
const allListItems = "[*]"
const startList = "["
const endList = "]"

// Store all of the matchers in here
var matchers map[string]matcherType

// Recurse the Matcher tree and build up an example body and set of matchers for
// the Pact file. Ideally this stays as a pure function, but probably might need
// to store matchers externally.
//
// Update path to the specific path -> Path
// See PactBodyBuilder.groovy line 96 for inspiration/logic.
//
// Arguments:
// 	- key => Current key in the body to set
// 	- value => Value held in the next Matcher (which may be another Matcher)
// 	- body => Current state of the body map
// 	- path => Path to the current key TODO: Path not doing anything yet.
func build(key string, value interface{}, body map[string]interface{}, path string) (string, map[string]interface{}) {
	fmt.Println("Recursing => key:", key, ", body:", body, ", value: ", value)

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

			build("0", t.Value, arrayMap, path+buildPath(key, fmt.Sprintf("%s%d%s", startList, 0, endList)))
			for i := 0; i < times; i++ {
				minArray[i] = arrayMap["0"]
			}
			body[key] = minArray

			// Simple Matchers (Terminal cases)
		case TermMatcher:
			body[key] = t.Value
		case LikeMatcher:
			body[key] = t.Value
		default:
			log.Fatalf("unknown matcher: %d", t.Type)
		}

	// Slice/Array types
	case []interface{}:
		arrayValues := make([]interface{}, len(t))
		arrayMap := make(map[string]interface{})

		for i, el := range t {
			k := fmt.Sprintf("%d", i)
			build(k, el, arrayMap, path+buildPath(key, fmt.Sprintf("%s%d%s", startList, i, endList)))
			arrayValues[i] = arrayMap[k]
		}
		body[key] = arrayValues

	// Map -> Recurse keys (All objects start here!)
	case map[string]interface{}:
		entry := make(map[string]interface{})

		for k, v := range t {
			fmt.Println("\t=> Map type. recursing into key =>", k)

			// Starting position
			if key == "" {
				_, body = build(k, v, copyMap(body), path)
			} else {
				_, body[key] = build(k, v, entry, path)
			}
		}

	// Primitives (terminal cases)
	default:
		fmt.Println("\t=> Unknown type. Probably just a primitive (string/int/etc.)", value)
		body[key] = value
	}

	fmt.Println("Returning body: ", body)

	return path, body
}

// TODO: allow regex in paths.
func buildPath(name string, children string) string {
	return name + children
}

// EachLike specifies that a given element in a JSON body can be repeated
// "minRequired" times. Number needs to be 1 or greater
func EachLike(content interface{}, minRequired int) string {
	return fmt.Sprintf(`
		{
		  "json_class": "Pact::ArrayLike",
		  "contents": %v,
		  "min": %d
		}`, content, minRequired)
}

// Term specifies that the matching should generate a value
// and also match using a regular expression.
// Synonym of Regex.
func Term(generate string, matcher string) string {
	return fmt.Sprintf(`
		{
			"json_class": "Pact::Term",
			"data": {
			  "generate": "%s",
			  "matcher": {
			    "json_class": "Regexp",
			    "o": 0,
			    "s": "%s"
			  }
			}
		}`, generate, matcher)
}

// Matching generation rules
// 1 - matchingRules takes a map of "JSONPath -> Matching Rule Object" (e.g. "$.body.animals": {""})
//    - 5 types of matchers (https://github.com/pact-foundation/pact-reference/tree/master/rust/libpact_matching#supported-matchers)
//       - t{"match":"regex", "regex": "red|mblue"}
//       - t{"match":"type"}
//    - Need to keep track of the depth in the selectors, which are fairly simplistic
// 2 - Need to generate the body as the Pact Interactions are done

// type Matcher interface {
// 	Path string // <- path to the current level? e.g. $.body.foo[*]
//
// }
//
// Builder {
// 	matchingRules []
// 	body interface{}
//
//
// }
//
// i := pact.
// 	AddInteraction()
//
// 	// Setup a complex interaction
// 	jumper := Like(`"jumper"`)
// 	shirt := Like(`"shirt"`)
// 	tag := EachLike(fmt.Sprintf(`[%s, %s]`, jumper, shirt), 2)
// 	size := Like(10)
// 	colour := Term("red", "red|green|blue")
//
// body :=
// 	// NewJsonDslBody(&i).				// <- needs the Interaction struct to generate matchers as it goes. Returns a string
// 	NewJsonDslBody().
// 		ArrayLikeMin(3, "foobararray",
// 			map[string]string{
// 				"foo": StringType("foovalue"),
// 				"bar": IntegerType(27)})
// 		EachLikeMin(1, "someotherobjectarray",
// 			fmt.Sprintf(
// 				`{
// 					"size": 10,
// 					"colour": "red",
// 					"tag": "red"
// 				}`))

// -> Should produce this generated body

// {
// 	"foobararray": [
// 		{
// 			"foo": "foovalue",
// 			"bar": 27,
// 		},
// 		{
// 			"foo": "foovalue",
// 			"bar": 27,
// 		},
// 		{
// 			"foo": "foovalue",
// 			"bar": 27,
// 		}
// 	],
// 	"someotherobjectarray": [
// 		{
// 			"size": 10,
// 			"colour": "red",
// 			"tag": "red"
// 		}
// 	]
// }

// -> Should produce the following matching rules:
// "matchingRules": {
// 	"$.foobararray": {"min": 3, "match": "type"}
// 	"$.foobararray[*].*":
// }

// Types supported by JVM
// method	description
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
