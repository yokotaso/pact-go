/*
Package daemon implements the RPC server side interface to remotely manage
external Pact dependencies: The Pact Mock Service and Provider Verification
"binaries."

See https://github.com/pact-foundation/pact-provider-verifier and
https://github.com/bethesque/pact-mock_service for more on the Ruby "binaries".

NOTE: The ultimate goal here is to replace the Ruby dependencies with a shared
library (Pact Reference - (https://github.com/pact-foundation/pact-reference/).
*/
package daemon

// Runs the RPC daemon for remote communication

import (
	"bytes"
	"fmt"
	"log"
	"net"
	"net/http"
	"net/rpc"
	"os"
	"os/signal"

	"github.com/pact-foundation/pact-go/types"
)

// Daemon wraps the commands for the RPC server.
type Daemon struct {
	verificationSvcManager Service
	signalChan             chan os.Signal
}

// NewDaemon returns a new Daemon with all instance variables initialised.
func NewDaemon(verificationServiceManager Service) *Daemon {
	verificationServiceManager.Setup()

	return &Daemon{
		verificationSvcManager: verificationServiceManager,
		signalChan:             make(chan os.Signal, 1),
	}
}

// StartDaemon starts the daemon RPC server.
func (d Daemon) StartDaemon(port int) {
	log.Println("[INFO] daemon - starting daemon on port", port)

	serv := rpc.NewServer()
	serv.Register(d)

	// Workaround for multiple RPC ServeMux's
	oldMux := http.DefaultServeMux
	mux := http.NewServeMux()
	http.DefaultServeMux = mux

	serv.HandleHTTP(rpc.DefaultRPCPath, rpc.DefaultDebugPath)

	// Workaround for multiple RPC ServeMux's
	http.DefaultServeMux = oldMux

	l, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		panic(err)
	}
	go http.Serve(l, mux)

	// Wait for sigterm
	signal.Notify(d.signalChan, os.Interrupt, os.Kill)
	s := <-d.signalChan
	log.Println("[INFO] daemon - received signal:", s, ", shutting down all services")

	d.Shutdown()
}

// StopDaemon allows clients to programmatically shuts down the running Daemon
// via RPC.
func (d Daemon) StopDaemon(request string, reply *string) error {
	log.Println("[DEBUG] daemon - stop daemon")
	d.signalChan <- os.Interrupt
	return nil
}

// Shutdown ensures all services are cleanly destroyed.
func (d Daemon) Shutdown() {
	log.Println("[DEBUG] daemon - shutdown")
}

// VerifyProvider runs the Pact Provider Verification Process.
func (d Daemon) VerifyProvider(request types.VerifyRequest, reply *types.CommandResponse) error {
	log.Println("[DEBUG] daemon - verifying provider")
	exitCode := 1

	// Convert request into flags, and validate request
	err := request.Validate()
	if err != nil {
		*reply = types.CommandResponse{
			ExitCode: exitCode,
			Message:  err.Error(),
		}
		return nil
	}

	var out bytes.Buffer
	_, svc := d.verificationSvcManager.NewService(request.Args)
	cmd, err := svc.Run(&out)

	if cmd.ProcessState.Success() && err == nil {
		exitCode = 0
	}

	*reply = types.CommandResponse{
		ExitCode: exitCode,
		Message:  string(out.Bytes()),
	}

	return nil
}
