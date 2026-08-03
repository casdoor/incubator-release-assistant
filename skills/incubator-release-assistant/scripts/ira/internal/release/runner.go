package release

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"
)

type Runner struct {
	Out              io.Writer
	ProgressInterval time.Duration
}

type lockedWriter struct {
	mu sync.Mutex
	w  io.Writer
}

func (w *lockedWriter) Write(p []byte) (int, error) {
	w.mu.Lock()
	defer w.mu.Unlock()
	return w.w.Write(p)
}

func (r Runner) Run(logPath, dir, name string, args ...string) error {
	return r.run(false, logPath, dir, name, args...)
}

func (r Runner) RunInteractive(logPath, dir, name string, args ...string) error {
	return r.run(true, logPath, dir, name, args...)
}

func (r Runner) run(interactive bool, logPath, dir, name string, args ...string) error {
	if r.Out == nil {
		r.Out = os.Stdout
	}
	if logPath != "" {
		if err := os.MkdirAll(filepath.Dir(logPath), 0o700); err != nil {
			return err
		}
	}
	logFile, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o600)
	if err != nil {
		return err
	}
	defer logFile.Close()
	sink := &lockedWriter{w: io.MultiWriter(r.Out, logFile)}
	started := time.Now()
	displayDir := dir
	if displayDir == "" {
		displayDir, _ = os.Getwd()
	}
	if abs, absErr := filepath.Abs(displayDir); absErr == nil {
		displayDir = abs
	}
	displayLog := logPath
	if abs, absErr := filepath.Abs(logPath); absErr == nil {
		displayLog = abs
	}
	command := renderCommand(name, args)
	fmt.Fprintf(sink, "[IRA] START time=%s\n[IRA] COMMAND %s\n[IRA] WORKDIR %s\n[IRA] LOG %s\n",
		started.Format(time.RFC3339), command, displayDir, displayLog)

	cmd := exec.Command(name, args...)
	cmd.Dir = dir
	cmd.Stdout = sink
	cmd.Stderr = sink
	if interactive {
		cmd.Stdin = os.Stdin
	}
	if err := cmd.Start(); err != nil {
		fmt.Fprintf(sink, "[IRA] FAILED duration=%s error=%q\n", time.Since(started).Round(time.Millisecond), err)
		return fmt.Errorf("%s failed to start: %w (see %s)", name, err, logPath)
	}
	fmt.Fprintf(sink, "[IRA] PID %d\n", cmd.Process.Pid)

	interval := r.ProgressInterval
	if interval <= 0 {
		interval = 30 * time.Second
	}
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	done := make(chan error, 1)
	go func() { done <- cmd.Wait() }()
	for {
		select {
		case err := <-done:
			duration := time.Since(started).Round(time.Millisecond)
			if err != nil {
				fmt.Fprintf(sink, "[IRA] FAILED duration=%s error=%q\n", duration, err)
				return fmt.Errorf("%s failed after %s: %w (see %s)", name, duration, err, logPath)
			}
			fmt.Fprintf(sink, "[IRA] DONE duration=%s\n", duration)
			return nil
		case <-ticker.C:
			fmt.Fprintf(sink, "[IRA] RUNNING elapsed=%s pid=%d command=%s log=%s\n",
				time.Since(started).Round(time.Second), cmd.Process.Pid, command, displayLog)
		}
	}
}

func renderCommand(name string, args []string) string {
	parts := make([]string, 0, len(args)+1)
	parts = append(parts, quoteCommandArg(name))
	for _, arg := range args {
		parts = append(parts, quoteCommandArg(arg))
	}
	return strings.Join(parts, " ")
}

func quoteCommandArg(value string) string {
	if value == "" || strings.ContainsAny(value, " \t\r\n\"") {
		return strconv.Quote(value)
	}
	return value
}

func (r Runner) Output(dir, name string, args ...string) (string, error) {
	cmd := exec.Command(name, args...)
	cmd.Dir = dir
	raw, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("%s %s failed: %w\n%s", name, strings.Join(args, " "), err, raw)
	}
	return strings.TrimSpace(string(raw)), nil
}

func commandExists(name string) error {
	if _, err := exec.LookPath(name); err != nil {
		return fmt.Errorf("required command %q was not found", name)
	}
	return nil
}

func writeText(path, value string) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return err
	}
	return os.WriteFile(path, []byte(value), 0o600)
}

func readChecksum(path, artifactName string) (string, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	expectedLength := 128 + 2 + len(artifactName) + 1
	if len(raw) != expectedLength {
		return "", fmt.Errorf("checksum file must contain exactly one canonical LF-terminated line")
	}
	if raw[len(raw)-1] != '\n' || bytes.ContainsRune(raw, '\r') {
		return "", fmt.Errorf("checksum file must end with exactly one LF and contain no CR bytes")
	}
	digest := string(raw[:128])
	if string(raw[128:]) != "  "+artifactName+"\n" {
		return "", fmt.Errorf("checksum format must be '<128 lowercase hex>  %s'", artifactName)
	}
	for _, ch := range digest {
		if !strings.ContainsRune("0123456789abcdef", ch) {
			return "", fmt.Errorf("checksum must use lowercase hexadecimal")
		}
	}
	return digest, nil
}
