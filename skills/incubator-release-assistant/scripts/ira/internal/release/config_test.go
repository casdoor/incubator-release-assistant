package release

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func validConfig(t *testing.T) *Config {
	t.Helper()
	raw, err := os.ReadFile(filepath.Join("..", "..", "..", "..", "assets", "examples", "casbin-go.json"))
	if err != nil {
		t.Fatal(err)
	}
	var cfg Config
	if err := json.Unmarshal(raw, &cfg); err != nil {
		t.Fatal(err)
	}
	cfg.Source.Commit = "1234567890abcdef1234567890abcdef12345678"
	cfg.Signing.ApacheID = "release-manager"
	cfg.Signing.Fingerprint = "ABCDEF1234567890ABCDEF1234567890ABCDEF12"
	cfg.Raw, _ = json.Marshal(cfg)
	cfg.Path = filepath.Join(t.TempDir(), "release.json")
	return &cfg
}

func TestValidCasbinConfig(t *testing.T) {
	cfg := validConfig(t)
	if err := cfg.Validate(); err != nil {
		t.Fatalf("valid config rejected: %v", err)
	}
	if cfg.RunID() != "casbin-3.11.0-incubating-rc2" {
		t.Fatalf("unexpected run id: %s", cfg.RunID())
	}
}

func TestRejectsUnsupportedAdapter(t *testing.T) {
	cfg := validConfig(t)
	cfg.Project.Adapter = "java"
	err := cfg.Validate()
	if err == nil || !strings.Contains(err.Error(), "no other adapter is implemented") {
		t.Fatalf("expected unsupported adapter error, got %v", err)
	}
}

func TestRejectsUnsafeAndNonASFConfiguration(t *testing.T) {
	cfg := validConfig(t)
	cfg.Source.ArchivePrefix = "../../escape"
	cfg.Distribution.DevURL = "https://example.com/upload"
	cfg.Checks.RequiredFiles = []string{"LICENSE", "NOTICE", "go.mod", "go.sum", ".rat-excludes"}
	err := cfg.Validate()
	if err == nil {
		t.Fatal("unsafe config was accepted")
	}
	for _, expected := range []string{"unsafe path", "official Casbin", "DISCLAIMER"} {
		if !strings.Contains(err.Error(), expected) {
			t.Errorf("missing %q in error: %v", expected, err)
		}
	}
}

func TestLoadRejectsUnknownFields(t *testing.T) {
	cfg := validConfig(t)
	var object map[string]any
	if err := json.Unmarshal(cfg.Raw, &object); err != nil {
		t.Fatal(err)
	}
	object["surprise"] = true
	raw, _ := json.Marshal(object)
	path := filepath.Join(t.TempDir(), "invalid.json")
	if err := os.WriteFile(path, raw, 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := LoadConfig(path); err == nil || !strings.Contains(err.Error(), "unknown field") {
		t.Fatalf("expected unknown-field rejection, got %v", err)
	}
}

func TestLoadAcceptsWindowsPowerShellUTF8BOM(t *testing.T) {
	cfg := validConfig(t)
	raw := append([]byte{0xEF, 0xBB, 0xBF}, cfg.Raw...)
	path := filepath.Join(t.TempDir(), "release.json")
	if err := os.WriteFile(path, raw, 0o600); err != nil {
		t.Fatal(err)
	}
	loaded, err := LoadConfig(path)
	if err != nil {
		t.Fatalf("UTF-8 BOM config rejected: %v", err)
	}
	if loaded.RunID() != cfg.RunID() {
		t.Fatalf("unexpected run id: %s", loaded.RunID())
	}
}

func TestChecksumFormat(t *testing.T) {
	path := filepath.Join(t.TempDir(), "artifact.sha512")
	digest := strings.Repeat("a", 128)
	valid := digest + "  artifact.tar.gz\n"
	if err := os.WriteFile(path, []byte(valid), 0o600); err != nil {
		t.Fatal(err)
	}
	got, err := readChecksum(path, "artifact.tar.gz")
	if err != nil || got != digest {
		t.Fatalf("checksum rejected: %s %v", got, err)
	}

	invalid := map[string]string{
		"CRLF":             digest + "  artifact.tar.gz\r\n",
		"missing final LF": digest + "  artifact.tar.gz",
		"BOM":              "\ufeff" + valid,
		"uppercase digest": strings.Repeat("A", 128) + "  artifact.tar.gz\n",
		"path prefix":      digest + "  ./artifact.tar.gz\n",
		"star format":      digest + " *artifact.tar.gz\n",
		"extra line":       valid + "\n",
	}
	for name, content := range invalid {
		t.Run(name, func(t *testing.T) {
			if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
				t.Fatal(err)
			}
			if _, err := readChecksum(path, "artifact.tar.gz"); err == nil {
				t.Fatalf("non-canonical checksum was accepted: %q", content)
			}
		})
	}
}

func TestRemoveContained(t *testing.T) {
	root := t.TempDir()
	child := filepath.Join(root, "child")
	if err := os.Mkdir(child, 0o700); err != nil {
		t.Fatal(err)
	}
	if err := removeContained(child, root); err != nil {
		t.Fatal(err)
	}
	if err := removeContained(root, root); err == nil {
		t.Fatal("allowed root itself could be removed")
	}
}

func TestStateCanBeSavedRepeatedlyOnWindows(t *testing.T) {
	runRoot := t.TempDir()
	state := NewState(validConfig(t))
	if err := state.Save(runRoot); err != nil {
		t.Fatal(err)
	}
	state.Prepared = true
	if err := state.Save(runRoot); err != nil {
		t.Fatalf("second state save failed: %v", err)
	}
	loaded, err := LoadState(runRoot)
	if err != nil || !loaded.Prepared {
		t.Fatalf("updated state was not persisted: %+v %v", loaded, err)
	}
}

func TestCasbinGoSandboxMountsOnlyReadOnlySource(t *testing.T) {
	cfg := validConfig(t)
	args := casbinGoTestArgs(cfg, filepath.Join("safe", "extracted"))
	joined := strings.Join(args, " ")
	for _, required := range []string{"--read-only", "--cap-drop=ALL", "no-new-privileges", "dst=/input,readonly", "go test ./..."} {
		if !strings.Contains(joined, required) {
			t.Errorf("sandbox arguments missing %q: %s", required, joined)
		}
	}
	for _, forbidden := range []string{".gnupg", ".ssh", "artifacts", "dst=/root", "--privileged"} {
		if strings.Contains(strings.ToLower(joined), forbidden) {
			t.Errorf("sandbox arguments contain forbidden mount/flag %q: %s", forbidden, joined)
		}
	}
}

func TestRATCountSupportsDetailedReportFormat(t *testing.T) {
	report := []byte("  Unapproved:         0    A count of unapproved licenses.\n  Unknown:            0    A count of unknown file types.\n")
	for _, label := range []string{"Unapproved", "Unknown"} {
		count, err := ratCount(report, label)
		if err != nil || count != 0 {
			t.Fatalf("%s count failed: %d %v", label, count, err)
		}
	}
}

func TestSignRejectsWrongDigestBeforeAccessingKey(t *testing.T) {
	cfg := validConfig(t)
	runRoot, err := cfg.RunRoot()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(runRoot, "artifacts"), 0o700); err != nil {
		t.Fatal(err)
	}
	artifact := filepath.Join(runRoot, "artifacts", cfg.ArtifactName())
	if err := os.WriteFile(artifact, []byte("fixture artifact"), 0o600); err != nil {
		t.Fatal(err)
	}
	digest, err := sha512File(artifact)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(artifact+".sha512", []byte(digest+"  "+cfg.ArtifactName()+"\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	checksumDigest, err := sha512File(artifact + ".sha512")
	if err != nil {
		t.Fatal(err)
	}
	state := NewState(cfg)
	state.Prepared = true
	state.ArtifactSHA512 = digest
	state.ChecksumSHA512 = checksumDigest
	if err := state.Save(runRoot); err != nil {
		t.Fatal(err)
	}
	_, err = (Engine{}).Sign(cfg, "wrong-digest")
	if err == nil || !strings.Contains(err.Error(), "confirmation") {
		t.Fatalf("expected confirmation failure, got %v", err)
	}
}

func TestExactFileDigestRejectsChangedBytes(t *testing.T) {
	path := filepath.Join(t.TempDir(), "candidate.asc")
	if err := os.WriteFile(path, []byte("original signature bytes\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	digest, err := sha512File(path)
	if err != nil {
		t.Fatal(err)
	}
	if err := verifyExactFileDigest(path, digest, "signature"); err != nil {
		t.Fatalf("original bytes rejected: %v", err)
	}
	if err := os.WriteFile(path, []byte("different signature bytes\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := verifyExactFileDigest(path, digest, "signature"); err == nil || !strings.Contains(err.Error(), "frozen candidate") {
		t.Fatalf("changed bytes were accepted: %v", err)
	}
}

func TestRejectsLegacyStateWithoutExactByteDigests(t *testing.T) {
	cfg := validConfig(t)
	state := NewState(cfg)
	state.SchemaVersion = 1
	if err := state.VerifyConfig(cfg); err == nil || !strings.Contains(err.Error(), "exact-byte resume") {
		t.Fatalf("legacy state was accepted: %v", err)
	}
}

func TestRejectsArchiveSymlinkEscapingSourceRoot(t *testing.T) {
	root := t.TempDir()
	link := filepath.Join(root, "escape")
	if err := os.Symlink(filepath.Join("..", "outside"), link); err != nil {
		t.Skipf("symlink creation is unavailable: %v", err)
	}
	if err := verifyNoEscapingSymlinks(root); err == nil || !strings.Contains(err.Error(), "escapes") {
		t.Fatalf("expected escaping symlink rejection, got %v", err)
	}
}
