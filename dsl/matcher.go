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

// Matcher defines how a given part of an API (request/response/header/etc.)
// is checked during Pact verification
// type Matcher struct {
// 	// The type of Matcher
// 	Type string
//
// 	// Matcher to be serialised. Keys must be strings to be valid JSON.
// 	Value map[string]interface{}
// }

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

const (
	LikeMatcher = iota
	TermMatcher
	ArrayMinLikeMatcher
	ArrayMaxLikeMatcher
)

type keyedArrayElement map[string][]interface{}

// type keyedArrayElement map[string][]map[string]interface{}

// Creates sample array contents given an array Matcher.
func makeArrayContents(times int, key string, value interface{}) keyedArrayElement {
	contents := make([]interface{}, times)
	for i := 0; i < times; i++ {
		contents[i] = value
	}
	return keyedArrayElement{
		key: contents,
	}
}

// ArrayMinLike matches nested arrays in request bodies.
// Ensure that each item in the list matches the provided example and the list
// is no smaller than the provided min
func ArrayMinLike(min int, key string, value interface{}) Matcher {
	return Matcher{
		Matcher: map[string]interface{}{
			"min":   min,
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

// Like specifies that the given content type should be matched based
// on type (int, string etc.) instead of a verbatim match.
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

type PactDslBuilder struct {
	// Matching rules used by the verifier to confirm Provider confirms to Pact.
	MatchingRules map[string]string `json:"matchingRules"`

	// Generated test body for the consumer testing via the Mock Server.
	Body map[string]interface{} `json:"body"`

	path string
}

func BuildPact(root Matcher) PactDslBuilder {

	dsl := PactDslBuilder{}
	dsl.path = "$.body"
	body := make(map[string]interface{})
	// Recurse through the matcher, updating path as we go

	// 1.1 Recurse through Matcher, building generated body first
	// 1.2 Update PATH as we go -> deferred

	dsl.path, dsl.Body = recurseStructure(root, body, dsl.path)

	return dsl
}

// Update path to the specific path
// See PactBodyBuilder.groovy line 96 for inspiration/logic
func recurseStructure(value interface{}, body map[string]interface{}, path string) (string, map[string]interface{}) {
	fmt.Println("Recursing => value:", "", ", body:", body, "path:", path)

	switch t := value.(type) {

	case Matcher:
		fmt.Println("=> Matcher")
		switch t.Type {

		// Like Matchers
		case ArrayMaxLikeMatcher:
			fmt.Println("\t=> ArrayMaxLikeMatcher")
			// body[key] = recurseStructure(t.Value, path + buildPath(key, ALL_LIST_ITEMS))
			recurseStructure(t.Value, body, buildPath(path, ""))
		case ArrayMinLikeMatcher:
			fmt.Println("\t=> ArrayMinLikeMatcher")
			recurseStructure(t.Value, body, buildPath(path, ""))

		// Simple Matchers
		case TermMatcher:
			fmt.Println("\t=> TermMatcher", t)
			recurseStructure(t.Value, body, buildPath(path, ""))
		default:
			// should probably throw an error here!?
			log.Fatalf("Unknown matcher detected: %d", t.Type)
		}

	// Slice/Array types
	// case []interface{}
	// body[key]  = recurseStructure(value, body, path + buildPath(key, START_LIST + i + END_LIST)) <- matchers
	// body[key] = recurseStructure(value, body, path + buildPath(key)) <- terminating case (primitives)

	// Map -> Recurse keys
	case map[string]interface{}:
		// iterate over Keys
		for k, v := range t {
			fmt.Println("\t=> Map type. recursing...", k, v)
			path, body[k] = recurseStructure(v, body, path)
		}

	// Primitives
	default:
		fmt.Println("\t=> Unknown type 2. Probably just a string/int/etc.", value)
	}

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

// Like specifies that the given content type should be matched based
// on type (int, string etc.) instead of a verbatim match.
// func Like(content interface{}) string {
// 	return fmt.Sprintf(`
// 		{
// 		  "json_class": "Pact::SomethingLike",
// 		  "contents": %v
// 		}`, content)
// }

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

// Regex specifies that the matching should generate a value
// and also match using a regular expression.
// Synonym of Term.
func Regex(generate string, matcher string) string {
	return Term(generate, matcher)
}

// MinType executes a type based match against the values, that is, they are
// equal if they are the same type. In addition, if the values represent a
// collection, the length of the actual value is compared against the minimum.
func MinType(content interface{}, min int) string {
	return ""
}

// MaxType executes a type based match against the values, that is, they are
// equal if they are the same type. In addition, if the values represent a
// collection, the length of the actual value is compared against the maximum.
func MaxType(content interface{}, max int) string {
	return ""
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
