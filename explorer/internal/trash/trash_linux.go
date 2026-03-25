package trash

import (
	"fmt"
	"os/exec"
)

// ThrowToTrash moves the file to system Trash bin, using
// the gio tool (https://docs.gtk.org/gio/). Using os.Rename
// in go will get 'permission denied' error in sandbox environment
// of Flatpak. To make it work across host OS and sandbox env, it's
// better to use gio here.
func ThrowToTrash(filePath string) error {
	cmd := exec.Command("gio", "trash", filePath)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to move file to trash: %s, error: %v", string(output), err)
	}

	return nil
}
