package release

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

type Runner struct {
	Out io.Writer
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
	cmd := exec.Command(name, args...)
	cmd.Dir = dir
	cmd.Stdout = io.MultiWriter(r.Out, logFile)
	cmd.Stderr = io.MultiWriter(r.Out, logFile)
	if interactive {
		cmd.Stdin = os.Stdin
	}
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("%s failed: %w (see %s)", name, err, logPath)
	}
	return nil
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
