package dsl

import (
	"fmt"
	"log"
	"net"
	"net/rpc"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/pact-foundation/pact-go/types"
)

var (
	timeoutDuration       = 1 * time.Second
	commandVerifyProvider = "Daemon.VerifyProvider"
	commandStopDaemon     = "Daemon.StopDaemon"
)

// Client is the simplified remote interface to the Pact Daemon.
type Client interface {
	StartServer() *types.MockServer
}

// PactClient is the default implementation of the Client interface.
type PactClient struct {
	// Port the daemon is running on.
	Port int
}

// // Get a port given a URL
func getPort(rawURL string) int {
	parsedURL, err := url.Parse(rawURL)
	if err == nil {
		if len(strings.Split(parsedURL.Host, ":")) == 2 {
			port, err := strconv.Atoi(strings.Split(parsedURL.Host, ":")[1])
			if err == nil {
				return port
			}
		}
		if parsedURL.Scheme == "https" {
			return 443
		}
		return 80
	}

	return -1
}

func getHTTPClient(port int) (*rpc.Client, error) {
	log.Println("[DEBUG] creating an HTTP client")
	err := waitForPort(port, fmt.Sprintf(`Timed out waiting for Daemon on port %d - are you
		sure it's running?`, port))
	if err != nil {
		return nil, err
	}
	return rpc.DialHTTP("tcp", fmt.Sprintf(":%d", port))
}

// Use this to wait for a daemon to be running prior
// to running tests.
var waitForPort = func(port int, message string) error {
	log.Println("[DEBUG] waiting for port", port, "to become available")
	timeout := time.After(timeoutDuration)

	for {
		select {
		case <-timeout:
			log.Printf("[ERROR] Expected server to start < %s. %s", timeoutDuration, message)
			return fmt.Errorf("Expected server to start < %s. %s", timeoutDuration, message)
		case <-time.After(50 * time.Millisecond):
			_, err := net.Dial("tcp", fmt.Sprintf(":%d", port))
			if err == nil {
				return nil
			}
		}
	}
}

// VerifyProvider runs the verification process against a running Provider.
func (p *PactClient) VerifyProvider(request types.VerifyRequest) (string, error) {
	log.Println("[DEBUG] client: verifying a provider")

	port := getPort(request.ProviderBaseURL)

	waitForPort(port, fmt.Sprintf(`Timed out waiting for Provider API to start
		 on port %d - are you sure it's running?`, port))

	var res types.CommandResponse
	client, err := getHTTPClient(p.Port)
	if err == nil {
		err = client.Call(commandVerifyProvider, request, &res)
		if err != nil {
			log.Println("[ERROR] rpc: ", err.Error())
		}
	}

	if res.ExitCode > 0 {
		return "", fmt.Errorf("provider verification failed: %s", res.Message)
	}

	return res.Message, err
}

// StopDaemon remotely shuts down the Pact Daemon.
func (p *PactClient) StopDaemon() error {
	log.Println("[DEBUG] client: stop daemon")
	var req, res string
	client, err := getHTTPClient(p.Port)
	if err == nil {
		err = client.Call(commandStopDaemon, &req, &res)
		if err != nil {
			log.Println("[ERROR] rpc:", err.Error())
		}
	}
	return err
}
