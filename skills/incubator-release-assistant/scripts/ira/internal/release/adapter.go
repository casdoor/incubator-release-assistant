package release

import (
	"fmt"
	"regexp"
	"sort"
	"strings"
)

var casbinIncubatingVersionPattern = regexp.MustCompile(`^[0-9]+\.[0-9]+\.[0-9]+-incubating$`)

// Adapter owns project-specific release rules. The engine owns the shared
// state, evidence, signing, staging, and queue mechanics.
type Adapter interface {
	ID() string
	Description() string
	Validate(*Config) []string
}

type casbinGoAdapter struct{}

func (casbinGoAdapter) ID() string {
	return "casbin-go"
}

func (casbinGoAdapter) Description() string {
	return "Apache Casbin Go source RC and the tag-driven GitHub/Go module publication contract"
}

func (casbinGoAdapter) Validate(c *Config) []string {
	var problems []string
	add := func(condition bool, message string) {
		if condition {
			problems = append(problems, message)
		}
	}

	add(c.Project.ID != "casbin", "project.id must be casbin in the current release")
	add(c.Project.DisplayName != "Apache Casbin", "project.displayName must be Apache Casbin")
	add(!c.Project.Incubating, "project.incubating must be true")
	add(c.Source.Repository != "https://github.com/apache/casbin.git", "source.repository must be the canonical https://github.com/apache/casbin.git")
	add(!hex40Pattern.MatchString(c.Source.Commit), "source.commit must be a full 40-character hexadecimal commit")
	add(strings.EqualFold(c.Source.Commit, strings.Repeat("0", 40)), "source.commit is still a placeholder")
	add(!casbinIncubatingVersionPattern.MatchString(c.Release.Version), "release.version must look like 3.11.0-incubating")
	add(c.Release.RC < 1, "release.rc must be at least 1")
	expectedBase := "apache-casbin-" + c.Release.Version + "-src"
	add(c.Release.ArtifactBaseName != expectedBase, "release.artifactBaseName must be "+expectedBase)
	add(c.Source.ArchivePrefix != expectedBase, "source.archivePrefix must match release.artifactBaseName")

	required := stringSet(c.Checks.RequiredFiles)
	for _, name := range []string{"LICENSE", "NOTICE", "go.mod", "go.sum", ".rat-excludes"} {
		add(!required[name], "checks.requiredFiles must include "+name)
	}
	add(!required["DISCLAIMER"] && !required["DISCLAIMER-WIP"], "checks.requiredFiles must include DISCLAIMER or DISCLAIMER-WIP")
	add(!c.Checks.RAT.Enabled, "checks.rat.enabled must be true")
	add(c.Checks.RAT.Version != "0.18", "checks.rat.version must be the reviewed version 0.18")
	add(c.Checks.RAT.ExcludeFile != ".rat-excludes", "checks.rat.excludeFile must be .rat-excludes")
	add(!required[c.Checks.RAT.ExcludeFile], "RAT exclude file must also appear in checks.requiredFiles")

	add(strings.TrimSpace(c.Signing.ApacheID) == "" || strings.Contains(c.Signing.ApacheID, "replace-with"), "signing.apacheId is missing or still a placeholder")
	add(!hex40Pattern.MatchString(c.Signing.Fingerprint), "signing.fingerprint must be a full 40-character hexadecimal fingerprint")
	add(strings.EqualFold(c.Signing.Fingerprint, strings.Repeat("0", 40)), "signing.fingerprint is still a placeholder")
	add(c.Signing.KeysURL != "https://dist.apache.org/repos/dist/release/incubator/casbin/KEYS", "signing.keysUrl must be the official Casbin KEYS URL")
	add(!c.Signing.RequireApacheUID, "signing.requireApacheUid must remain enabled for the Casbin release policy")
	add(c.Signing.MinimumRSABits < 4096, "signing.minimumRsaBits must be at least 4096 for the Casbin release policy")
	add(c.Distribution.DevURL != "https://dist.apache.org/repos/dist/dev/incubator/casbin", "distribution.devUrl must be the official Casbin incubator dev URL")
	return problems
}

var adapters = []Adapter{casbinGoAdapter{}}

func FindAdapter(id string) (Adapter, bool) {
	for _, adapter := range adapters {
		if adapter.ID() == id {
			return adapter, true
		}
	}
	return nil, false
}

func DescribeAdapters() string {
	registered := append([]Adapter(nil), adapters...)
	sort.Slice(registered, func(i, j int) bool { return registered[i].ID() < registered[j].ID() })
	result := "Registered IRA adapters:\n"
	for _, adapter := range registered {
		result += fmt.Sprintf("  - %s: %s\n", adapter.ID(), adapter.Description())
	}
	return result
}
