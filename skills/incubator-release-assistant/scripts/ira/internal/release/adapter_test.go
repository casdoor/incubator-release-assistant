package release

import (
	"strings"
	"testing"
)

func TestCasbinGoAdapterIsRegistered(t *testing.T) {
	adapter, ok := FindAdapter("casbin-go")
	if !ok {
		t.Fatal("casbin-go adapter is not registered")
	}
	if !strings.Contains(adapter.Description(), "Casbin Go") {
		t.Fatalf("adapter description is not operator-facing: %q", adapter.Description())
	}
	if !strings.Contains(DescribeAdapters(), "casbin-go") {
		t.Fatal("adapter list does not include casbin-go")
	}
}

func TestCasbinGoAdapterPreservesExistingReleasePolicy(t *testing.T) {
	tests := []struct {
		name     string
		mutate   func(*Config)
		expected string
	}{
		{"project ID", func(c *Config) { c.Project.ID = "other" }, "project.id must be casbin"},
		{"display name", func(c *Config) { c.Project.DisplayName = "Other" }, "project.displayName must be Apache Casbin"},
		{"incubating", func(c *Config) { c.Project.Incubating = false }, "project.incubating must be true"},
		{"source repository", func(c *Config) { c.Source.Repository = "https://example.com/casbin.git" }, "source.repository must be the canonical"},
		{"placeholder commit", func(c *Config) { c.Source.Commit = strings.Repeat("0", 40) }, "source.commit is still a placeholder"},
		{"release version", func(c *Config) { c.Release.Version = "3.11.0" }, "release.version must look like 3.11.0-incubating"},
		{"release RC", func(c *Config) { c.Release.RC = 0 }, "release.rc must be at least 1"},
		{"artifact name", func(c *Config) { c.Release.ArtifactBaseName = "different" }, "release.artifactBaseName must be"},
		{"archive prefix", func(c *Config) { c.Source.ArchivePrefix = "different" }, "source.archivePrefix must match"},
		{"required files", func(c *Config) {
			c.Checks.RequiredFiles = []string{"LICENSE", "NOTICE", "DISCLAIMER", "go.mod", "go.sum"}
		}, "checks.requiredFiles must include .rat-excludes"},
		{"RAT version", func(c *Config) { c.Checks.RAT.Version = "0.17" }, "checks.rat.version must be the reviewed version 0.18"},
		{"KEYS URL", func(c *Config) { c.Signing.KeysURL = "https://example.com/KEYS" }, "signing.keysUrl must be the official Casbin KEYS URL"},
		{"Apache UID", func(c *Config) { c.Signing.RequireApacheUID = false }, "signing.requireApacheUid must remain enabled"},
		{"RSA bits", func(c *Config) { c.Signing.MinimumRSABits = 2048 }, "signing.minimumRsaBits must be at least 4096"},
		{"dist URL", func(c *Config) { c.Distribution.DevURL = "https://example.com/dist" }, "distribution.devUrl must be the official Casbin incubator dev URL"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			cfg := validConfig(t)
			test.mutate(cfg)
			if err := cfg.Validate(); err == nil || !strings.Contains(err.Error(), test.expected) {
				t.Fatalf("existing Casbin policy %q was not preserved: %v", test.expected, err)
			}
		})
	}
}
