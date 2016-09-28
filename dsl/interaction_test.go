package dsl

import (
	"encoding/json"
	"reflect"
	"testing"
)

func TestInteraction_NewInteraction(t *testing.T) {
	i := (&Interaction{}).
		Given("Some state").
		UponReceiving("Some name for the test").
		WithRequest(Request{}).
		WillRespondWith(Response{})

	if i.State != "Some state" {
		t.Fatalf("Expected 'Some state' but got '%s'", i.State)
	}
	if i.Description != "Some name for the test" {
		t.Fatalf("Expected 'Some name for the test' but got '%s'", i.Description)
	}
}

func TestInteraction_WithRequest(t *testing.T) {
	// Pass in plain string, should be left alone
	i := (&Interaction{}).
		Given("Some state").
		UponReceiving("Some name for the test").
		WithRequest(Request{
			Body: "somestring",
		})

	content, ok := i.Request.Body.(string)

	if !ok {
		t.Fatalf("must be a string")
	}

	if content != "somestring" {
		t.Fatalf("Expected 'somestring' but got '%s'", content)
	}

	// structured string should be changed to an interface{}
	i = (&Interaction{}).
		Given("Some state").
		UponReceiving("Some name for the test").
		WithRequest(Request{
			Body: `{
			"foo": "bar",
			"baz": "bat"
			}`,
		})

	obj := map[string]string{
		"foo": "bar",
		"baz": "bat",
	}

	var expect interface{}
	body, _ := json.Marshal(obj)
	json.Unmarshal(body, &expect)

	if _, ok := i.Request.Body.(map[string]interface{}); !ok {
		t.Fatalf("Expected response to be of type 'map[string]string'")
	}

	if !reflect.DeepEqual(i.Request.Body, expect) {
		t.Fatalf("Expected response object body '%v' to match '%v'", i.Request.Body, expect)
	}
}

func TestInteraction_WillRespondWith(t *testing.T) {
	// Pass in plain string, should be left alone
	i := (&Interaction{}).
		Given("Some state").
		UponReceiving("Some name for the test").
		WithRequest(Request{}).
		WillRespondWith(Response{
			Body: "somestring",
		})

	content, ok := i.Response.Body.(string)

	if !ok {
		t.Fatalf("must be a string")
	}

	if content != "somestring" {
		t.Fatalf("Expected 'somestring' but got '%s'", content)
	}

	// structured string should be changed to an interface{}
	i = (&Interaction{}).
		Given("Some state").
		UponReceiving("Some name for the test").
		WithRequest(Request{}).
		WillRespondWith(Response{
			Body: `{
				"foo": "bar",
				"baz": "bat"
			}`,
		})

	obj := map[string]string{
		"foo": "bar",
		"baz": "bat",
	}

	var expect interface{}
	body, _ := json.Marshal(obj)
	json.Unmarshal(body, &expect)

	if _, ok := i.Response.Body.(map[string]interface{}); !ok {
		t.Fatalf("Expected response to be of type 'map[string]string'")
	}

	if !reflect.DeepEqual(i.Response.Body, expect) {
		t.Fatalf("Expected response object body '%v' to match '%v'", i.Response.Body, expect)
	}
}

func TestInteraction_toObject(t *testing.T) {
	// unstructured string should not be changed
	res := toObject([]byte("somestring"))
	content, ok := res.(string)

	if !ok {
		t.Fatalf("must be a string")
	}

	if content != "somestring" {
		t.Fatalf("Expected 'somestring' but got '%s'", content)
	}

	// errors should return a string repro of original interface{}
	res = toObject([]byte(""))
	content, ok = res.(string)

	if !ok {
		t.Fatalf("must be a string")
	}

	if content != "" {
		t.Fatalf("Expected '' but got '%s'", content)
	}
}

func TestInteraction_WithHeaderMatchers(t *testing.T) {
	// Pass in plain string, should be left alone
	i := (&Interaction{}).
		Given("Some state").
		UponReceiving("Some name for the test").
		WithRequest(Request{
			Headers: map[string]interface{}{
				"FOO": "bar",
				"BAZ": Regex("\\d+", 1234),
			},
			Path: Regex("\\d+", 1234),
			Body: `{"foo": "bar"}`,
		}).
		WillRespondWith(Response{
			Status: 200,
		})

	expected := formatJSON(`{
			"request": {
				"method": "",
				"path": 1234,
				"headers": {
					"BAZ": 1234,
					"FOO": "bar"
				},
				"body": {
					"foo": "bar"
				}
			},
			"response": {
				"status": 200
			},
			"description": "Some name for the test",
			"provider_state": "Some state",
			"matchingRules": {
				"$.headers.BAZ": {
					"match": "regex",
					"regex": "\\d+"
				},
				"$.path": {
					"match": "regex",
					"regex": "\\d+"
				}
			}
		}`)

	if expected != formatJSONObject(i) {
		t.Fatalf("Expected %s, got %s", expected, formatJSONObject(i))
	}
}

func TestInteraction_WithPactBodyBuilderRequestAsBody(t *testing.T) {
	matcher := map[string]interface{}{
		"user": map[string]interface{}{
			"phone":     Regex("\\d+", 12345678),
			"name":      Regex("\\s+", "someusername"),
			"address":   Regex("\\s+", "some address"),
			"plaintext": "plaintext",
		},
		"pass": Regex("\\d+", 1234),
	}

	// Pass in plain string, should be left alone
	i := (&Interaction{}).
		Given("Some state").
		UponReceiving("Some name for the test").
		WithRequest(Request{
			Path: "/",
			Body: PactBodyBuilder(matcher),
		})
	expected := formatJSON(`{
			"request": {
				"method": "",
				"path": "/",
				"body": {
					"pass": 1234,
					"user": {
						"address": "some address",
						"name": "someusername",
						"phone": 12345678,
						"plaintext": "plaintext"
					}
				}
			},
			"response": {
				"status": 0
			},
			"description": "Some name for the test",
			"provider_state": "Some state",
			"matchingRules": {
				"$.body.pass": {
					"match": "regex",
					"regex": "\\d+"
				},
				"$.body.user.address": {
					"match": "regex",
					"regex": "\\s+"
				},
				"$.body.user.name": {
					"match": "regex",
					"regex": "\\s+"
				},
				"$.body.user.phone": {
					"match": "regex",
					"regex": "\\d+"
				}
			}
		}`)

	if expected != formatJSONObject(i) {
		t.Fatalf("Expected %s, got %s", expected, formatJSONObject(i))
	}
}
