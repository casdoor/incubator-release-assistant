package release

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

type State struct {
	SchemaVersion   int       `json:"schemaVersion"`
	RunID           string    `json:"runId"`
	ConfigDigest    string    `json:"configSha256"`
	Repository      string    `json:"repository"`
	Commit          string    `json:"commit"`
	ArtifactName    string    `json:"artifactName"`
	ArtifactSHA512  string    `json:"artifactSha512,omitempty"`
	ChecksumSHA512  string    `json:"checksumFileSha512,omitempty"`
	SignatureSHA512 string    `json:"signatureFileSha512,omitempty"`
	Signer          string    `json:"signerFingerprint,omitempty"`
	Prepared        bool      `json:"prepared"`
	Signed          bool      `json:"signed"`
	Staged          bool      `json:"staged"`
	PublicVerified  bool      `json:"publicVerified"`
	PublicURL       string    `json:"publicUrl,omitempty"`
	UpdatedAt       time.Time `json:"updatedAt"`
}

func NewState(cfg *Config) *State {
	return &State{
		SchemaVersion: 2,
		RunID:         cfg.RunID(),
		ConfigDigest:  cfg.Digest(),
		Repository:    cfg.Source.Repository,
		Commit:        stringsLower(cfg.Source.Commit),
		ArtifactName:  cfg.ArtifactName(),
		UpdatedAt:     time.Now().UTC(),
	}
}

func LoadState(runRoot string) (*State, error) {
	raw, err := os.ReadFile(filepath.Join(runRoot, "state.json"))
	if err != nil {
		return nil, err
	}
	var state State
	if err := json.Unmarshal(raw, &state); err != nil {
		return nil, fmt.Errorf("decode state: %w", err)
	}
	return &state, nil
}

func (s *State) Save(runRoot string) error {
	s.UpdatedAt = time.Now().UTC()
	raw, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return err
	}
	raw = append(raw, '\n')
	tmp := filepath.Join(runRoot, "state.json.tmp")
	final := filepath.Join(runRoot, "state.json")
	if err := os.WriteFile(tmp, raw, 0o600); err != nil {
		return err
	}
	if err := os.Remove(final); err != nil && !os.IsNotExist(err) {
		return err
	}
	return os.Rename(tmp, final)
}

func (s *State) VerifyConfig(cfg *Config) error {
	if s.SchemaVersion != 2 {
		return fmt.Errorf("existing state schema %d is not safe for exact-byte resume; use a new RC or --clean only if no staged candidate exists", s.SchemaVersion)
	}
	if s.RunID != cfg.RunID() || s.ConfigDigest != cfg.Digest() || s.Commit != stringsLower(cfg.Source.Commit) {
		return fmt.Errorf("existing state does not match this configuration; use a new RC or --clean only if no staged candidate exists")
	}
	return nil
}

func stringsLower(value string) string {
	for i := 0; i < len(value); i++ {
		if value[i] >= 'A' && value[i] <= 'F' {
			b := []byte(value)
			for j := range b {
				if b[j] >= 'A' && b[j] <= 'F' {
					b[j] += 'a' - 'A'
				}
			}
			return string(b)
		}
	}
	return value
}
