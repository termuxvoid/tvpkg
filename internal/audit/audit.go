// Package audit persists a record of every mutating package operation so
// users can see what was installed, removed or upgraded and when. Logging is
// best-effort: a missing log directory or unwritable file never fails the
// underlying operation.
package audit

import (
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// Log writes one line "RFC3339 <op> <targets> <ok|error>" to
// $PREFIX/var/log/tvpkg.log. pass err == nil on success.
func Log(prefix, op, targets string, err error) {
	if prefix == "" {
		prefix = "/data/data/com.termux/files/usr"
	}
	dir := filepath.Join(prefix, "var", "log")
	if errMK := os.MkdirAll(dir, 0o755); errMK != nil {
		return
	}
	f, errOpen := os.OpenFile(filepath.Join(dir, "tvpkg.log"),
		os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
	if errOpen != nil {
		return
	}
	defer f.Close()

	status := "ok"
	if err != nil {
		status = "error: " + err.Error()
	}
	fmt.Fprintf(f, "%s %s %s %s\n", time.Now().Format(time.RFC3339), op, targets, status)
}