package native

/*
#cgo LDFLAGS: ${SRCDIR}/../../libs/libpact_mock_server.dylib

// Library headers
typedef int bool;
#define true 1
#define false 0

int create_mock_server(char* pact, int port);
int mock_server_matched(int port);
char* mock_server_mismatches(int port);
bool cleanup_mock_server(int port);
int write_pact_file(int port, char* dir);

*/
import "C"
import (
	"encoding/json"
	"fmt"
)

// Request is the sub-struct of Mismatch
type Request struct {
	Method  string            `json:"method"`
	Path    string            `json:"path"`
	Query   string            `json:"query,omitempty"`
	Headers map[string]string `json:"headers,omitempty"`
	Body    interface{}       `json:"body,omitempty"`
}

// Mismatch is a type returned from the validation process
// [
//   {
//     "method": "GET",
//     "path": "/",
//     "request": {
//       "body": {
//         "pass": 1234,
//         "user": {
//           "address": "some address",
//           "name": "someusername",
//           "phone": 12345678,
//           "plaintext": "plaintext"
//         }
//       },
//       "method": "GET",
//       "path": "/"
//     },
//     "type": "missing-request"
//   }
// ]
type Mismatch struct {
	Request Request
	Type    string
}

// CreateMockServer creates a new Mock Server from a given Pact file
func CreateMockServer(pact string) int {
	res := C.create_mock_server(C.CString(pact), 0)
	fmt.Println("Mock Server running on port:", res)
	return int(res)
}

// Verify verifies that all interactions were successful. If not, returns a slice
// of Mismatch-es.
func Verify(port int, dir string) (bool, []Mismatch) {
	res := C.mock_server_matched(C.int(port))
	fmt.Println("Match result: ", res)

	mismatches := MockServerMismatches(port)
	fmt.Println("Mismatches! :", mismatches)

	if int(res) == 1 {
		WritePactFile(port, dir)
	}

	CleanupMockServer(port)

	return int(res) == 1, mismatches
}

// MockServerMismatches returns a JSON object containing any mismatches from
// the last set of interactions.
// TODO: create a specific struct type to marshal this into
func MockServerMismatches(port int) []Mismatch {
	var res []Mismatch

	mismatches := C.mock_server_mismatches(C.int(port))
	json.Unmarshal([]byte(C.GoString(mismatches)), &res)

	return res
}

// CleanupMockServer frees the memory from the previous mock server.
func CleanupMockServer(port int) {
	C.cleanup_mock_server(C.int(port))
}

// WritePactFile writes the Pact to file.
func WritePactFile(port int, dir string) int {
	return int(C.write_pact_file(C.int(port), C.CString(dir)))
}
