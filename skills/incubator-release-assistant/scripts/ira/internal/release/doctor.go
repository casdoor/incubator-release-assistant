package release

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const (
	WorkspaceReference     = "references/workspace-bootstrap.md"
	ConfigurationReference = "references/configuration.md"
	SigningKeyReference    = "references/signing-key-setup.md"
	OfficialKeysReference  = "references/asf-keys-publication.md"
	PrerequisitesReference = "references/prerequisites.md"
	RecoveryReference      = "references/release-recovery.md"
)

type DoctorPaths struct {
	ReleaseConfig   string `json:"releaseConfig"`
	SecretDirectory string `json:"secretDirectory"`
	PublicKeyExport string `json:"publicKeyExport"`
	KeyMetadata     string `json:"keyMetadata"`
	EvidenceRoot    string `json:"evidenceRoot"`
}

type DoctorReport struct {
	Status     string      `json:"status"`
	Code       string      `json:"code"`
	Summary    string      `json:"summary"`
	Paths      DoctorPaths `json:"paths"`
	Missing    []string    `json:"missing,omitempty"`
	Reference  string      `json:"reference,omitempty"`
	NextAction string      `json:"nextAction"`
}

type ErrorGuidance struct {
	Code       string
	Reference  string
	NextAction string
}

func (e Engine) Doctor(configPath string) DoctorReport {
	paths := doctorPaths(configPath)
	blocked := func(code, summary, reference, next string, missing ...string) DoctorReport {
		return DoctorReport{
			Status:     "blocked",
			Code:       code,
			Summary:    summary,
			Paths:      paths,
			Missing:    missing,
			Reference:  reference,
			NextAction: next,
		}
	}

	repositoryRoot := doctorRepositoryRoot(paths.ReleaseConfig)
	inside, err := pathInsideOrEqual(paths.SecretDirectory, repositoryRoot)
	if err != nil {
		return blocked("IRA-WORKSPACE-001", err.Error(), WorkspaceReference, "choose_external_secret_directory")
	}
	if inside {
		return blocked(
			"IRA-WORKSPACE-001",
			"The signing home must be outside the release-assistant repository.",
			WorkspaceReference,
			"choose_external_secret_directory",
			"external secret directory",
		)
	}
	if gitRoot := containingGitRoot(paths.SecretDirectory); gitRoot != "" {
		return blocked(
			"IRA-WORKSPACE-001",
			fmt.Sprintf("The signing home is inside a Git worktree rooted at %s.", gitRoot),
			WorkspaceReference,
			"choose_plain_parent_workspace",
			"plain parent workspace",
		)
	}

	cfg, err := loadConfig(paths.ReleaseConfig)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return blocked(
				"IRA-CONFIG-001",
				"The release configuration does not exist yet.",
				WorkspaceReference,
				"create_config_template",
				"release config",
			)
		}
		return blocked("IRA-CONFIG-002", err.Error(), ConfigurationReference, "repair_config_json", "valid release config")
	}

	var publicInputs []string
	if strings.TrimSpace(cfg.Signing.ApacheID) == "" || strings.Contains(cfg.Signing.ApacheID, "replace-with") {
		publicInputs = append(publicInputs, "signing.apacheId")
	}
	if !hex40Pattern.MatchString(cfg.Source.Commit) || strings.EqualFold(cfg.Source.Commit, strings.Repeat("0", 40)) {
		publicInputs = append(publicInputs, "source.commit")
	}
	if len(publicInputs) > 0 {
		return blocked(
			"IRA-CONFIG-003",
			"The non-secret release inputs are incomplete.",
			WorkspaceReference,
			"collect_release_inputs",
			publicInputs...,
		)
	}
	if !hex40Pattern.MatchString(cfg.Signing.Fingerprint) || strings.EqualFold(cfg.Signing.Fingerprint, strings.Repeat("0", 40)) {
		return blocked(
			"IRA-KEY-001",
			"A signing-key fingerprint must be obtained from an existing or newly generated key before config validation can finish.",
			SigningKeyReference,
			"inspect_or_create_signing_key",
			"signing.fingerprint",
		)
	}
	if err := cfg.Validate(); err != nil {
		return blocked("IRA-CONFIG-002", err.Error(), ConfigurationReference, "repair_release_config", "release-ready config")
	}

	info, err := os.Stat(paths.SecretDirectory)
	if err != nil {
		if os.IsNotExist(err) {
			return blocked(
				"IRA-KEY-001",
				"The external GPG home does not exist yet.",
				SigningKeyReference,
				"inspect_or_create_signing_key",
				"external GPG home",
			)
		}
		return blocked("IRA-KEY-001", err.Error(), SigningKeyReference, "inspect_signing_key", "readable external GPG home")
	}
	if !info.IsDir() {
		return blocked("IRA-KEY-001", "The configured signing home is not a directory.", SigningKeyReference, "choose_external_gpg_home", "external GPG home")
	}
	if !hasGPGHomeMaterial(paths.SecretDirectory) {
		return blocked(
			"IRA-KEY-001",
			"The external GPG home exists but does not contain a signing keyring.",
			SigningKeyReference,
			"inspect_or_create_signing_key",
			"compliant signing private key",
		)
	}

	for _, command := range []string{"git", "tar", "java", "svn", "gpg", "go"} {
		if err := commandExists(command); err != nil {
			return blocked("IRA-DEPENDENCY-001", err.Error(), PrerequisitesReference, "install_missing_dependency", command)
		}
	}
	if err := e.inspectSigningKey(cfg, paths.SecretDirectory); err != nil {
		return blocked("IRA-KEY-001", err.Error(), SigningKeyReference, "inspect_or_create_signing_key", "compliant signing private key")
	}
	if err := e.doctorOfficialKeys(cfg); err != nil {
		if strings.Contains(strings.ToLower(err.Error()), "not present in the official keys") {
			return blocked("IRA-KEYS-001", err.Error(), OfficialKeysReference, "prepare_official_keys_update", "fingerprint in official KEYS")
		}
		return blocked("IRA-PREFLIGHT-001", err.Error(), PrerequisitesReference, "retry_public_keys_check", "public KEYS check")
	}

	return DoctorReport{
		Status:     "ready",
		Code:       "IRA-READY",
		Summary:    "Configuration, toolchain, signing key, and official KEYS checks passed.",
		Paths:      paths,
		NextAction: "run_plan",
	}
}

func hasGPGHomeMaterial(home string) bool {
	for _, name := range []string{"pubring.kbx", "pubring.gpg", "private-keys-v1.d"} {
		if _, err := os.Stat(filepath.Join(home, name)); err == nil {
			return true
		}
	}
	return false
}

func doctorPaths(configPath string) DoctorPaths {
	if strings.TrimSpace(configPath) == "" {
		configPath = filepath.Join("config", "local", "casbin.local.json")
	}
	configAbs, err := filepath.Abs(configPath)
	if err != nil {
		configAbs = filepath.Clean(configPath)
	}
	repositoryRoot := doctorRepositoryRoot(configAbs)
	workspace := filepath.Dir(repositoryRoot)
	secret := strings.TrimSpace(os.Getenv(SecretDirectoryEnvironment))
	if secret == "" {
		secret = filepath.Join(workspace, "secretkey")
	}
	if abs, err := filepath.Abs(secret); err == nil {
		secret = abs
	}
	publicRoot := filepath.Join(workspace, "public-key")
	return DoctorPaths{
		ReleaseConfig:   configAbs,
		SecretDirectory: filepath.Clean(secret),
		PublicKeyExport: filepath.Join(publicRoot, "apache-casbin-release-key.asc"),
		KeyMetadata:     filepath.Join(publicRoot, "key-metadata.json"),
		EvidenceRoot:    filepath.Join(repositoryRoot, ".ira", "runs"),
	}
}

func doctorRepositoryRoot(configPath string) string {
	if raw := strings.TrimSpace(os.Getenv(RepositoryRootEnvironment)); raw != "" {
		if abs, err := filepath.Abs(raw); err == nil {
			return abs
		}
		return filepath.Clean(raw)
	}
	return findWorkspaceRoot(filepath.Dir(configPath))
}

func (e Engine) doctorOfficialKeys(cfg *Config) error {
	tempRoot, err := os.MkdirTemp("", "ira-doctor-keys-")
	if err != nil {
		return fmt.Errorf("create temporary public keyring: %w", err)
	}
	defer os.RemoveAll(tempRoot)

	keyring := filepath.Join(tempRoot, "keyring")
	if err := os.MkdirAll(keyring, 0o700); err != nil {
		return err
	}
	keys := filepath.Join(tempRoot, "KEYS")
	if _, err := e.Runner.Output("", "svn", "export", "--force", cfg.Signing.KeysURL, keys); err != nil {
		return err
	}
	if _, err := e.Runner.Output("", "gpg", "--homedir", keyring, "--batch", "--import", keys); err != nil {
		return err
	}
	out, err := e.Runner.Output("", "gpg", "--homedir", keyring, "--batch", "--with-colons", "--fingerprint")
	if err != nil {
		return err
	}
	wanted := strings.ToUpper(cfg.Signing.Fingerprint)
	for _, line := range strings.Split(out, "\n") {
		fields := strings.Split(line, ":")
		if len(fields) > 9 && fields[0] == "fpr" && strings.ToUpper(fields[9]) == wanted {
			return nil
		}
	}
	return fmt.Errorf("configured fingerprint is not present in the official KEYS file")
}

func GuidanceForError(err error) ErrorGuidance {
	message := strings.ToLower(err.Error())
	switch {
	case strings.Contains(message, "secret directory") && (strings.Contains(message, "git worktree") || strings.Contains(message, "outside")):
		return ErrorGuidance{"IRA-WORKSPACE-001", WorkspaceReference, "run_doctor"}
	case strings.Contains(message, "official keys") || strings.Contains(message, "official casbin keys"):
		return ErrorGuidance{"IRA-KEYS-001", OfficialKeysReference, "prepare_official_keys_update"}
	case strings.Contains(message, "signing private key") || strings.Contains(message, "signing key") || strings.Contains(message, "apache.org uid"):
		return ErrorGuidance{"IRA-KEY-001", SigningKeyReference, "run_doctor"}
	case strings.Contains(message, "required command"):
		return ErrorGuidance{"IRA-DEPENDENCY-001", PrerequisitesReference, "install_missing_dependency"}
	case strings.Contains(message, "existing state") || strings.Contains(message, "frozen candidate") || strings.Contains(message, "--clean"):
		return ErrorGuidance{"IRA-RECOVERY-001", RecoveryReference, "inspect_release_state"}
	case strings.Contains(message, "--config is required") || strings.Contains(message, "read config"):
		return ErrorGuidance{"IRA-CONFIG-001", WorkspaceReference, "run_doctor"}
	case strings.Contains(message, "configuration") || strings.Contains(message, "config"):
		return ErrorGuidance{"IRA-CONFIG-002", ConfigurationReference, "run_doctor"}
	default:
		return ErrorGuidance{"IRA-EXEC-001", RecoveryReference, "inspect_failure_evidence"}
	}
}
