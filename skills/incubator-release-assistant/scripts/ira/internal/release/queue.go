package release

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

const SupportedQueueSchema = "1"

// Queue is an ordered, reviewed release worklist.  It deliberately contains
// data only: the engine never accepts commands from the queue.
type Queue struct {
	SchemaVersion string      `json:"schemaVersion"`
	Name          string      `json:"name"`
	Items         []QueueItem `json:"items"`
	Path          string      `json:"-"`
}

type QueueItem struct {
	ID            string `json:"id"`
	DisplayName   string `json:"displayName"`
	Repository    string `json:"repository"`
	Adapter       string `json:"adapter"`
	ReleaseConfig string `json:"releaseConfig,omitempty"`
	State         string `json:"state"`
	Note          string `json:"note,omitempty"`
}

type QueueProgress struct {
	Item       QueueItem
	State      string
	NextAction string
	Detail     string
	Config     *Config
}

func LoadQueue(path string) (*Queue, error) {
	abs, err := filepath.Abs(path)
	if err != nil {
		return nil, fmt.Errorf("resolve queue path: %w", err)
	}
	raw, err := os.ReadFile(abs)
	if err != nil {
		return nil, fmt.Errorf("read queue: %w", err)
	}
	dec := json.NewDecoder(bytes.NewReader(bytes.TrimPrefix(raw, []byte{0xEF, 0xBB, 0xBF})))
	dec.DisallowUnknownFields()
	var queue Queue
	if err := dec.Decode(&queue); err != nil {
		return nil, fmt.Errorf("decode queue: %w", err)
	}
	var trailing any
	if err := dec.Decode(&trailing); err != io.EOF {
		return nil, errors.New("queue must contain exactly one JSON object")
	}
	queue.Path = abs
	if err := queue.Validate(); err != nil {
		return nil, err
	}
	return &queue, nil
}

func (q *Queue) Validate() error {
	var problems []string
	add := func(condition bool, message string) {
		if condition {
			problems = append(problems, message)
		}
	}
	add(q.SchemaVersion != SupportedQueueSchema, "queue schemaVersion must be 1")
	add(strings.TrimSpace(q.Name) == "", "queue name is required")
	add(len(q.Items) == 0, "queue must contain at least one item")
	seen := map[string]bool{}
	for index, item := range q.Items {
		prefix := fmt.Sprintf("items[%d]", index)
		add(!safeName(item.ID), prefix+".id must be a safe identifier")
		add(seen[item.ID], prefix+".id is duplicated: "+item.ID)
		seen[item.ID] = true
		add(strings.TrimSpace(item.DisplayName) == "", prefix+".displayName is required")
		add(!validHTTPSURL(item.Repository), prefix+".repository must be an absolute HTTPS URL")
		add(!safeName(item.Adapter), prefix+".adapter must be a safe identifier")
		switch item.State {
		case "queued":
			add(strings.TrimSpace(item.ReleaseConfig) == "", prefix+".releaseConfig is required for queued items")
			add(!safeQueueConfigPath(item.ReleaseConfig), prefix+".releaseConfig must be a safe relative JSON path")
		case "blocked", "manual":
			add(strings.TrimSpace(item.Note) == "", prefix+".note is required for "+item.State+" items")
		case "complete":
			add(strings.TrimSpace(item.Note) == "", prefix+".note is required for complete items")
		default:
			add(true, prefix+".state must be queued, blocked, manual, or complete")
		}
	}
	if len(problems) > 0 {
		sort.Strings(problems)
		return errors.New("release queue is invalid:\n- " + strings.Join(problems, "\n- "))
	}
	return nil
}

func safeQueueConfigPath(value string) bool {
	if strings.TrimSpace(value) == "" || filepath.IsAbs(value) || filepath.Ext(value) != ".json" {
		return false
	}
	clean := filepath.Clean(value)
	return clean != "." && clean != ".." && !strings.HasPrefix(clean, ".."+string(filepath.Separator)) && !strings.Contains(value, ":")
}

func (q *Queue) configPath(item QueueItem) string {
	return filepath.Join(filepath.Dir(q.Path), filepath.FromSlash(item.ReleaseConfig))
}

func (q *Queue) Progress(item QueueItem) QueueProgress {
	progress := QueueProgress{Item: item}
	switch item.State {
	case "blocked", "manual":
		progress.State = item.State
		progress.NextAction = "resolve queue note"
		progress.Detail = item.Note
		return progress
	case "complete":
		progress.State = "complete"
		progress.NextAction = "none"
		progress.Detail = item.Note
		return progress
	}
	if _, ok := FindAdapter(item.Adapter); !ok {
		progress.State = "blocked"
		progress.NextAction = "implement and register adapter"
		progress.Detail = "adapter is not registered: " + item.Adapter
		return progress
	}
	cfg, err := LoadConfig(q.configPath(item))
	if err != nil {
		progress.State = "blocked"
		progress.NextAction = "fix release configuration"
		progress.Detail = err.Error()
		return progress
	}
	progress.Config = cfg
	if cfg.Project.Adapter != item.Adapter {
		progress.State = "blocked"
		progress.NextAction = "use a matching adapter"
		progress.Detail = fmt.Sprintf("queue adapter %q differs from release config adapter %q", item.Adapter, cfg.Project.Adapter)
		return progress
	}
	if cfg.Source.Repository != item.Repository {
		progress.State = "blocked"
		progress.NextAction = "align repository identity"
		progress.Detail = fmt.Sprintf("queue repository %q differs from release config repository %q", item.Repository, cfg.Source.Repository)
		return progress
	}
	runRoot, err := cfg.RunRoot()
	if err != nil {
		progress.State = "blocked"
		progress.NextAction = "fix release configuration"
		progress.Detail = err.Error()
		return progress
	}
	state, err := LoadState(runRoot)
	if err != nil {
		if os.IsNotExist(err) {
			progress.State = "ready"
			progress.NextAction = "prepare"
			progress.Detail = "no local candidate state exists"
			return progress
		}
		progress.State = "blocked"
		progress.NextAction = "inspect local candidate state"
		progress.Detail = err.Error()
		return progress
	}
	if err := state.VerifyConfig(cfg); err != nil {
		progress.State = "blocked"
		progress.NextAction = "inspect local candidate state"
		progress.Detail = err.Error()
		return progress
	}
	switch {
	case state.PublicVerified:
		progress.State = "complete"
		progress.NextAction = "none"
		progress.Detail = "public candidate bytes were verified"
	case state.Staged:
		progress.State = "blocked"
		progress.NextAction = "inspect staging evidence"
		progress.Detail = "staged state is missing public verification"
	case state.Signed:
		progress.State = "ready"
		progress.NextAction = "stage"
		progress.Detail = "candidate signature is frozen"
	case state.Prepared:
		progress.State = "ready"
		progress.NextAction = "sign"
		progress.Detail = "prepared artifact is ready for the explicit signing confirmation"
	default:
		progress.State = "blocked"
		progress.NextAction = "inspect incomplete prepare evidence"
		progress.Detail = "local candidate state is neither prepared nor complete"
	}
	return progress
}

func (q *Queue) Current() (int, QueueProgress) {
	for index, item := range q.Items {
		progress := q.Progress(item)
		if progress.State != "complete" {
			return index, progress
		}
	}
	return -1, QueueProgress{}
}

func (e Engine) QueueStatus(q *Queue) string {
	var output strings.Builder
	fmt.Fprintf(&output, "Release queue: %s\n", q.Name)
	for index, item := range q.Items {
		progress := q.Progress(item)
		fmt.Fprintf(&output, "  %d. %s [%s] — %s\n", index+1, item.DisplayName, progress.State, progress.NextAction)
	}
	index, current := q.Current()
	if index == -1 {
		output.WriteString("\nCurrent: none; every queue item is complete.\nNext queued item: none.\n")
		return output.String()
	}
	fmt.Fprintf(&output, "\nCurrent: %d. %s\nState: %s\nNext action: %s\nDetail: %s\n", index+1, current.Item.DisplayName, current.State, current.NextAction, current.Detail)
	for next := index + 1; next < len(q.Items); next++ {
		candidate := q.Progress(q.Items[next])
		if candidate.State != "complete" {
			fmt.Fprintf(&output, "Next queued item: %d. %s [%s]\n", next+1, candidate.Item.DisplayName, candidate.State)
			return output.String()
		}
	}
	output.WriteString("Next queued item: none.\n")
	return output.String()
}

func (e Engine) PrepareCurrentQueueItem(q *Queue) error {
	index, current := q.Current()
	if index == -1 {
		fmt.Fprint(e.out(), e.QueueStatus(q))
		return nil
	}
	fmt.Fprintf(e.out(), "[IRA] QUEUE CURRENT %d/%d: %s\n", index+1, len(q.Items), current.Item.DisplayName)
	if current.State != "ready" || current.NextAction != "prepare" || current.Config == nil {
		return fmt.Errorf("queue cannot prepare %s: %s; next action is %s", current.Item.DisplayName, current.Detail, current.NextAction)
	}
	if _, err := e.Prepare(current.Config, false); err != nil {
		return fmt.Errorf("queue item %s prepare failed: %w", current.Item.ID, err)
	}
	fmt.Fprintln(e.out(), "[IRA] QUEUE UPDATED")
	fmt.Fprint(e.out(), e.QueueStatus(q))
	return nil
}
