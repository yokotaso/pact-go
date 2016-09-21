package dsl

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
	"testing"

	"github.com/pact-foundation/pact-go/dsl/native"
	"github.com/pact-foundation/pact-go/types"
	"github.com/pact-foundation/pact-go/utils"
)

func TestPact_setupLogging(t *testing.T) {
	res := captureOutput(func() {
		(&Pact{LogLevel: "DEBUG"}).setupLogging()
		log.Println("[DEBUG] this should display")
	})

	if !strings.Contains(res, "[DEBUG] this should display") {
		t.Fatalf("Expected log message to contain '[DEBUG] this should display' but got '%s'", res)
	}

	res = captureOutput(func() {
		(&Pact{LogLevel: "INFO"}).setupLogging()
		log.Print("[DEBUG] this should not display")
	})

	if res != "" {
		t.Fatalf("Expected log message to be empty but got '%s'", res)
	}

	res = captureOutput(func() {
		(&Pact{LogLevel: "NONE"}).setupLogging()
		log.Print("[ERROR] this should not display")
	})

	if res != "" {
		t.Fatalf("Expected log message to be empty but got '%s'", res)
	}
}

// Capture output from a log write
func captureOutput(action func()) string {
	rescueStderr := os.Stderr
	r, w, _ := os.Pipe()
	os.Stderr = w

	action()

	w.Close()
	out, _ := ioutil.ReadAll(r)
	os.Stderr = rescueStderr

	return strings.TrimSpace(string(out))
}

func TestPact_Verify(t *testing.T) {
	pact := &Pact{
		Consumer: "My Consumer",
		Provider: "My Provider",
	}
	defer pact.Teardown()

	testCalled := false
	var testFunc = func() error {
		http.Get(fmt.Sprintf("http://localhost:%d/", pact.ServerPort))
		testCalled = true
		return nil
	}

	pact.
		AddInteraction().
		Given("Some state").
		UponReceiving("Some name for the test").
		WithRequest(Request{Path: "/", Method: "GET"}).
		WillRespondWith(Response{})

	err := pact.Verify(testFunc)
	if err != nil {
		t.Fatalf("Error: %v", err)
	}

	if testCalled == false {
		t.Fatalf("Expected test function to be called but it was not")
	}
}

func TestPact_VerifyMockServerFail(t *testing.T) {
	var testFunc = func() error { return nil }

	pact := &Pact{
		Consumer: "My Consumer",
		Provider: "My Provider",
	}
	defer pact.Teardown()

	pact.
		AddInteraction().
		Given("Some state").
		UponReceiving("Some name for the test").
		WithRequest(Request{Path: "/", Method: "GET"}).
		WillRespondWith(Response{})

	err := pact.Verify(testFunc)

	if err == nil {
		t.Fatalf("Expected error but got none")
	}
}

func TestPact_WritePact(t *testing.T) {
	pact := &Pact{
		Consumer: "My Consumer",
		Provider: "My Provider",
	}

	pact.Setup()
	native.CreateMockServer(pactFileContents, pact.ServerPort)

	err := pact.WritePact()
	if err != nil {
		t.Fatalf("Error: %v", err)
	}
}

func TestPact_WritePactFail(t *testing.T) {
	pact := &Pact{
		Consumer: "My Consumer",
		Provider: "My Provider",
	}

	err := pact.WritePact()
	if err == nil {
		t.Fatalf("Expected error but got none")
	}
}

func TestPact_VerifyFail(t *testing.T) {
	var testFunc = func() error { return nil }

	pact := &Pact{}

	pact.
		AddInteraction().
		Given("Some state").
		UponReceiving("Some name for the test").
		WithRequest(Request{}).
		WillRespondWith(Response{})

	err := pact.Verify(testFunc)
	if err == nil {
		t.Fatalf("Expected error but got none")
	}

	if !strings.Contains(err.Error(), "Pact validation failed!") {
		t.Fatalf("Expected response body to contain an error message 'something went wrong' but got '%s'", err.Error())
	}
}

func TestPact_Setup(t *testing.T) {
	pact := &Pact{LogLevel: "DEBUG"}
	pact.ServerPort = 0
	pact.Setup()
	if pact.ServerPort == 0 {
		t.Fatalf("want > 0, got 0")
	}

	if pact.pactClient == nil {
		t.Fatal("want an object, got nil")
	}
}

func TestPact_Teardown(t *testing.T) {
	old := waitForPort
	defer func() { waitForPort = old }()
	waitForPort = func(int, string) error {
		return nil
	}
	port, _ := utils.GetFreePort()
	createDaemon(port, true)
	waitForPortInTest(port, t)

	pact := &Pact{Port: port, LogLevel: "DEBUG"}
	pact.Setup()
	native.CreateMockServer(pactFileContents, pact.ServerPort)

	if pact.ServerPort == 0 {
		t.Fatalf("want > 0, got 0")
	}

	pact.Teardown()
	if pact.ServerPort != 0 {
		t.Fatalf("want '0', got '%d'", pact.ServerPort)
	}
}

func TestPact_TeardownFail(t *testing.T) {
	pact := &Pact{LogLevel: "DEBUG"}
	pact.ServerPort, _ = utils.GetFreePort()

	if pact.ServerPort == 0 {
		t.Fatalf("want > 0, got 0")
	}

	pact.Teardown()
	if pact.ServerPort == 0 {
		t.Fatal("want > 0, got 0")
	}
}

func TestPact_VerifyProvider(t *testing.T) {
	old := waitForPort
	defer func() { waitForPort = old }()
	waitForPort = func(int, string) error {
		return nil
	}
	port, _ := utils.GetFreePort()
	createDaemon(port, true)
	waitForPortInTest(port, t)

	pact := &Pact{Port: port, LogLevel: "DEBUG", pactClient: &PactClient{Port: port}}
	err := pact.VerifyProvider(types.VerifyRequest{
		ProviderBaseURL: "http://www.foo.com",
		PactURLs:        []string{"foo.json", "bar.json"},
	})

	if err != nil {
		t.Fatal("Error:", err)
	}
}

func TestPact_VerifyProviderBroker(t *testing.T) {
	brokerPort := setupMockBroker(false)
	old := waitForPort
	defer func() { waitForPort = old }()
	waitForPort = func(int, string) error {
		return nil
	}
	port, _ := utils.GetFreePort()
	createDaemon(port, true)
	waitForPortInTest(port, t)

	pact := &Pact{Port: port, LogLevel: "DEBUG", pactClient: &PactClient{Port: port}, Provider: "bobby"}
	err := pact.VerifyProvider(types.VerifyRequest{
		ProviderBaseURL: "http://www.foo.com",
		BrokerURL:       fmt.Sprintf("http://localhost:%d", brokerPort),
	})

	if err != nil {
		t.Fatal("Error:", err)
	}
}

func TestPact_VerifyProviderBrokerNoConsumers(t *testing.T) {
	brokerPort := setupMockBroker(false)
	old := waitForPort
	defer func() { waitForPort = old }()
	waitForPort = func(int, string) error {
		return nil
	}
	port, _ := utils.GetFreePort()
	createDaemon(port, true)
	waitForPortInTest(port, t)

	pact := &Pact{Port: port, LogLevel: "DEBUG", pactClient: &PactClient{Port: port}, Provider: "providernotexist"}
	err := pact.VerifyProvider(types.VerifyRequest{
		ProviderBaseURL: "http://www.foo.com",
		BrokerURL:       fmt.Sprintf("http://localhost:%d", brokerPort),
	})

	if err == nil {
		t.Fatalf("Expected error but got none")
	}
}

func TestPact_VerifyProviderFail(t *testing.T) {
	old := waitForPort
	defer func() { waitForPort = old }()
	waitForPort = func(int, string) error {
		return nil
	}
	port, _ := utils.GetFreePort()
	createDaemon(port, false)
	waitForPortInTest(port, t)

	pact := &Pact{Port: port, LogLevel: "DEBUG", pactClient: &PactClient{Port: port}}
	err := pact.VerifyProvider(types.VerifyRequest{
		ProviderBaseURL: "http://www.foo.com",
		PactURLs:        []string{"foo.json", "bar.json"},
	})

	if err == nil {
		t.Fatalf("Expected error but got none")
	}
}

func TestPact_AddInteraction(t *testing.T) {
	pact := &Pact{}

	pact.
		AddInteraction().
		Given("Some state").
		UponReceiving("Some name for the test").
		WithRequest(Request{}).
		WillRespondWith(Response{})

	pact.
		AddInteraction().
		Given("Some state2").
		UponReceiving("Some name for the test2").
		WithRequest(Request{}).
		WillRespondWith(Response{})

	if len(pact.Interactions) != 2 {
		t.Fatalf("Expected 2 interactions to be added to Pact but got %d", len(pact.Interactions))
	}
}
