package release

import (
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func doctorFixture(t *testing.T) (workspace, repository, configPath, secret string) {
	t.Helper()
	workspace = t.TempDir()
	repository = filepath.Join(workspace, "Incubator-release-assistant")
	configPath = filepath.Join(repository, "config", "local", "casbin.local.json")
	secret = filepath.Join(workspace, "secretkey")
	if err := os.MkdirAll(filepath.Dir(configPath), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(repository, "ira.ps1"), []byte("fixture\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	t.Setenv(RepositoryRootEnvironment, repository)
	t.Setenv(SecretDirectoryEnvironment, secret)
	return workspace, repository, configPath, secret
}

func writeDoctorConfig(t *testing.T, path string, cfg *Config) {
	t.Helper()
	raw, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, raw, 0o600); err != nil {
		t.Fatal(err)
	}
}

func TestDoctorExplainsMissingConfigWithResolvedPaths(t *testing.T) {
	_, _, configPath, secret := doctorFixture(t)
	report := (Engine{}).Doctor(configPath)
	if report.Code != "IRA-CONFIG-001" || report.NextAction != "create_config_template" {
		t.Fatalf("unexpected report: %+v", report)
	}
	if report.Paths.ReleaseConfig != configPath || report.Paths.SecretDirectory != secret {
		t.Fatalf("doctor did not resolve expected paths: %+v", report.Paths)
	}
	if report.Reference != WorkspaceReference {
		t.Fatalf("unexpected reference: %s", report.Reference)
	}
}

func TestDoctorCollectsPublicInputsBeforeRequestingKey(t *testing.T) {
	_, _, configPath, _ := doctorFixture(t)
	cfg := validConfig(t)
	cfg.Signing.ApacheID = "replace-with-apache-id"
	cfg.Source.Commit = "0000000000000000000000000000000000000000"
	cfg.Signing.Fingerprint = "0000000000000000000000000000000000000000"
	writeDoctorConfig(t, configPath, cfg)

	report := (Engine{}).Doctor(configPath)
	if report.Code != "IRA-CONFIG-003" || report.NextAction != "collect_release_inputs" {
		t.Fatalf("unexpected report: %+v", report)
	}
	want := []string{"signing.apacheId", "source.commit"}
	if !reflect.DeepEqual(report.Missing, want) {
		t.Fatalf("unexpected missing inputs: %#v", report.Missing)
	}
}

func TestDoctorRoutesPlaceholderFingerprintToSigningKeySetup(t *testing.T) {
	_, _, configPath, _ := doctorFixture(t)
	cfg := validConfig(t)
	cfg.Signing.Fingerprint = "0000000000000000000000000000000000000000"
	writeDoctorConfig(t, configPath, cfg)

	report := (Engine{}).Doctor(configPath)
	if report.Code != "IRA-KEY-001" || report.Reference != SigningKeyReference || report.NextAction != "inspect_or_create_signing_key" {
		t.Fatalf("unexpected report: %+v", report)
	}
}

func TestDoctorDoesNotInitializeAnEmptyGPGHome(t *testing.T) {
	_, _, configPath, secret := doctorFixture(t)
	cfg := validConfig(t)
	writeDoctorConfig(t, configPath, cfg)
	if err := os.Mkdir(secret, 0o700); err != nil {
		t.Fatal(err)
	}

	report := (Engine{}).Doctor(configPath)
	if report.Code != "IRA-KEY-001" || report.NextAction != "inspect_or_create_signing_key" {
		t.Fatalf("unexpected report: %+v", report)
	}
	entries, err := os.ReadDir(secret)
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 0 {
		t.Fatalf("doctor initialized the empty GPG home: %#v", entries)
	}
}

func TestDoctorRejectsSecretDirectoryInsideGitWorkspace(t *testing.T) {
	workspace, _, configPath, _ := doctorFixture(t)
	if err := os.Mkdir(filepath.Join(workspace, ".git"), 0o700); err != nil {
		t.Fatal(err)
	}
	report := (Engine{}).Doctor(configPath)
	if report.Code != "IRA-WORKSPACE-001" || report.Reference != WorkspaceReference {
		t.Fatalf("unexpected report: %+v", report)
	}
}

func TestGuidanceForKnownSetupFailures(t *testing.T) {
	cases := []struct {
		message string
		code    string
	}{
		{"configured signing private key was not found", "IRA-KEY-001"},
		{"configured fingerprint is not present in the official KEYS file", "IRA-KEYS-001"},
		{"required command \"gpg\" was not found", "IRA-DEPENDENCY-001"},
		{"--config is required", "IRA-CONFIG-001"},
		{"existing state does not match this configuration", "IRA-RECOVERY-001"},
	}
	for _, tc := range cases {
		if got := GuidanceForError(assertionError(tc.message)); got.Code != tc.code {
			t.Errorf("%q: got %s, want %s", tc.message, got.Code, tc.code)
		}
	}
}

func TestPathOnDifferentWindowsVolumeIsExternal(t *testing.T) {
	if filepath.VolumeName(`C:\\workspace\\secretkey`) == "" {
		t.Skip("volume semantics are Windows-specific")
	}
	inside, err := pathInsideOrEqual(`C:\\workspace\\secretkey`, `D:\\repository`)
	if err != nil || inside {
		t.Fatalf("different-volume path should be external: inside=%v err=%v", inside, err)
	}
}

type assertionError string

func (e assertionError) Error() string { return string(e) }
