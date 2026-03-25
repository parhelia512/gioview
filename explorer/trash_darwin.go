package explorer

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

/*
#cgo LDFLAGS: -framework Foundation
#include <Foundation/Foundation.h>

int MoveToTrash(const char* path) {
    @autoreleasepool {
        NSString *nsPath = [NSString stringWithUTF8String:path];
        NSURL *url = [NSURL fileURLWithPath:nsPath];
        NSError *error = nil;
        
        // The macOS “Portal” API
		// It will handle sandbox permission, cross-partition operation
		// and generate .DS_Store recovery data.
        BOOL success = [[NSFileManager defaultManager] 
                        trashItemAtURL:url 
                        resultingItemURL:nil 
                        error:&error];
        
        return success ? 0 : 1;
    }
}
*/

import "C"
import (
	"fmt"
	"unsafe"
)

// throwToTrash moves file to trash bin in Darwin based OS.
// When running in sandbox, the app need to declare com.apple.security.files.user-selected.read-write 
// permissions in 'Entitlements' file.
func throwToTrash(path string) error {
	cPath := C.CString(path)
	defer C.free(unsafe.Pointer(cPath))

	result := C.MoveToTrash(cPath)
	if result != 0 {
		return fmt.Errorf("macOS failed to move file to trash: %s", path)
	}
	return nil
}
