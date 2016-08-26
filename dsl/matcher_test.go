package dsl

import (
	"bytes"
	"encoding/json"
	"fmt"
	"testing"
)

func TestMatcher_ArrayMinLike(t *testing.T) {

	expected := formatJSON(`
		{
			"match": "type",
			"min": 3
		}`)

	match := formatJSONObject(ArrayMinLike(3, "passwords", "myawesomeword").Matcher)
	fmt.Println(formatJSONObject(ArrayMinLike(3, "passwords", map[string]string{"pass": "myawesomeword"})))

	if expected != match {
		t.Fatalf("Expected Term to match. '%s' != '%s'", expected, match)
	}
}

func TestMatcher_NestedMaps(t *testing.T) {
	// matcher := ArrayMinLike(3, "passwords", map[string]interface{}{"pass": PactTerm("\\d+", 1234)})
	// b1 := BuildPact(map[string]interface{}{"users": matcher})
	matcher := map[string]interface{}{
		"user": map[string]interface{}{
			"phone":     PactTerm("\\d+", 12345678),
			"name":      PactTerm("\\s+", "someusername"),
			"address":   PactTerm("\\s+", "some address"),
			"plaintext": "plaintext",
		},
		"pass": PactTerm("\\d+", 1234),
	}
	b1 := BuildPact(matcher)
	fmt.Println(b1.Body)
	fmt.Println(formatJSONObject(b1.Body))
}

func TestMatcher_Arrays(t *testing.T) {
	// matcher := ArrayMinLike(3, "passwords", map[string]interface{}{"pass": PactTerm("\\d+", 1234)})
	// b1 := BuildPact(map[string]interface{}{"users": matcher})
	matcher := map[string]interface{}{
		"users": []interface{}{
			PactTerm("\\s+", "someusername1"),
			PactTerm("\\s+", "someusername2"),
			PactTerm("\\s+", "someusername3"),
		},
		// "user": []map[string]interface{}{
		// 	"name": PactTerm("\\s+", "someusername"),
		// },
		"pass": PactTerm("\\d+", 1234),
	}
	b1 := BuildPact(matcher)
	fmt.Println(b1.Body)
	fmt.Println(formatJSONObject(b1.Body))

	// TODO: make this a table-driven test
	// b2 := BuildPact(map[string]interface{}{"password": PactTerm("\\d+", 1234)})
	// expected := map[string]interface{}{
	// 	"password": 1234,
	// }
	// if !reflect.DeepEqual(b2.Body, expected) {
	// 	t.Fatalf("wanted %v, got %v", expected, b2.Body)
	// }
}

// Format a JSON document to make comparison easier.
func formatJSONObject(object interface{}) string {
	// var out bytes.Buffer
	out, _ := json.Marshal(object)
	return formatJSON(string(out))
}

// Format a JSON document to make comparison easier.
func formatJSON(object string) string {
	var out bytes.Buffer
	json.Indent(&out, []byte(object), "", "\t")
	return string(out.Bytes())
}
