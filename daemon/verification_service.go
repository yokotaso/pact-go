package daemon

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/kardianos/osext"
)

// VerificationService is a wrapper for the Pact Provider Verifier Service.
type VerificationService struct {
	ServiceManager
}

// NewService creates a new VerificationService with default settings.
// Arguments allowed:
//
// 		--provider-base-url
// 		--pact-urls
// 		--provider-states-url
// 		--provider-states-setup-url
// 		--broker-username
// 		--broker-password
//    --publish_verification_results
//    --provider_app_version
func (m *VerificationService) NewService(args []string) (int, Service) {
	log.Printf("[DEBUG] starting verification service with args: %v\n", args)

	m.Args = args
	m.Env = append(os.Environ(), `PACT_INTERACTION_RERUN_COMMAND="To re-run this specific test, set the following environment variables and run your test again: PACT_DESCRIPTION=\"<PACT_DESCRIPTION>\" PACT_PROVIDER_STATE=\"<PACT_PROVIDER_STATE>\""`)

	m.Command = getVerifierCommandPath()
	return -1, m
}

func getVerifierCommandPath() string {
	dir, _ := osext.ExecutableFolder()
	return fmt.Sprintf(filepath.Join(dir, "pact", "bin", "pact-provider-verifier"))
}
