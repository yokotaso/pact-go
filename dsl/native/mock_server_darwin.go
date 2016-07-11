package native

/*
#cgo LDFLAGS: ${SRCDIR}/../../libs/libpact_mock_server.dylib

// Library headers
typedef int bool;
#define true 1
#define false 0

int create_mock_server(char* pact, int port);
int mock_server_matched(int port);
*/
import "C"
import "fmt"

// CreateMockServer creates a new Mock Server from a given Pact file
func CreateMockServer(pact string) int {
	res := C.create_mock_server(C.CString(pact), 0)
	fmt.Println("Mock Server running on port:", res)
	return int(res)
}

// Verify verifies that all interactions were successful. If not, returns a slice
// of Mismatch-es.
func Verify(port int) bool {
	res := C.mock_server_matched(C.int(port))
	fmt.Println("Match result: ", res)
	return int(res) == 0
}
