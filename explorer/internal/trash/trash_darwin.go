package trash

import (
	"fmt"
	"unsafe"
)

/*
#cgo LDFLAGS: -framework Foundation
#include <stdlib.h>

int MoveToTrash(const char* path);
*/
import "C"

// throwToTrash moves file to trash bin in Darwin based OS.
// When running in sandbox, the app need to declare com.apple.security.files.user-selected.read-write
// permissions in 'Entitlements' file.
func ThrowToTrash(path string) error {
	cPath := C.CString(path)
	defer C.free(unsafe.Pointer(cPath))

	result := C.MoveToTrash(cPath)
	if result != 0 {
		return fmt.Errorf("failed to move file to trash: %s", path)
	}
	return nil
}
