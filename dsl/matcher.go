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
	"math/rand"
	"strconv"
	"time"

	uuidv4 "github.com/twinj/uuid"
)

// matcherType is essentially a key value JSON pairs for serialisation
type matcherType map[string]interface{}

// Matching Rule
type matchingRuleType map[string]matcherType

// Matcher regexes
const (
	hexadecimal = `[0-9a-fA-F]+`
	ipAddress   = `(\d{1,3}\.)+\d{1,3}`
	ipv6Address = `(\A([0-9a-f]{1,4}:){1,1}(:[0-9a-f]{1,4}){1,6}\Z)|(\A([0-9a-f]{1,4}:){1,2}(:[0-9a-f]{1,4}){1,5}\Z)|(\A([0-9a-f]{1,4}:){1,3}(:[0-9a-f]{1,4}){1,4}\Z)|(\A([0-9a-f]{1,4}:){1,4}(:[0-9a-f]{1,4}){1,3}\Z)|(\A([0-9a-f]{1,4}:){1,5}(:[0-9a-f]{1,4}){1,2}\Z)|(\A([0-9a-f]{1,4}:){1,6}(:[0-9a-f]{1,4}){1,1}\Z)|(\A(([0-9a-f]{1,4}:){1,7}|:):\Z)|(\A:(:[0-9a-f]{1,4}){1,7}\Z)|(\A((([0-9a-f]{1,4}:){6})(25[0-5]|2[0-4]\d|[0-1]?\d?\d)(\.(25[0-5]|2[0-4]\d|[0-1]?\d?\d)){3})\Z)|(\A(([0-9a-f]{1,4}:){5}[0-9a-f]{1,4}:(25[0-5]|2[0-4]\d|[0-1]?\d?\d)(\.(25[0-5]|2[0-4]\d|[0-1]?\d?\d)){3})\Z)|(\A([0-9a-f]{1,4}:){5}:[0-9a-f]{1,4}:(25[0-5]|2[0-4]\d|[0-1]?\d?\d)(\.(25[0-5]|2[0-4]\d|[0-1]?\d?\d)){3}\Z)|(\A([0-9a-f]{1,4}:){1,1}(:[0-9a-f]{1,4}){1,4}:(25[0-5]|2[0-4]\d|[0-1]?\d?\d)(\.(25[0-5]|2[0-4]\d|[0-1]?\d?\d)){3}\Z)|(\A([0-9a-f]{1,4}:){1,2}(:[0-9a-f]{1,4}){1,3}:(25[0-5]|2[0-4]\d|[0-1]?\d?\d)(\.(25[0-5]|2[0-4]\d|[0-1]?\d?\d)){3}\Z)|(\A([0-9a-f]{1,4}:){1,3}(:[0-9a-f]{1,4}){1,2}:(25[0-5]|2[0-4]\d|[0-1]?\d?\d)(\.(25[0-5]|2[0-4]\d|[0-1]?\d?\d)){3}\Z)|(\A([0-9a-f]{1,4}:){1,4}(:[0-9a-f]{1,4}){1,1}:(25[0-5]|2[0-4]\d|[0-1]?\d?\d)(\.(25[0-5]|2[0-4]\d|[0-1]?\d?\d)){3}\Z)|(\A(([0-9a-f]{1,4}:){1,5}|:):(25[0-5]|2[0-4]\d|[0-1]?\d?\d)(\.(25[0-5]|2[0-4]\d|[0-1]?\d?\d)){3}\Z)|(\A:(:[0-9a-f]{1,4}){1,5}:(25[0-5]|2[0-4]\d|[0-1]?\d?\d)(\.(25[0-5]|2[0-4]\d|[0-1]?\d?\d)){3}\Z)`
	uuid        = `[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}`
	Timestamp   = `^([\+-]?\d{4}(?!\d{2}\b))((-?)((0[1-9]|1[0-2])(\3([12]\d|0[1-9]|3[01]))?|W([0-4]\d|5[0-2])(-?[1-7])?|(00[1-9]|0[1-9]\d|[12]\d{2}|3([0-5]\d|6[1-6])))([T\s]((([01]\d|2[0-3])((:?)[0-5]\d)?|24\:?00)([\.,]\d+(?!:))?)?(\17[0-5]\d([\.,]\d+)?)?([zZ]|([\+-])([01]\d|2[0-3]):?([0-5]\d)?)?)?)?$`
	Date        = `^([\+-]?\d{4}(?!\d{2}\b))((-?)((0[1-9]|1[0-2])(\3([12]\d|0[1-9]|3[01]))?|W([0-4]\d|5[0-2])(-?[1-7])?|(00[1-9]|0[1-9]\d|[12]\d{2}|3([0-5]\d|6[1-6])))?)`
	Time        = `^(T\d\d:\d\d(:\d\d)?(\.\d+)?(([+-]\d\d:\d\d)|Z)?)?$`
)

// Matcher is responsible for generating Pact values and matching in the Pact file.
// It can be used as an alternative to plain string matches in the DSL.
//
// The Matcher part will be the right hand side of a matchingRule declaration
// The Value part will be used in the example body/header/request/etc.
type Matcher struct {
	// Matcher contains the matching strategy associated with the current Matcher.
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

	// RegexMatcher is the ID for the Term Matcher
	RegexMatcher

	// ArrayMinLikeMatcher is the ID for the ArrayMinLike Matcher
	ArrayMinLikeMatcher

	// ArrayMaxLikeMatcher is the ID for the ArrayMaxLikeMatcher Matcher
	ArrayMaxLikeMatcher
)

// jsonArray is the type for JSON arrays
type jsonArray map[string][]interface{}

// ArrayMinLike matches nested arrays in request bodies.
// Ensure that each item in the list matches the provided example and the list
// is no smaller than the provided min.
func ArrayMinLike(min int, value interface{}) Matcher {
	return Matcher{
		Matcher: map[string]interface{}{
			"min":   min,
			"match": "type",
		},
		Value: value,
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

// Regex specifies that the given content type should be matched based
// on a regular expression.
func Regex(matcher string, content interface{}) Matcher {
	return Matcher{
		Matcher: map[string]interface{}{
			"match": "regex",
			"regex": matcher,
		},
		Value: content,
		Type:  RegexMatcher,
	}
}

// HexValue defines a matcher that accepts hexidecimal values.
func HexValue() Matcher {
	return Matcher{
		Matcher: map[string]interface{}{
			"match": "regex",
			"regex": hexadecimal,
		},
		Value: "3F",
		Type:  RegexMatcher,
	}
}

// Identifier defines a matcher that accepts integer values.
func Identifier() Matcher {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	return Matcher{
		Matcher: map[string]interface{}{
			"match": "type",
		},
		Value: r.Int63(),
		Type:  LikeMatcher,
	}
}

// Integer defines a matcher that accepts ints. Identical to Identifier.
var Integer = Identifier

// IPAddress defines a matcher that accepts valid IPv4 addresses.
func IPAddress() Matcher {
	return Matcher{
		Matcher: map[string]interface{}{
			"match": "regex",
			"regex": ipAddress,
		},
		Value: "127.0.0.1",
		Type:  RegexMatcher,
	}
}

// IPv4Address matches valid IPv4 addresses.
var IPv4Address = IPAddress

// IPv6Address defines a matcher that accepts IP addresses.
func IPv6Address() Matcher {
	return Matcher{
		Matcher: map[string]interface{}{
			"match": "regex",
			"regex": ipAddress,
		},
		Value: "::ffff:192.0.2.128",
		Type:  RegexMatcher,
	}
}

// Decimal defines a matcher that accepts any decimal value.
func Decimal() Matcher {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	return Matcher{
		Matcher: map[string]interface{}{
			"match": "type",
		},
		Value: r.Float64(),
		Type:  LikeMatcher,
	}
}

// Timestamp matches a pattern corresponding to the ISO_DATETIME_FORMAT, which
// is "yyyy-MM-dd'T'HH:mm:ss". The current date and time is used as the eaxmple.
func Timestamp() Matcher {
	return Matcher{
		Matcher: map[string]interface{}{
			"match": "regex",
			"regex": Timestamp,
		},
		Value: time.Now().Format(time.RFC3339),
		Type:  RegexMatcher,
	}
}

// Date matches a pattern corresponding to the ISO_DATE_FORMAT, which
// is "yyyy-MM-dd". The current date is used as the eaxmple.
func Date() Matcher {
	return Matcher{
		Matcher: map[string]interface{}{
			"match": "regex",
			"regex": Date,
		},
		Value: time.Now().Format("2006-01-02"),
		Type:  RegexMatcher,
	}
}

// Time matches a pattern corresponding to the ISO_DATE_FORMAT, which
// is "'T'HH:mm:ss". The current tem is used as the eaxmple.
func Time() Matcher {
	return Matcher{
		Matcher: map[string]interface{}{
			"match": "regex",
			"regex": Time,
		},
		Value: time.Now().Format("T15:04:05"),
		Type:  RegexMatcher,
	}
}

// UUID defines a matcher that accepts UUIDs. Produces a v4 UUID as the example.
func UUID() Matcher {
	return Matcher{
		Matcher: map[string]interface{}{
			"match": "regex",
			"regex": uuid,
		},
		Value: uuidv4.NewV4().String(),
		Type:  RegexMatcher,
	}
}

// PactBody contains the struct generates examples and matching rules
// given a structure containing matchers.
type PactBody struct {
	// Matching rules used by the verifier to confirm Provider confirms to Pact.
	MatchingRules matchingRuleType `json:"matchingRules"`

	// Generated test body for the consumer testing via the Mock Server.
	Body map[string]interface{} `json:"body"`
}

// PactBodyBuilder takes a map containing recursive Matchers and generates the rules
// to be serialised into the Pact file.
func PactBodyBuilder(root map[string]interface{}) PactBody {
	dsl := PactBody{}
	_, dsl.Body, dsl.MatchingRules = build("", root, make(map[string]interface{}),
		"$.body", make(matchingRuleType))

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
// See PactBody.groovy line 96 for inspiration/logic.
//
// Arguments:
// 	- key           => Current key in the body to set
// 	- value         => Value held in the next Matcher (which may be another Matcher)
// 	- body          => Current state of the body map
// 	- path          => Path to the current key
//  - matchingRules => Current set of matching rules
func build(key string, value interface{}, body map[string]interface{}, path string,
	matchingRules matchingRuleType) (string, map[string]interface{}, matchingRuleType) {
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
		case RegexMatcher, LikeMatcher:
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

// TODO: allow regex in request paths.
func buildPath(name string, children string) string {
	// We know if a key is an integer, it's not valid JSON and therefore is Probably
	// the shitty array hack from above. Skip creating a new path if the key is bungled
	// TODO: save the children?
	if _, err := strconv.Atoi(name); err != nil && name != "" {
		return pathSep + name + children
	}

	return ""
}
