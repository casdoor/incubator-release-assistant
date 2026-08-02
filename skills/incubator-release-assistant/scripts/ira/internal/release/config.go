package release

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

const (
	SupportedSchema  = "2"
	SupportedAdapter = "casbin-go"
)

var (
	hex40Pattern    = regexp.MustCompile(`^[0-9a-fA-F]{40}$`)
	versionPattern  = regexp.MustCompile(`^[0-9]+\.[0-9]+\.[0-9]+-incubating$`)
	safeNamePattern = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9._-]*$`)
)

type Config struct {
	SchemaVersion string       `json:"schemaVersion"`
	Project       Project      `json:"project"`
	Source        Source       `json:"source"`
	Release       Release      `json:"release"`
	Checks        Checks       `json:"checks"`
	Signing       Signing      `json:"signing"`
	Distribution  Distribution `json:"distribution"`
	Votes         Votes        `json:"votes"`
	Runtime       Runtime      `json:"runtime"`
	Raw           []byte       `json:"-"`
	Path          string       `json:"-"`
}

type Project struct {
	ID          string `json:"id"`
	DisplayName string `json:"displayName"`
	Adapter     string `json:"adapter"`
	Incubating  bool   `json:"incubating"`
}

type Source struct {
	Repository    string `json:"repository"`
	Commit        string `json:"commit"`
	ArchivePrefix string `json:"archivePrefix"`
}

type Release struct {
	Version          string `json:"version"`
	RC               int    `json:"rc"`
	ArtifactBaseName string `json:"artifactBaseName"`
}

type Checks struct {
	RequiredFiles []string `json:"requiredFiles"`
	RAT           RAT      `json:"rat"`
}

type RAT struct {
	Enabled     bool   `json:"enabled"`
	Version     string `json:"version"`
	ExcludeFile string `json:"excludeFile"`
}

type Signing struct {
	ApacheID         string `json:"apacheId"`
	Fingerprint      string `json:"fingerprint"`
	KeysURL          string `json:"keysUrl"`
	RequireApacheUID bool   `json:"requireApacheUid"`
	MinimumRSABits   int    `json:"minimumRsaBits"`
}

type Distribution struct {
	DevURL     string `json:"devUrl"`
	ReleaseURL string `json:"releaseUrl"`
}

type Votes struct {
	DevList      string `json:"devList"`
	GeneralList  string `json:"generalList"`
	MinimumHours int    `json:"minimumHours"`
}

type Runtime struct {
	StateDirectory string    `json:"stateDirectory"`
	Container      Container `json:"container"`
}

type Container struct {
	Engine  string `json:"engine"`
	Image   string `json:"image"`
	Network string `json:"network"`
}

func LoadConfig(path string) (*Config, error) {
	abs, err := filepath.Abs(path)
	if err != nil {
		return nil, fmt.Errorf("resolve config path: %w", err)
	}
	raw, err := os.ReadFile(abs)
	if err != nil {
		return nil, fmt.Errorf("read config: %w", err)
	}
	parseRaw := bytes.TrimPrefix(raw, []byte{0xEF, 0xBB, 0xBF})
	dec := json.NewDecoder(bytes.NewReader(parseRaw))
	dec.DisallowUnknownFields()
	var cfg Config
	if err := dec.Decode(&cfg); err != nil {
		return nil, fmt.Errorf("decode config: %w", err)
	}
	var trailing any
	if err := dec.Decode(&trailing); err != io.EOF {
		return nil, errors.New("config must contain exactly one JSON object")
	}
	cfg.Raw = raw
	cfg.Path = abs
	if err := cfg.Validate(); err != nil {
		return nil, err
	}
	return &cfg, nil
}

func (c *Config) Validate() error {
	var problems []string
	add := func(condition bool, message string) {
		if condition {
			problems = append(problems, message)
		}
	}

	add(c.SchemaVersion != SupportedSchema, "schemaVersion must be 2")
	add(c.Project.ID != "casbin", "project.id must be casbin in the current release")
	add(c.Project.DisplayName != "Apache Casbin", "project.displayName must be Apache Casbin")
	add(c.Project.Adapter != SupportedAdapter, "project.adapter must be casbin-go; no other adapter is implemented yet")
	add(!c.Project.Incubating, "project.incubating must be true")

	add(c.Source.Repository != "https://github.com/apache/casbin.git", "source.repository must be the canonical https://github.com/apache/casbin.git")
	add(!hex40Pattern.MatchString(c.Source.Commit), "source.commit must be a full 40-character hexadecimal commit")
	add(strings.EqualFold(c.Source.Commit, strings.Repeat("0", 40)), "source.commit is still a placeholder")
	add(!versionPattern.MatchString(c.Release.Version), "release.version must look like 3.11.0-incubating")
	add(c.Release.RC < 1, "release.rc must be at least 1")
	expectedBase := "apache-casbin-" + c.Release.Version + "-src"
	add(c.Release.ArtifactBaseName != expectedBase, "release.artifactBaseName must be "+expectedBase)
	add(c.Source.ArchivePrefix != expectedBase, "source.archivePrefix must match release.artifactBaseName")
	add(!safeName(c.Source.ArchivePrefix), "source.archivePrefix contains unsafe path characters")

	required := stringSet(c.Checks.RequiredFiles)
	add(len(required) != len(c.Checks.RequiredFiles), "checks.requiredFiles must not contain duplicates")
	for _, name := range []string{"LICENSE", "NOTICE", "go.mod", "go.sum", ".rat-excludes"} {
		add(!required[name], "checks.requiredFiles must include "+name)
	}
	add(!required["DISCLAIMER"] && !required["DISCLAIMER-WIP"], "checks.requiredFiles must include DISCLAIMER or DISCLAIMER-WIP")
	for _, name := range c.Checks.RequiredFiles {
		add(!safeRelativePath(name), "unsafe required file path: "+name)
	}
	add(!c.Checks.RAT.Enabled, "checks.rat.enabled must be true")
	add(c.Checks.RAT.Version != "0.18", "checks.rat.version must be the reviewed version 0.18")
	add(!safeRelativePath(c.Checks.RAT.ExcludeFile), "checks.rat.excludeFile must be a safe relative path")
	add(c.Checks.RAT.ExcludeFile != ".rat-excludes", "checks.rat.excludeFile must be .rat-excludes")
	add(!required[c.Checks.RAT.ExcludeFile], "RAT exclude file must also appear in checks.requiredFiles")

	add(strings.TrimSpace(c.Signing.ApacheID) == "" || strings.Contains(c.Signing.ApacheID, "replace-with"), "signing.apacheId is missing or still a placeholder")
	add(!hex40Pattern.MatchString(c.Signing.Fingerprint), "signing.fingerprint must be a full 40-character hexadecimal fingerprint")
	add(strings.EqualFold(c.Signing.Fingerprint, strings.Repeat("0", 40)), "signing.fingerprint is still a placeholder")
	add(c.Signing.KeysURL != "https://dist.apache.org/repos/dist/release/incubator/casbin/KEYS", "signing.keysUrl must be the official Casbin KEYS URL")
	add(!c.Signing.RequireApacheUID, "signing.requireApacheUid must remain enabled for the Casbin release policy")
	add(c.Signing.MinimumRSABits < 4096, "signing.minimumRsaBits must be at least 4096 for the Casbin release policy")

	add(c.Distribution.DevURL != "https://dist.apache.org/repos/dist/dev/incubator/casbin", "distribution.devUrl must be the official Casbin incubator dev URL")
	add(c.Distribution.ReleaseURL != "https://dist.apache.org/repos/dist/release/incubator/casbin", "distribution.releaseUrl must be the official Casbin incubator release URL")
	add(c.Votes.DevList != "dev@casbin.apache.org", "votes.devList must be dev@casbin.apache.org")
	add(c.Votes.GeneralList != "general@incubator.apache.org", "votes.generalList must be general@incubator.apache.org")
	add(c.Votes.MinimumHours < 72, "votes.minimumHours must be at least 72")

	add(c.Runtime.StateDirectory != ".ira", "runtime.stateDirectory must be .ira")
	add(c.Runtime.Container.Engine != "docker" && c.Runtime.Container.Engine != "podman", "runtime.container.engine must be docker or podman")
	add(c.Runtime.Container.Image != "golang:1.24", "runtime.container.image must be the reviewed golang:1.24 image")
	add(c.Runtime.Container.Network != "default" && c.Runtime.Container.Network != "none", "runtime.container.network must be default or none")

	for _, rawURL := range []string{c.Source.Repository, c.Signing.KeysURL, c.Distribution.DevURL, c.Distribution.ReleaseURL} {
		add(!validHTTPSURL(rawURL), "URL must be an absolute HTTPS URL: "+rawURL)
	}

	if len(problems) > 0 {
		sort.Strings(problems)
		return errors.New("configuration is not release-ready:\n- " + strings.Join(problems, "\n- "))
	}
	return nil
}

func (c *Config) Digest() string {
	sum := sha256.Sum256(c.Raw)
	return hex.EncodeToString(sum[:])
}

func (c *Config) RunID() string {
	return fmt.Sprintf("%s-%s-rc%d", c.Project.ID, c.Release.Version, c.Release.RC)
}

func (c *Config) ArtifactName() string { return c.Release.ArtifactBaseName + ".tar.gz" }

func (c *Config) RunRoot() (string, error) {
	base := c.Runtime.StateDirectory
	if !filepath.IsAbs(base) {
		base = filepath.Join(findWorkspaceRoot(filepath.Dir(c.Path)), base)
	}
	abs, err := filepath.Abs(base)
	if err != nil {
		return "", err
	}
	return filepath.Join(abs, "runs", c.RunID()), nil
}

func findWorkspaceRoot(start string) string {
	current, err := filepath.Abs(start)
	if err != nil {
		return start
	}
	for {
		if info, err := os.Stat(filepath.Join(current, "ira.ps1")); err == nil && !info.IsDir() {
			return current
		}
		parent := filepath.Dir(current)
		if parent == current {
			return start
		}
		current = parent
	}
}

func safeName(value string) bool { return safeNamePattern.MatchString(value) }

func safeRelativePath(value string) bool {
	if strings.TrimSpace(value) == "" || filepath.IsAbs(value) {
		return false
	}
	clean := filepath.Clean(value)
	return clean != "." && clean != ".." && !strings.HasPrefix(clean, ".."+string(filepath.Separator)) && !strings.Contains(value, ":")
}

func validHTTPSURL(value string) bool {
	u, err := url.Parse(value)
	return err == nil && u.Scheme == "https" && u.Host != "" && u.User == nil
}

func stringSet(values []string) map[string]bool {
	result := make(map[string]bool, len(values))
	for _, value := range values {
		result[value] = true
	}
	return result
}
