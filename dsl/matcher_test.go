package dsl

import (
	"bytes"
	"encoding/json"
	"reflect"
	"testing"
)

func TestMatcher_ArrayMinLike(t *testing.T) {
	matcher := map[string]interface{}{
		"users": ArrayMinLike(3, map[string]interface{}{
			"user": PactTerm("\\s+", "someusername")})}

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

	dsl := BuildPact(matcher)
	result := formatJSONObject(dsl.Body)

	if expectedBody != result {
		t.Fatalf("got '%v' wanted '%v'", result, expectedBody)
	}
	if !reflect.DeepEqual(dsl.MatchingRules, expectedMatchingRules) {
		t.Fatalf("got '%v' wanted '%v'", dsl.MatchingRules, expectedMatchingRules)
	}
}

func TestMatcher_ArrayMaxLike(t *testing.T) {
	matcher := map[string]interface{}{
		"users": ArrayMaxLike(3, map[string]interface{}{
			"user": PactTerm("\\s+", "someusername")})}

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

	dsl := BuildPact(matcher)
	result := formatJSONObject(dsl.Body)

	if expectedBody != result {
		t.Fatalf("got '%v' wanted '%v'", result, expectedBody)
	}
	if !reflect.DeepEqual(dsl.MatchingRules, expectedMatchingRules) {
		t.Fatalf("got '%v' wanted '%v'", dsl.MatchingRules, expectedMatchingRules)
	}
}

func TestMatcher_NestedMaps(t *testing.T) {
	matcher := map[string]interface{}{
		"user": map[string]interface{}{
			"phone":     PactTerm("\\d+", 12345678),
			"name":      PactTerm("\\s+", "someusername"),
			"address":   PactTerm("\\s+", "some address"),
			"plaintext": "plaintext",
		},
		"pass": PactTerm("\\d+", 1234),
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

	dsl := BuildPact(matcher)
	result := formatJSONObject(dsl.Body)

	if expectedBody != result {
		t.Fatalf("got '%v' wanted '%v'", result, expectedBody)
	}
	if !reflect.DeepEqual(dsl.MatchingRules, expectedMatchingRules) {
		t.Fatalf("got '%v' wanted '%v'", dsl.MatchingRules, expectedMatchingRules)
	}
}

func TestMatcher_Arrays(t *testing.T) {
	matcher := map[string]interface{}{
		"users": []interface{}{
			PactTerm("\\s+", "someusername1"),
			PactTerm("\\s+", "someusername2"),
			PactTerm("\\s+", "someusername3"),
		},
		"pass": PactTerm("\\d+", 1234),
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

	dsl := BuildPact(matcher)
	result := formatJSONObject(dsl.Body)

	if expectedBody != result {
		t.Fatalf("got '%v' wanted '%v'", result, expectedBody)
	}
	if !reflect.DeepEqual(dsl.MatchingRules, expectedMatchingRules) {
		t.Fatalf("got '%v' wanted '%v'", dsl.MatchingRules, expectedMatchingRules)
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

	dsl := BuildPact(matcher)
	result := formatJSONObject(dsl.Body)

	if expectedBody != result {
		t.Fatalf("got '%v' wanted '%v'", result, expectedBody)
	}
	if !reflect.DeepEqual(dsl.MatchingRules, expectedMatchingRules) {
		t.Fatalf("got '%v' wanted '%v'", dsl.MatchingRules, expectedMatchingRules)
	}
}

func TestMatcher_Term(t *testing.T) {
	matcher := map[string]interface{}{
		"user": PactTerm("\\s+", "someusername3"),
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

	dsl := BuildPact(matcher)
	result := formatJSONObject(dsl.Body)

	if expectedBody != result {
		t.Fatalf("got '%v' wanted '%v'", result, expectedBody)
	}
	if !reflect.DeepEqual(dsl.MatchingRules, expectedMatchingRules) {
		t.Fatalf("got '%v' wanted '%v'", dsl.MatchingRules, expectedMatchingRules)
	}
}

func TestMatcher_DeepArrayMinLike(t *testing.T) {
	matcher1 := map[string]interface{}{
		"users": ArrayMinLike(3, map[string]interface{}{
			"user": PactTerm("\\s+", "someusername")})}
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

	dsl := BuildPact(matcher3)
	result := formatJSONObject(dsl.Body)

	if !reflect.DeepEqual(dsl.MatchingRules, expectedMatchingRules) {
		t.Fatalf("got '%v' wanted '%v'", dsl.MatchingRules, expectedMatchingRules)
	}
	if expectedBody != result {
		t.Fatalf("got '%v' wanted '%v'", result, expectedBody)
	}
}

// Format a JSON document to make comparison easier.
func formatJSONObject(object interface{}) string {
	out, _ := json.Marshal(object)
	return formatJSON(string(out))
}

// Format a JSON document to make comparison easier.
func formatJSON(object string) string {
	var out bytes.Buffer
	json.Indent(&out, []byte(object), "", "\t")
	return string(out.Bytes())
}
