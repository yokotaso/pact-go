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

func TestMatcher_RecursivePath(t *testing.T) {
	matcher := ArrayMinLike(3, "passwords", map[string]interface{}{"pass": PactTerm("\\d+", 1234)})
	matcher2 := ArrayMinLike(3, "passwords", matcher)
	matcher3 := ArrayMinLike(3, "passwords", matcher2)
	BuildPact(map[string]interface{}{"passwords": matcher3})
	// fmt.Println(b)
	// fmt.Println(formatJSONObject(b))
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
