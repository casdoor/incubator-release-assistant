package release

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func writeQueueReleaseConfig(t *testing.T, directory string) string {
	t.Helper()
	cfg := validConfig(t)
	if err := os.WriteFile(filepath.Join(directory, "ira.ps1"), []byte("fixture\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	configDirectory := filepath.Join(directory, "config")
	if err := os.MkdirAll(configDirectory, 0o700); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(configDirectory, "casbin.json")
	raw, err := json.Marshal(cfg)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, raw, 0o600); err != nil {
		t.Fatal(err)
	}
	return path
}

func writeQueue(t *testing.T, directory string, items []QueueItem) *Queue {
	t.Helper()
	queue := Queue{SchemaVersion: "1", Name: "Casbin release queue", Items: items}
	raw, err := json.Marshal(queue)
	if err != nil {
		t.Fatal(err)
	}
	configDirectory := filepath.Join(directory, "config")
	if err := os.MkdirAll(configDirectory, 0o700); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(configDirectory, "queue.json")
	if err := os.WriteFile(path, raw, 0o600); err != nil {
		t.Fatal(err)
	}
	loaded, err := LoadQueue(path)
	if err != nil {
		t.Fatal(err)
	}
	return loaded
}

func queuedCasbinItem() QueueItem {
	return QueueItem{
		ID:            "casbin",
		DisplayName:   "Apache Casbin",
		Repository:    "https://github.com/apache/casbin.git",
		Adapter:       "casbin-go",
		ReleaseConfig: "casbin.json",
		State:         "queued",
	}
}

func TestQueueStatusShowsCurrentAndNextItem(t *testing.T) {
	directory := t.TempDir()
	writeQueueReleaseConfig(t, directory)
	queue := writeQueue(t, directory, []QueueItem{
		queuedCasbinItem(),
		{ID: "sqlx", DisplayName: "Casbin SQLX adapter", Repository: "https://github.com/apache/casbin-sqlx-adapter.git", Adapter: "go", State: "blocked", Note: "adapter has not been implemented"},
	})
	status := (Engine{}).QueueStatus(queue)
	for _, expected := range []string{"Current: 1. Apache Casbin", "Next action: prepare", "Next queued item: 2. Casbin SQLX adapter [blocked]"} {
		if !strings.Contains(status, expected) {
			t.Fatalf("queue status is missing %q:\n%s", expected, status)
		}
	}
}

func TestQueueAdvancesAfterPublicVerification(t *testing.T) {
	directory := t.TempDir()
	configPath := writeQueueReleaseConfig(t, directory)
	queue := writeQueue(t, directory, []QueueItem{
		queuedCasbinItem(),
		{ID: "sqlx", DisplayName: "Casbin SQLX adapter", Repository: "https://github.com/apache/casbin-sqlx-adapter.git", Adapter: "go", State: "blocked", Note: "adapter has not been implemented"},
	})
	cfg, err := LoadConfig(configPath)
	if err != nil {
		t.Fatal(err)
	}
	runRoot, err := cfg.RunRoot()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(runRoot, 0o700); err != nil {
		t.Fatal(err)
	}
	state := NewState(cfg)
	state.Prepared = true
	state.Signed = true
	state.Staged = true
	state.PublicVerified = true
	if err := state.Save(runRoot); err != nil {
		t.Fatal(err)
	}
	status := (Engine{}).QueueStatus(queue)
	if !strings.Contains(status, "Current: 2. Casbin SQLX adapter") {
		t.Fatalf("queue did not advance after completion:\n%s", status)
	}
}

func TestQueueRejectsUnknownFieldsAndUnsafeConfigPaths(t *testing.T) {
	directory := t.TempDir()
	path := filepath.Join(directory, "queue.json")
	content := `{"schemaVersion":"1","name":"queue","items":[{"id":"casbin","displayName":"Apache Casbin","repository":"https://github.com/apache/casbin.git","adapter":"casbin-go","releaseConfig":"../casbin.json","state":"queued","surprise":true}]}`
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := LoadQueue(path); err == nil || (!strings.Contains(err.Error(), "unknown field") && !strings.Contains(err.Error(), "safe relative")) {
		t.Fatalf("unsafe queue was accepted: %v", err)
	}
}

func TestQueueExplainsUnregisteredAdapter(t *testing.T) {
	directory := t.TempDir()
	queue := writeQueue(t, directory, []QueueItem{
		{ID: "future", DisplayName: "Future adapter", Repository: "https://github.com/casbin/future.git", Adapter: "npm", ReleaseConfig: "future.json", State: "queued"},
	})
	status := (Engine{}).QueueStatus(queue)
	for _, expected := range []string{"State: blocked", "Next action: implement and register adapter", "adapter is not registered: npm"} {
		if !strings.Contains(status, expected) {
			t.Fatalf("queue did not explain unsupported adapter %q:\n%s", expected, status)
		}
	}
}
