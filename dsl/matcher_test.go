package dsl

import (
	"reflect"
	"testing"
)

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
