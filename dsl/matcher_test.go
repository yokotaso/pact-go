package dsl

import (
	"bytes"
	"encoding/json"
	"testing"
)

func TestMatcher_ArrayMinLike(t *testing.T) {
	matcher := map[string]interface{}{
		"users": ArrayMinLike(3, map[string]interface{}{
			"billy": PactTerm("\\s+", "someusername")})}

	expected := formatJSON(`{
		"users": [
			{
				"billy": "someusername"
			},
			{
				"billy": "someusername"
			},
			{
				"billy": "someusername"
			}
		]
	}`)

	result := formatJSONObject(BuildPact(matcher).Body)

	if expected != result {
		t.Fatalf("got '%s' wanted '%s'", result, expected)
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

	expected := formatJSON(`{
		"pass": 1234,
		"user": {
			"address": "some address",
			"name": "someusername",
			"phone": 12345678,
			"plaintext": "plaintext"
		}
	}`)

	result := formatJSONObject(BuildPact(matcher).Body)

	if expected != result {
		t.Fatalf("got '%s' wanted '%s'", result, expected)
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
	}
	expected := formatJSON(`{
		"pass": 1234,
		"users": [
			"someusername1",
			"someusername2",
			"someusername3"
		]
	}`)

	result := formatJSONObject(BuildPact(matcher).Body)

	if expected != result {
		t.Fatalf("got '%s' wanted '%s'", result, expected)
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
	expected := formatJSON(`{
		"pass": 1234,
		"users": [
			"someusername1",
			"someusername2",
			"someusername3"
		]
	}`)

	result := formatJSONObject(BuildPact(matcher).Body)

	if expected != result {
		t.Fatalf("got '%s' wanted '%s'", result, expected)
	}
}

func TestMatcher_Term(t *testing.T) {
	matcher := map[string]interface{}{
		"user": PactTerm("\\s+", "someusername3"),
	}
	expected := formatJSON(`{
		"user":	"someusername3"
	}`)

	result := formatJSONObject(BuildPact(matcher).Body)

	if expected != result {
		t.Fatalf("got '%s' wanted '%s'", result, expected)
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
