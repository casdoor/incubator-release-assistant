package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/EmryZhang/incubator-release-assistant/skills/incubator-release-assistant/scripts/ira/internal/release"
)

const version = "0.2.0"

func main() {
	if len(os.Args) < 2 {
		usage()
		os.Exit(2)
	}
	command := os.Args[1]
	if command == "version" || command == "--version" {
		fmt.Println("ira", version)
		return
	}
	if command == "adapters" {
		fmt.Print(release.DescribeAdapters())
		return
	}
	set := flag.NewFlagSet(command, flag.ExitOnError)
	configPath := set.String("config", "", "path to a release-ready JSON configuration")
	queuePath := set.String("queue", "", "path to a release queue JSON configuration")
	confirmation := set.String("confirm", "", "exact confirmation required for signing or staging")
	clean := set.Bool("clean", false, "remove only the matching unstaged local run before preparing")
	set.SetOutput(os.Stderr)
	if err := set.Parse(os.Args[2:]); err != nil {
		fail(err)
	}
	engine := release.Engine{Runner: release.Runner{Out: os.Stdout}}
	if command == "doctor" {
		report := engine.Doctor(*configPath)
		encoder := json.NewEncoder(os.Stdout)
		encoder.SetIndent("", "  ")
		if err := encoder.Encode(report); err != nil {
			fail(err)
		}
		return
	}
	if command == "queue-status" || command == "queue-prepare" {
		if *queuePath == "" {
			fail(fmt.Errorf("--queue is required"))
		}
		if *clean {
			fail(fmt.Errorf("--clean is not available for queue commands; inspect each run before removing local evidence"))
		}
		queue, err := release.LoadQueue(*queuePath)
		if err != nil {
			fail(err)
		}
		switch command {
		case "queue-status":
			fmt.Print(engine.QueueStatus(queue))
		case "queue-prepare":
			if err := engine.PrepareCurrentQueueItem(queue); err != nil {
				fail(err)
			}
		}
		return
	}
	if *configPath == "" {
		fail(fmt.Errorf("--config is required"))
	}
	cfg, err := release.LoadConfig(*configPath)
	if err != nil {
		fail(err)
	}
	switch command {
	case "validate":
		fmt.Printf("Release-ready configuration is valid.\nConfig: %s\nRun: %s\nCommit: %s\n", cfg.Path, cfg.RunID(), cfg.Source.Commit)
	case "plan":
		fmt.Print(engine.Plan(cfg))
	case "prepare":
		_, err = engine.Prepare(cfg, *clean)
	case "sign":
		_, err = engine.Sign(cfg, *confirmation)
	case "stage":
		_, err = engine.Stage(cfg, *confirmation)
	case "verify-public":
		err = engine.VerifyPublic(cfg)
	default:
		usage()
		os.Exit(2)
	}
	if err != nil {
		fail(err)
	}
}

func usage() {
	name := filepath.Base(os.Args[0])
	fmt.Fprintf(os.Stderr, `IRA - Apache Incubator release assistant

Usage:
  %s doctor        [--config <file>]
  %s validate      --config <file>
  %s plan          --config <file>
  %s prepare       --config <file> [--clean]
  %s sign          --config <file> --confirm <artifact-sha512>
  %s stage         --config <file> --confirm "STAGE RC<number>"
  %s verify-public --config <file>

  %s queue-status  --queue <file>
  %s queue-prepare --queue <file>
  %s adapters
  %s version

Queue commands traverse one ordered release queue. They print the current item,
its next required action, and the next queued item. queue-prepare runs only the
current item's safe prepare step; signing and staging remain separate commands.

The current build intentionally supports only the Apache Casbin Go adapter.
`, name, name, name, name, name, name, name, name, name, name, name)
}

func fail(err error) {
	guidance := release.GuidanceForError(err)
	fmt.Fprintf(os.Stderr, "FAILED [%s]: %v\nGuidance: %s\nNext: %s\n", guidance.Code, err, guidance.Reference, guidance.NextAction)
	os.Exit(1)
}
