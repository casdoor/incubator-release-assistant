package release

import (
	"archive/zip"
	"crypto/sha512"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

type Engine struct {
	Runner Runner
}

func (e Engine) Plan(cfg *Config) string {
	return fmt.Sprintf(`Release plan
  Project:       %s (%s)
  Adapter:       %s
  Repository:    %s
  Commit:        %s
  Candidate:     %s
  Artifact:      %s
  Test sandbox:  %s image %s (network: %s)
  State:         %s

Trust boundaries
  1. prepare: clone/archive/RAT; run Go tests only inside the container, without host credentials or artifact access
  2. sign: verify the prepared SHA-512, then require the exact digest as human confirmation before using GPG
  3. stage: require "STAGE RC%d", refuse an existing remote RC, upload, then re-download and verify public bytes
`, cfg.Project.DisplayName, cfg.Project.ID, cfg.Project.Adapter, cfg.Source.Repository,
		stringsLower(cfg.Source.Commit), cfg.RunID(), cfg.ArtifactName(),
		cfg.Runtime.Container.Engine, cfg.Runtime.Container.Image, cfg.Runtime.Container.Network,
		cfg.Runtime.StateDirectory, cfg.Release.RC)
}

func (e Engine) Prepare(cfg *Config, clean bool) (*State, error) {
	for _, command := range []string{"git", "tar", "java", "svn", cfg.Runtime.Container.Engine} {
		if err := commandExists(command); err != nil {
			return nil, err
		}
	}
	runRoot, err := cfg.RunRoot()
	if err != nil {
		return nil, err
	}
	if clean {
		if old, err := LoadState(runRoot); err == nil && old.Staged {
			return nil, fmt.Errorf("refusing --clean because this run is recorded as staged; use a new RC number")
		}
		if err := removeContained(runRoot, filepath.Dir(runRoot)); err != nil {
			return nil, err
		}
	}
	if state, err := LoadState(runRoot); err == nil {
		if err := state.VerifyConfig(cfg); err != nil {
			return nil, err
		}
		if state.Prepared {
			if err := verifyPreparedFiles(cfg, runRoot, state); err != nil {
				return nil, fmt.Errorf("prepared state failed revalidation: %w", err)
			}
			fmt.Fprintln(e.out(), "Prepared artifacts already exist and were revalidated; no project code was rerun.")
			return state, nil
		}
		return nil, fmt.Errorf("an incomplete prepare run exists; inspect its evidence, then rerun with --clean if it is safe to discard")
	} else if _, statErr := os.Stat(runRoot); statErr == nil {
		return nil, fmt.Errorf("run directory exists without readable state; inspect it, then rerun with --clean if it is safe to discard")
	}

	work := filepath.Join(runRoot, "work")
	artifacts := filepath.Join(runRoot, "artifacts")
	evidence := filepath.Join(runRoot, "evidence")
	for _, dir := range []string{work, artifacts, evidence} {
		if err := os.MkdirAll(dir, 0o700); err != nil {
			return nil, err
		}
	}
	state := NewState(cfg)
	if err := state.Save(runRoot); err != nil {
		return nil, err
	}

	sourceRepo := filepath.Join(work, "source-repository")
	if err := e.Runner.Run(filepath.Join(evidence, "git-clone.log"), "", "git", "clone", "--no-checkout", "--no-tags", cfg.Source.Repository, sourceRepo); err != nil {
		return nil, err
	}
	resolved, err := e.Runner.Output("", "git", "-C", sourceRepo, "rev-parse", cfg.Source.Commit+"^{commit}")
	if err != nil {
		return nil, err
	}
	if stringsLower(resolved) != stringsLower(cfg.Source.Commit) {
		return nil, fmt.Errorf("resolved commit %s differs from configured commit", resolved)
	}

	artifactPath := filepath.Join(artifacts, cfg.ArtifactName())
	if err := e.Runner.Run(filepath.Join(evidence, "git-archive.log"), "", "git", "-C", sourceRepo, "archive", "--format=tar.gz", "--prefix="+cfg.Source.ArchivePrefix+"/", "--output="+artifactPath, cfg.Source.Commit); err != nil {
		return nil, err
	}
	extractRoot := filepath.Join(work, "extracted")
	if err := os.MkdirAll(extractRoot, 0o700); err != nil {
		return nil, err
	}
	if err := e.Runner.Run(filepath.Join(evidence, "extract.log"), "", "tar", "-xzf", artifactPath, "-C", extractRoot); err != nil {
		return nil, err
	}
	extracted := filepath.Join(extractRoot, cfg.Source.ArchivePrefix)
	if err := verifyArchiveLayout(extractRoot, extracted); err != nil {
		return nil, err
	}
	if err := verifyNoEscapingSymlinks(extracted); err != nil {
		return nil, err
	}
	for _, required := range cfg.Checks.RequiredFiles {
		if info, err := os.Stat(filepath.Join(extracted, filepath.FromSlash(required))); err != nil || info.IsDir() {
			return nil, fmt.Errorf("source archive is missing required file %s", required)
		}
	}
	if err := e.runRAT(cfg, runRoot, extracted); err != nil {
		return nil, err
	}
	if err := e.runCasbinGoTests(cfg, runRoot, extracted); err != nil {
		return nil, err
	}

	digest, err := sha512File(artifactPath)
	if err != nil {
		return nil, err
	}
	checksumPath := artifactPath + ".sha512"
	if err := os.WriteFile(checksumPath, []byte(digest+"  "+cfg.ArtifactName()+"\n"), 0o600); err != nil {
		return nil, err
	}
	if parsed, err := readChecksum(checksumPath, cfg.ArtifactName()); err != nil || parsed != digest {
		return nil, fmt.Errorf("generated checksum did not pass independent validation: %w", err)
	}
	checksumDigest, err := sha512File(checksumPath)
	if err != nil {
		return nil, err
	}
	state.ArtifactSHA512 = digest
	state.ChecksumSHA512 = checksumDigest
	state.Prepared = true
	if err := state.Save(runRoot); err != nil {
		return nil, err
	}
	if err := writeText(filepath.Join(evidence, "release-manifest.txt"), fmt.Sprintf("Repository: %s\nCommit: %s\nArtifact: %s\nArtifact SHA512: %s\nChecksum file SHA512: %s\n", cfg.Source.Repository, stringsLower(cfg.Source.Commit), cfg.ArtifactName(), digest, checksumDigest)); err != nil {
		return nil, err
	}
	fmt.Fprintf(e.out(), "Prepared and verified %s\nSHA-512: %s\nNext: ira sign --config %q --confirm %s\n", artifactPath, digest, cfg.Path, digest)
	return state, nil
}

func (e Engine) runCasbinGoTests(cfg *Config, runRoot, extracted string) error {
	evidence := filepath.Join(runRoot, "evidence", "go-test.log")
	args := casbinGoTestArgs(cfg, extracted)
	fmt.Fprintln(e.out(), "Running repository code in an isolated container. The artifact directory and host credentials are not mounted.")
	if err := e.Runner.Run(evidence, "", cfg.Runtime.Container.Engine, args...); err != nil {
		return err
	}
	imageID, err := e.Runner.Output("", cfg.Runtime.Container.Engine, "image", "inspect", "--format={{.Id}}", cfg.Runtime.Container.Image)
	if err != nil {
		return fmt.Errorf("tests passed but container image identity could not be recorded: %w", err)
	}
	return writeText(filepath.Join(runRoot, "evidence", "container-image.txt"), cfg.Runtime.Container.Image+"\n"+imageID+"\n")
}

func casbinGoTestArgs(cfg *Config, extracted string) []string {
	args := []string{
		"run", "--rm", "--read-only", "--cap-drop=ALL",
		"--security-opt=no-new-privileges", "--pids-limit=512",
		"--tmpfs", "/tmp:rw,exec,nosuid,nodev,size=2g",
		"--tmpfs", "/work:rw,nosuid,size=2g",
		"--mount", "type=bind,src=" + extracted + ",dst=/input,readonly",
		"--workdir", "/work", "--env", "GOCACHE=/tmp/go-build",
		"--env", "GOMODCACHE=/tmp/go-mod",
	}
	if cfg.Runtime.Container.Network == "none" {
		args = append(args, "--network", "none")
	}
	args = append(args, cfg.Runtime.Container.Image, "sh", "-c", "cp -a /input/. /work/ && go test ./...")
	return args
}

func (e Engine) runRAT(cfg *Config, runRoot, extracted string) error {
	tools := filepath.Join(runRoot, "tools")
	evidence := filepath.Join(runRoot, "evidence")
	if err := os.MkdirAll(tools, 0o700); err != nil {
		return err
	}
	base := "apache-rat-" + cfg.Checks.RAT.Version + "-bin.zip"
	baseURL := "https://dist.apache.org/repos/dist/release/creadur/apache-rat-" + cfg.Checks.RAT.Version
	zipPath := filepath.Join(tools, base)
	checksumPath := zipPath + ".sha512"
	if _, err := os.Stat(zipPath); os.IsNotExist(err) {
		if err := e.Runner.Run(filepath.Join(evidence, "rat-download.log"), "", "svn", "export", "--force", baseURL+"/"+base, zipPath); err != nil {
			return err
		}
	}
	if err := e.Runner.Run(filepath.Join(evidence, "rat-checksum-download.log"), "", "svn", "export", "--force", baseURL+"/"+base+".sha512", checksumPath); err != nil {
		return err
	}
	raw, err := os.ReadFile(checksumPath)
	if err != nil {
		return err
	}
	fields := strings.Fields(string(raw))
	if len(fields) == 0 {
		return fmt.Errorf("Apache RAT checksum file is empty")
	}
	actual, err := sha512File(zipPath)
	if err != nil {
		return err
	}
	if !strings.EqualFold(fields[0], actual) {
		return fmt.Errorf("Apache RAT download SHA-512 mismatch")
	}
	extract := filepath.Join(tools, "rat")
	jar, err := extractRATJar(zipPath, extract, cfg.Checks.RAT.Version)
	if err != nil {
		return err
	}
	report := filepath.Join(evidence, "rat-report.txt")
	if err := e.Runner.Run(filepath.Join(evidence, "rat-console.log"), "", "java", "-jar", jar, "--output-file", report, "--input-exclude-file", filepath.Join(extracted, filepath.FromSlash(cfg.Checks.RAT.ExcludeFile)), extracted); err != nil {
		return err
	}
	text, err := os.ReadFile(report)
	if err != nil {
		return err
	}
	unapproved, err := ratCount(text, "Unapproved")
	if err != nil {
		return err
	}
	unknown, err := ratCount(text, "Unknown")
	if err != nil {
		return err
	}
	if unapproved != 0 || unknown != 0 {
		return fmt.Errorf("Apache RAT did not report Unapproved: 0 and Unknown: 0; inspect %s", report)
	}
	return nil
}

func ratCount(report []byte, label string) (int, error) {
	pattern := regexp.MustCompile(`(?m)^\s*` + regexp.QuoteMeta(label) + `:\s*([0-9]+)\b`)
	match := pattern.FindSubmatch(report)
	if len(match) != 2 {
		return 0, fmt.Errorf("could not read %s count from Apache RAT report", label)
	}
	value, err := strconv.Atoi(string(match[1]))
	if err != nil {
		return 0, fmt.Errorf("parse Apache RAT %s count: %w", label, err)
	}
	return value, nil
}

func (e Engine) Sign(cfg *Config, confirmation string) (*State, error) {
	runRoot, _ := cfg.RunRoot()
	state, err := LoadState(runRoot)
	if err != nil {
		return nil, fmt.Errorf("load prepared state: %w", err)
	}
	if err := state.VerifyConfig(cfg); err != nil {
		return nil, err
	}
	if !state.Prepared {
		return nil, fmt.Errorf("candidate is not prepared")
	}
	if err := verifyPreparedFiles(cfg, runRoot, state); err != nil {
		return nil, err
	}
	if confirmation != state.ArtifactSHA512 {
		return nil, fmt.Errorf("signing confirmation must exactly equal the prepared artifact SHA-512")
	}
	for _, command := range []string{"gpg", "svn"} {
		if err := commandExists(command); err != nil {
			return nil, err
		}
	}
	if state.Signed {
		if err := e.verifyReleaseFiles(cfg, runRoot, state); err != nil {
			return nil, fmt.Errorf("signed state failed exact-byte revalidation: %w", err)
		}
		fmt.Fprintln(e.out(), "Signature already exists; all three candidate files were revalidated byte-for-byte and no new signature was created.")
		return state, nil
	}
	if err := e.verifySigningKey(cfg, runRoot); err != nil {
		return nil, err
	}
	artifact := filepath.Join(runRoot, "artifacts", cfg.ArtifactName())
	signature := artifact + ".asc"
	if err := e.Runner.RunInteractive(filepath.Join(runRoot, "evidence", "gpg-sign.log"), "", "gpg", "--armor", "--local-user", strings.ToUpper(cfg.Signing.Fingerprint), "--output", signature, "--detach-sign", artifact); err != nil {
		return nil, err
	}
	if err := e.verifySignatureWithOfficialKeys(cfg, runRoot, signature, artifact); err != nil {
		return nil, err
	}
	signatureDigest, err := sha512File(signature)
	if err != nil {
		return nil, err
	}
	state.Signed = true
	state.Signer = strings.ToUpper(cfg.Signing.Fingerprint)
	state.SignatureSHA512 = signatureDigest
	if err := state.Save(runRoot); err != nil {
		return nil, err
	}
	if err := writeText(filepath.Join(runRoot, "evidence", "candidate-manifest.txt"), fmt.Sprintf("Artifact: %s\nArtifact SHA512: %s\nChecksum file: %s.sha512\nChecksum file SHA512: %s\nSignature file: %s.asc\nSignature file SHA512: %s\nSigner primary fingerprint: %s\n", cfg.ArtifactName(), state.ArtifactSHA512, cfg.ArtifactName(), state.ChecksumSHA512, cfg.ArtifactName(), state.SignatureSHA512, state.Signer)); err != nil {
		return nil, err
	}
	fmt.Fprintf(e.out(), "Signature created and verified with the official KEYS file.\nNext: ira stage --config %q --confirm \"STAGE RC%d\"\n", cfg.Path, cfg.Release.RC)
	return state, nil
}

func (e Engine) Stage(cfg *Config, confirmation string) (*State, error) {
	runRoot, _ := cfg.RunRoot()
	state, err := LoadState(runRoot)
	if err != nil {
		return nil, err
	}
	if err := state.VerifyConfig(cfg); err != nil {
		return nil, err
	}
	if !state.Signed {
		return nil, fmt.Errorf("candidate must be signed before staging")
	}
	if err := e.verifyReleaseFiles(cfg, runRoot, state); err != nil {
		return nil, err
	}
	expectedConfirmation := fmt.Sprintf("STAGE RC%d", cfg.Release.RC)
	if confirmation != expectedConfirmation {
		return nil, fmt.Errorf("staging confirmation must exactly equal %q", expectedConfirmation)
	}
	if err := commandExists("svn"); err != nil {
		return nil, err
	}
	if state.Staged {
		if err := e.VerifyPublic(cfg); err != nil {
			return nil, err
		}
		return LoadState(runRoot)
	}
	entries, err := e.Runner.Output("", "svn", "list", cfg.Distribution.DevURL)
	if err != nil {
		return nil, err
	}
	remoteDir := fmt.Sprintf("%s-rc%d", cfg.Release.Version, cfg.Release.RC)
	remoteExists := false
	for _, entry := range strings.Split(entries, "\n") {
		if strings.TrimSuffix(strings.TrimSpace(entry), "/") == remoteDir {
			remoteExists = true
		}
	}
	if remoteExists {
		state.Staged = true
		state.PublicURL = cfg.Distribution.DevURL + "/" + remoteDir + "/"
		if err := state.Save(runRoot); err != nil {
			return nil, err
		}
		if err := e.VerifyPublic(cfg); err != nil {
			state.Staged = false
			state.PublicURL = ""
			_ = state.Save(runRoot)
			return nil, fmt.Errorf("remote RC directory exists but does not exactly match this prepared candidate; never overwrite it, increment release.rc: %w", err)
		}
		fmt.Fprintln(e.out(), "Recovered an earlier successful SVN commit by verifying the existing remote RC byte-for-byte.")
		return LoadState(runRoot)
	}
	workingCopy := filepath.Join(runRoot, "work", "dist-dev")
	if err := removeContained(workingCopy, filepath.Join(runRoot, "work")); err != nil {
		return nil, err
	}
	if err := e.Runner.Run(filepath.Join(runRoot, "evidence", "svn-checkout.log"), "", "svn", "checkout", "--depth", "immediates", cfg.Distribution.DevURL, workingCopy); err != nil {
		return nil, err
	}
	stageDir := filepath.Join(workingCopy, remoteDir)
	if err := os.MkdirAll(stageDir, 0o700); err != nil {
		return nil, err
	}
	for _, name := range releaseFileNames(cfg) {
		if err := copyFile(filepath.Join(runRoot, "artifacts", name), filepath.Join(stageDir, name)); err != nil {
			return nil, err
		}
	}
	if err := e.Runner.Run(filepath.Join(runRoot, "evidence", "svn-add.log"), "", "svn", "add", stageDir); err != nil {
		return nil, err
	}
	status, err := e.Runner.Output("", "svn", "status", stageDir)
	if err != nil {
		return nil, err
	}
	for _, line := range strings.Split(status, "\n") {
		if strings.TrimSpace(line) != "" && !strings.HasPrefix(line, "A") {
			return nil, fmt.Errorf("unexpected svn status before commit: %s", line)
		}
	}
	if err := writeText(filepath.Join(runRoot, "evidence", "svn-status-before-commit.txt"), status+"\n"); err != nil {
		return nil, err
	}
	message := fmt.Sprintf("[casbin] Stage Apache Casbin %s RC%d", cfg.Release.Version, cfg.Release.RC)
	if err := e.Runner.RunInteractive(filepath.Join(runRoot, "evidence", "svn-commit.log"), "", "svn", "commit", stageDir, "--username", cfg.Signing.ApacheID, "--no-auth-cache", "-m", message); err != nil {
		return nil, err
	}
	state.Staged = true
	state.PublicURL = cfg.Distribution.DevURL + "/" + remoteDir + "/"
	if err := state.Save(runRoot); err != nil {
		return nil, err
	}
	if err := e.VerifyPublic(cfg); err != nil {
		return nil, fmt.Errorf("staged, but public verification is pending: %w", err)
	}
	return LoadState(runRoot)
}

func (e Engine) VerifyPublic(cfg *Config) error {
	runRoot, _ := cfg.RunRoot()
	state, err := LoadState(runRoot)
	if err != nil {
		return err
	}
	if !state.Staged || state.PublicURL == "" {
		return fmt.Errorf("candidate has not been staged")
	}
	if state.PublicVerified {
		state.PublicVerified = false
		if err := state.Save(runRoot); err != nil {
			return err
		}
	}
	publicDir := filepath.Join(runRoot, "work", "public-download")
	if err := removeContained(publicDir, filepath.Join(runRoot, "work")); err != nil {
		return err
	}
	if err := os.MkdirAll(publicDir, 0o700); err != nil {
		return err
	}
	for _, name := range releaseFileNames(cfg) {
		if err := e.Runner.Run(filepath.Join(runRoot, "evidence", "public-download-"+safeLogName(name)+".log"), "", "svn", "export", "--force", state.PublicURL+name, filepath.Join(publicDir, name)); err != nil {
			return err
		}
	}
	expectedDigests := map[string]string{
		cfg.ArtifactName():             state.ArtifactSHA512,
		cfg.ArtifactName() + ".sha512": state.ChecksumSHA512,
		cfg.ArtifactName() + ".asc":    state.SignatureSHA512,
	}
	for name, expected := range expectedDigests {
		if err := verifyExactFileDigest(filepath.Join(publicDir, name), expected, "public "+name); err != nil {
			return err
		}
	}
	downloadedArtifact := filepath.Join(publicDir, cfg.ArtifactName())
	downloadedDigest, err := sha512File(downloadedArtifact)
	if err != nil {
		return err
	}
	if downloadedDigest != state.ArtifactSHA512 {
		return fmt.Errorf("public artifact bytes differ from the prepared artifact")
	}
	checksumDigest, err := readChecksum(downloadedArtifact+".sha512", cfg.ArtifactName())
	if err != nil || checksumDigest != downloadedDigest {
		return fmt.Errorf("public checksum verification failed: %w", err)
	}
	if err := e.verifySignatureWithOfficialKeys(cfg, runRoot, downloadedArtifact+".asc", downloadedArtifact); err != nil {
		return err
	}
	state.PublicVerified = true
	if err := state.Save(runRoot); err != nil {
		return err
	}
	fmt.Fprintf(e.out(), "Public candidate bytes, checksum, and signature verified: %s\n", state.PublicURL)
	return nil
}

func (e Engine) verifySigningKey(cfg *Config, runRoot string) error {
	out, err := e.Runner.Output("", "gpg", "--batch", "--with-colons", "--list-secret-keys", strings.ToUpper(cfg.Signing.Fingerprint))
	if err != nil {
		return fmt.Errorf("configured signing private key is unavailable: %w", err)
	}
	foundSec := false
	foundApacheUID := false
	for _, line := range strings.Split(out, "\n") {
		fields := strings.Split(line, ":")
		if len(fields) < 12 {
			continue
		}
		if fields[0] == "sec" {
			foundSec = true
			bits, _ := strconv.Atoi(fields[2])
			algorithm, _ := strconv.Atoi(fields[3])
			if bits < cfg.Signing.MinimumRSABits || (algorithm != 1 && algorithm != 2 && algorithm != 3) {
				return fmt.Errorf("signing key must be RSA with at least %d bits", cfg.Signing.MinimumRSABits)
			}
			if strings.ContainsAny(fields[1], "re") || !strings.Contains(fields[11], "s") {
				return fmt.Errorf("signing key is revoked, expired, or cannot sign")
			}
		}
		if fields[0] == "uid" && strings.Contains(strings.ToLower(fields[9]), "@apache.org") {
			foundApacheUID = true
		}
	}
	if !foundSec {
		return fmt.Errorf("configured signing private key was not found")
	}
	if cfg.Signing.RequireApacheUID && !foundApacheUID {
		return fmt.Errorf("project policy requires an apache.org UID on the signing key")
	}
	return e.ensureOfficialKeyring(cfg, runRoot)
}

func (e Engine) ensureOfficialKeyring(cfg *Config, runRoot string) error {
	keyring := filepath.Join(runRoot, "work", "official-keyring")
	if err := os.MkdirAll(keyring, 0o700); err != nil {
		return err
	}
	keys := filepath.Join(runRoot, "work", "KEYS")
	if err := e.Runner.Run(filepath.Join(runRoot, "evidence", "keys-download.log"), "", "svn", "export", "--force", cfg.Signing.KeysURL, keys); err != nil {
		return err
	}
	if err := e.Runner.Run(filepath.Join(runRoot, "evidence", "keys-import.log"), "", "gpg", "--homedir", keyring, "--batch", "--import", keys); err != nil {
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

func (e Engine) verifySignatureWithOfficialKeys(cfg *Config, runRoot, signature, artifact string) error {
	if err := e.ensureOfficialKeyring(cfg, runRoot); err != nil {
		return err
	}
	keyring := filepath.Join(runRoot, "work", "official-keyring")
	return e.Runner.Run(filepath.Join(runRoot, "evidence", "gpg-verify.log"), "", "gpg", "--homedir", keyring, "--batch", "--verify", signature, artifact)
}

func verifyArchiveLayout(extractRoot, expected string) error {
	entries, err := os.ReadDir(extractRoot)
	if err != nil {
		return err
	}
	if len(entries) != 1 || !entries[0].IsDir() || filepath.Join(extractRoot, entries[0].Name()) != expected {
		return fmt.Errorf("archive must contain exactly one top-level directory named %s", filepath.Base(expected))
	}
	return nil
}

func verifyNoEscapingSymlinks(root string) error {
	rootAbs, err := filepath.Abs(root)
	if err != nil {
		return err
	}
	return filepath.WalkDir(rootAbs, func(path string, entry os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry.Type()&os.ModeSymlink == 0 {
			return nil
		}
		target, err := os.Readlink(path)
		if err != nil {
			return err
		}
		if filepath.IsAbs(target) {
			return fmt.Errorf("archive contains absolute symlink %s -> %s", path, target)
		}
		resolved := filepath.Clean(filepath.Join(filepath.Dir(path), target))
		rel, err := filepath.Rel(rootAbs, resolved)
		if err != nil || rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
			return fmt.Errorf("archive symlink escapes the source root: %s -> %s", path, target)
		}
		return nil
	})
}

func verifyPreparedFiles(cfg *Config, runRoot string, state *State) error {
	artifact := filepath.Join(runRoot, "artifacts", cfg.ArtifactName())
	if err := verifyExactFileDigest(artifact, state.ArtifactSHA512, "source artifact"); err != nil {
		return err
	}
	checksumPath := artifact + ".sha512"
	if err := verifyExactFileDigest(checksumPath, state.ChecksumSHA512, "checksum file"); err != nil {
		return err
	}
	checksum, err := readChecksum(checksumPath, cfg.ArtifactName())
	if err != nil {
		return err
	}
	if checksum != state.ArtifactSHA512 {
		return fmt.Errorf("checksum file does not match artifact")
	}
	return nil
}

func (e Engine) verifyReleaseFiles(cfg *Config, runRoot string, state *State) error {
	if err := verifyPreparedFiles(cfg, runRoot, state); err != nil {
		return err
	}
	artifact := filepath.Join(runRoot, "artifacts", cfg.ArtifactName())
	signature := artifact + ".asc"
	if err := verifyExactFileDigest(signature, state.SignatureSHA512, "detached signature"); err != nil {
		return err
	}
	return e.verifySignatureWithOfficialKeys(cfg, runRoot, signature, artifact)
}

func verifyExactFileDigest(path, expected, label string) error {
	if expected == "" {
		return fmt.Errorf("%s has no frozen SHA-512 digest in state", label)
	}
	actual, err := sha512File(path)
	if err != nil {
		return err
	}
	if actual != expected {
		return fmt.Errorf("%s bytes differ from the frozen candidate", label)
	}
	return nil
}

func sha512File(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()
	h := sha512.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", err
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}

func extractRATJar(zipPath, destination, version string) (string, error) {
	if err := os.RemoveAll(destination); err != nil {
		return "", err
	}
	if err := os.MkdirAll(destination, 0o700); err != nil {
		return "", err
	}
	zr, err := zip.OpenReader(zipPath)
	if err != nil {
		return "", err
	}
	defer zr.Close()
	wanted := "apache-rat-" + version + ".jar"
	for _, entry := range zr.File {
		if filepath.Base(filepath.FromSlash(entry.Name)) != wanted {
			continue
		}
		rc, err := entry.Open()
		if err != nil {
			return "", err
		}
		defer rc.Close()
		jar := filepath.Join(destination, wanted)
		out, err := os.OpenFile(jar, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o600)
		if err != nil {
			return "", err
		}
		_, copyErr := io.Copy(out, rc)
		closeErr := out.Close()
		if copyErr != nil {
			return "", copyErr
		}
		if closeErr != nil {
			return "", closeErr
		}
		return jar, nil
	}
	return "", fmt.Errorf("could not find %s inside Apache RAT archive", wanted)
}

func releaseFileNames(cfg *Config) []string {
	base := cfg.ArtifactName()
	return []string{base, base + ".asc", base + ".sha512"}
}

func removeContained(target, allowedRoot string) error {
	rootAbs, err := filepath.Abs(allowedRoot)
	if err != nil {
		return err
	}
	targetAbs, err := filepath.Abs(target)
	if err != nil {
		return err
	}
	rel, err := filepath.Rel(rootAbs, targetAbs)
	if err != nil || rel == "." || rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
		return fmt.Errorf("refusing to remove path outside the allowed run directory: %s", targetAbs)
	}
	return os.RemoveAll(targetAbs)
}

func copyFile(source, destination string) error {
	in, err := os.Open(source)
	if err != nil {
		return err
	}
	defer in.Close()
	out, err := os.OpenFile(destination, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o600)
	if err != nil {
		return err
	}
	_, copyErr := io.Copy(out, in)
	closeErr := out.Close()
	if copyErr != nil {
		return copyErr
	}
	return closeErr
}

func safeLogName(value string) string {
	value = strings.ReplaceAll(value, ".", "-")
	value = strings.ReplaceAll(value, string(filepath.Separator), "-")
	return value
}

func (e Engine) out() io.Writer {
	if e.Runner.Out != nil {
		return e.Runner.Out
	}
	return os.Stdout
}
