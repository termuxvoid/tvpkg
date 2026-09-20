package pkgmanager

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"sync"
)

// DryRunEnv, when set to "1", makes every command print instead of run.
const DryRunEnv = "TVPKG_DRY_RUN"

// Runner executes package-manager binaries, honoring dry-run mode.
type Runner struct {
	Prefix string
	DryRun bool
}

// NewRunner builds a Runner for the given prefix. Dry-run is enabled when
// TVPKG_DRY_RUN=1, which is useful for development on a device that does not
// have a given package manager installed.
func NewRunner(prefix string) *Runner {
	return &Runner{
		Prefix: prefix,
		DryRun: os.Getenv(DryRunEnv) == "1",
	}
}

// Stream runs the command inheriting stdio (used for interactive progress and
// for CLI mode output).
func (r *Runner) Stream(bin string, args ...string) error {
	if r.DryRun {
		r.print(bin, args)
		return nil
	}
	cmd := exec.Command(r.cmdPath(bin), args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("%s %s: %w", bin, strings.Join(args, " "), err)
	}
	return nil
}

// Capture runs the command and returns its combined output. Used by the TUI
// so the result can be rendered inside a pane instead of corrupting the
// alternate screen.
func (r *Runner) Capture(bin string, args ...string) (string, error) {
	if r.DryRun {
		r.print(bin, args)
		return "", nil
	}
	cmd := exec.Command(r.cmdPath(bin), args...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return string(out), fmt.Errorf("%s %s: %w", bin, strings.Join(args, " "), err)
	}
	return string(out), nil
}

func (r *Runner) cmdPath(bin string) string {
	p := r.Prefix + "/bin/" + bin
	if _, err := os.Stat(p); err == nil {
		return p
	}
	return bin
}

func (r *Runner) print(bin string, args []string) {
	fmt.Printf("[tvpkg:dry-run] %s %s\n", bin, strings.Join(args, " "))
}

// Live is a running process whose output is captured incrementally so a TUI
// can render it in real time next to a spinner.
type Live struct {
	Cmd *exec.Cmd
	buf *safeBuf
}

// Output returns everything the process has written so far.
func (l *Live) Output() string { return l.buf.String() }

// Wait waits for the process to finish and returns its exit error.
func (l *Live) Wait() error { return l.Cmd.Wait() }

// RunLive starts the command with streaming output capture.
func (r *Runner) RunLive(bin string, args ...string) (*Live, error) {
	if r.DryRun {
		r.print(bin, args)
		return &Live{Cmd: exec.Command("/bin/true"), buf: newSafeBuf()}, nil
	}
	cmd := exec.Command(r.cmdPath(bin), args...)
	buf := newSafeBuf()
	cmd.Stdout = buf
	cmd.Stderr = buf
	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("%s %s: %w", bin, strings.Join(args, " "), err)
	}
	return &Live{Cmd: cmd, buf: buf}, nil
}

// safeBuf is a mutex-guarded output buffer written by a live process and read
// by the TUI render loop.
type safeBuf struct {
	mu  sync.Mutex
	buf bytes.Buffer
}

func newSafeBuf() *safeBuf { return &safeBuf{} }

func (b *safeBuf) Write(p []byte) (int, error) {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.buf.Write(p)
}

func (b *safeBuf) String() string {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.buf.String()
}

// trimError returns the tail of output for error messages, keeping them short.
func trimError(out []byte) string {
	s := strings.TrimSpace(string(bytes.TrimSpace(out)))
	if s == "" {
		return s
	}
	lines := strings.Split(s, "\n")
	if len(lines) > 4 {
		return strings.Join(lines[len(lines)-4:], "\n")
	}
	return s
}
