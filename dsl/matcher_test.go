package dsl

import (
	"fmt"
	"reflect"
	"regexp"
	"testing"
)

func TestMatcher_SugarMatchers(t *testing.T) {
	type matcherTestCase struct {
		matcher  Matcher
		testCase func(val interface{}) error
	}
	matchers := map[string]matcherTestCase{
		"HexValue": matcherTestCase{
			matcher: HexValue(),
			testCase: func(v interface{}) (err error) {
				if v.(string) != "3F" {
					err = fmt.Errorf("want '3F', got '%v'", v)
				}
				return
			},
		},
		"Identifier": matcherTestCase{
			matcher: Identifier(),
			testCase: func(v interface{}) (err error) {
				_, valid := v.(int64)
				if !valid {
					err = fmt.Errorf("want int64, got '%v'", v)
				}
				return
			},
		},
		"Integer": matcherTestCase{
			matcher: Integer(),
			testCase: func(v interface{}) (err error) {
				_, valid := v.(int64)
				if !valid {
					err = fmt.Errorf("want int64, got '%v'", v)
				}
				return
			},
		},
		"IPAddress": matcherTestCase{
			matcher: IPAddress(),
			testCase: func(v interface{}) (err error) {
				if v.(string) != "127.0.0.1" {
					err = fmt.Errorf("want '127.0.0.1', got '%v'", v)
				}
				return
			},
		},
		"IPv4Address": matcherTestCase{
			matcher: IPv4Address(),
			testCase: func(v interface{}) (err error) {
				if v.(string) != "127.0.0.1" {
					err = fmt.Errorf("want '127.0.0.1', got '%v'", v)
				}
				return
			},
		},
		"IPv6Address": matcherTestCase{
			matcher: IPv6Address(),
			testCase: func(v interface{}) (err error) {
				if v.(string) != "::ffff:192.0.2.128" {
					err = fmt.Errorf("want '::ffff:192.0.2.128', got '%v'", v)
				}
				return
			},
		},
		"Decimal": matcherTestCase{
			matcher: Decimal(),
			testCase: func(v interface{}) (err error) {
				_, valid := v.(float64)
				if !valid {
					err = fmt.Errorf("want float64, got '%v'", v)
				}
				return
			},
		},
		"Timestamp": matcherTestCase{
			matcher: Timestamp(),
			testCase: func(v interface{}) (err error) {
				_, valid := v.(string)
				if !valid {
					err = fmt.Errorf("want string, got '%v'", v)
				}
				return
			},
		},
		"Date": matcherTestCase{
			matcher: Date(),
			testCase: func(v interface{}) (err error) {
				_, valid := v.(string)
				if !valid {
					err = fmt.Errorf("want string, got '%v'", v)
				}
				return
			},
		},
		"Time": matcherTestCase{
			matcher: Time(),
			testCase: func(v interface{}) (err error) {
				_, valid := v.(string)
				if !valid {
					err = fmt.Errorf("want string, got '%v'", v)
				}
				return
			},
		},
		"UUID": matcherTestCase{
			matcher: UUID(),
			testCase: func(v interface{}) (err error) {
				match, err := regexp.MatchString(uuid, v.(string))

				if !match {
					err = fmt.Errorf("want string, got '%v'. Err: %v", v, err)
				}
				return
			},
		},
	}
	var err error
	for k, v := range matchers {
		if err = v.testCase(v.matcher.Value); err != nil {
			t.Fatalf("error validating matcher '%s': %v", k, err)
		}
	}
}

func TestMatcher_ArrayMinLike(t *testing.T) {
	matcher := map[string]interface{}{
		"users": ArrayMinLike(3, map[string]interface{}{
			"user": Regex("\\s+", "someusername")})}

	expectedBody := formatJSON(`{
		"users": [
			{
				"user": "someusername"
			},
			{
				"user": "someusername"
			},
			{
				"user": "someusername"
			}
		]
	}`)
	expectedMatchingRules := matchingRuleType{
		"$.body.users": map[string]interface{}{
			"match": "type",
			"min":   3,
		},
		"$.body.users[*].user": map[string]interface{}{
			"match": "regex",
			"regex": "\\s+",
		},
	}

	body := PactBodyBuilder(matcher)
	result := formatJSONObject(body.Body)

	if expectedBody != result {
		t.Fatalf("got '%v' wanted '%v'", result, expectedBody)
	}
	if !reflect.DeepEqual(body.MatchingRules, expectedMatchingRules) {
		t.Fatalf("got '%v' wanted '%v'", body.MatchingRules, expectedMatchingRules)
	}
}

func TestMatcher_ArrayMaxLike(t *testing.T) {
	matcher := map[string]interface{}{
		"users": ArrayMaxLike(3, map[string]interface{}{
			"user": Regex("\\s+", "someusername")})}

	expectedBody := formatJSON(`{
		"users": [
			{
				"user": "someusername"
			},
			{
				"user": "someusername"
			},
			{
				"user": "someusername"
			}
		]
	}`)
	expectedMatchingRules := matchingRuleType{
		"$.body.users": map[string]interface{}{
			"match": "type",
			"max":   3,
		},
		"$.body.users[*].user": map[string]interface{}{
			"match": "regex",
			"regex": "\\s+",
		},
	}

	body := PactBodyBuilder(matcher)
	result := formatJSONObject(body.Body)

	if expectedBody != result {
		t.Fatalf("got '%v' wanted '%v'", result, expectedBody)
	}
	if !reflect.DeepEqual(body.MatchingRules, expectedMatchingRules) {
		t.Fatalf("got '%v' wanted '%v'", body.MatchingRules, expectedMatchingRules)
	}
}

func TestMatcher_NestedMaps(t *testing.T) {
	matcher := map[string]interface{}{
		"user": map[string]interface{}{
			"phone":     Regex("\\d+", 12345678),
			"name":      Regex("\\s+", "someusername"),
			"address":   Regex("\\s+", "some address"),
			"plaintext": "plaintext",
		},
		"pass": Regex("\\d+", 1234),
	}

	expectedBody := formatJSON(`{
		"pass": 1234,
		"user": {
			"address": "some address",
			"name": "someusername",
			"phone": 12345678,
			"plaintext": "plaintext"
		}
	}`)
	expectedMatchingRules := matchingRuleType{
		"$.body.pass": map[string]interface{}{
			"match": "regex",
			"regex": "\\d+",
		},
		"$.body.user.phone": map[string]interface{}{
			"match": "regex",
			"regex": "\\d+",
		},
		"$.body.user.name": map[string]interface{}{
			"match": "regex",
			"regex": "\\s+",
		},
		"$.body.user.address": map[string]interface{}{
			"match": "regex",
			"regex": "\\s+",
		},
	}

	body := PactBodyBuilder(matcher)
	result := formatJSONObject(body.Body)

	if expectedBody != result {
		t.Fatalf("got '%v' wanted '%v'", result, expectedBody)
	}
	if !reflect.DeepEqual(body.MatchingRules, expectedMatchingRules) {
		t.Fatalf("got '%v' wanted '%v'", body.MatchingRules, expectedMatchingRules)
	}
}

func TestMatcher_Arrays(t *testing.T) {
	matcher := map[string]interface{}{
		"users": []interface{}{
			Regex("\\s+", "someusername1"),
			Regex("\\s+", "someusername2"),
			Regex("\\s+", "someusername3"),
		},
		"pass": Regex("\\d+", 1234),
		"id":   5678,
	}
	expectedBody := formatJSON(`{
		"id": 5678,
		"pass": 1234,
		"users": [
			"someusername1",
			"someusername2",
			"someusername3"
		]
	}`)
	expectedMatchingRules := matchingRuleType{
		"$.body.pass": map[string]interface{}{
			"match": "regex",
			"regex": "\\d+",
		},
		"$.body.users[0]": map[string]interface{}{
			"match": "regex",
			"regex": "\\s+",
		},
		"$.body.users[1]": map[string]interface{}{
			"match": "regex",
			"regex": "\\s+",
		},
		"$.body.users[2]": map[string]interface{}{
			"match": "regex",
			"regex": "\\s+",
		},
	}

	body := PactBodyBuilder(matcher)
	result := formatJSONObject(body.Body)

	if expectedBody != result {
		t.Fatalf("got '%v' wanted '%v'", result, expectedBody)
	}
	if !reflect.DeepEqual(body.MatchingRules, expectedMatchingRules) {
		t.Fatalf("got '%v' wanted '%v'", body.MatchingRules, expectedMatchingRules)
	}
}

func TestMatcher_Like(t *testing.T) {
	matcher := map[string]interface{}{
		"users": []interface{}{
			Like("someusername1"),
			Like("someusername2"),
			Like("someusername3"),
		},
		"pass": Like(1234),
	}
	expectedBody := formatJSON(`{
		"pass": 1234,
		"users": [
			"someusername1",
			"someusername2",
			"someusername3"
		]
	}`)
	expectedMatchingRules := matchingRuleType{
		"$.body.pass": map[string]interface{}{
			"match": "type",
		},
		"$.body.users[0]": map[string]interface{}{
			"match": "type",
		},
		"$.body.users[1]": map[string]interface{}{
			"match": "type",
		},
		"$.body.users[2]": map[string]interface{}{
			"match": "type",
		},
	}

	body := PactBodyBuilder(matcher)
	result := formatJSONObject(body.Body)

	if expectedBody != result {
		t.Fatalf("got '%v' wanted '%v'", result, expectedBody)
	}
	if !reflect.DeepEqual(body.MatchingRules, expectedMatchingRules) {
		t.Fatalf("got '%v' wanted '%v'", body.MatchingRules, expectedMatchingRules)
	}
}

func TestMatcher_Regex(t *testing.T) {
	matcher := map[string]interface{}{
		"user": Regex("\\s+", "someusername3"),
	}
	expectedBody := formatJSON(`{
		"user":	"someusername3"
	}`)
	expectedMatchingRules := matchingRuleType{
		"$.body.user": map[string]interface{}{
			"match": "regex",
			"regex": "\\s+",
		},
	}

	body := PactBodyBuilder(matcher)
	result := formatJSONObject(body.Body)

	if expectedBody != result {
		t.Fatalf("got '%v' wanted '%v'", result, expectedBody)
	}
	if !reflect.DeepEqual(body.MatchingRules, expectedMatchingRules) {
		t.Fatalf("got '%v' wanted '%v'", body.MatchingRules, expectedMatchingRules)
	}
}

func TestMatcher_DeepArrayMinLike(t *testing.T) {
	matcher1 := map[string]interface{}{
		"users": ArrayMinLike(3, map[string]interface{}{
			"user": Regex("\\s+", "someusername")})}
	matcher2 := map[string]interface{}{
		"users": ArrayMinLike(3, map[string]interface{}{
			"user": matcher1})}
	matcher3 := map[string]interface{}{
		"users": ArrayMinLike(3, map[string]interface{}{
			"user": matcher2})}

	expectedBody := formatJSON(`{ "users": [ { "user": { "users": [ { "user": { "users": [ { "user": "someusername" }, { "user": "someusername" }, { "user": "someusername" } ] } }, { "user": { "users": [ { "user": "someusername" }, { "user": "someusername" }, { "user": "someusername" } ] } }, { "user": { "users": [ { "user": "someusername" }, { "user": "someusername" }, { "user": "someusername" } ] } } ] } }, { "user": { "users": [ { "user": { "users": [ { "user": "someusername" }, { "user": "someusername" }, { "user": "someusername" } ] } }, { "user": { "users": [ { "user": "someusername" }, { "user": "someusername" }, { "user": "someusername" } ] } }, { "user": { "users": [ { "user": "someusername" }, { "user": "someusername" }, { "user": "someusername" } ] } } ] } }, { "user": { "users": [ { "user": { "users": [ { "user": "someusername" }, { "user": "someusername" }, { "user": "someusername" } ] } }, { "user": { "users": [ { "user": "someusername" }, { "user": "someusername" }, { "user": "someusername" } ] } }, { "user": { "users": [ { "user": "someusername" }, { "user": "someusername" }, { "user": "someusername" } ] } } ] } } ] }`)
	expectedMatchingRules := matchingRuleType{
		"$.body.users": map[string]interface{}{
			"match": "type",
			"min":   3,
		},
		"$.body.users[*].user.users": map[string]interface{}{
			"match": "type",
			"min":   3,
		},
		"$.body.users[*].user.users[*].user.users": map[string]interface{}{
			"match": "type",
			"min":   3,
		},
		"$.body.users[*].user.users[*].user.users[*].user": map[string]interface{}{
			"match": "regex",
			"regex": "\\s+",
		},
	}

	body := PactBodyBuilder(matcher3)
	result := formatJSONObject(body.Body)

	if !reflect.DeepEqual(body.MatchingRules, expectedMatchingRules) {
		t.Fatalf("got '%v' wanted '%v'", body.MatchingRules, expectedMatchingRules)
	}
	if expectedBody != result {
		t.Fatalf("got '%v' wanted '%v'", result, expectedBody)
	}
}

func TestMatcher_SerialisePactFile(t *testing.T) {
	matcher := map[string]interface{}{
		"users": []interface{}{
			Like("someusername1"),
		},
		"pass": Like(1234),
		"id":   5678,
	}
	expected := formatJSON(`{
			"matchingRules": {
				"$.body.pass": {
					"match": "type"
				},
				"$.body.users[0]": {
					"match": "type"
				}
			},
			"body": {
				"id": 5678,
				"pass": 1234,
				"users": [
					"someusername1"
				]
			}
		}`)
	body := PactBodyBuilder(matcher)
	if expected != formatJSONObject(body) {
		t.Fatalf("wanted %s, got %s", expected, formatJSONObject(body))
	}
}
