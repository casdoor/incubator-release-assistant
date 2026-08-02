package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/EmryZhang/incubator-release-assistant/skills/incubator-release-assistant/scripts/ira/internal/release"
)

const version = "0.1.0"

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
	set := flag.NewFlagSet(command, flag.ExitOnError)
	configPath := set.String("config", "", "path to a release-ready JSON configuration")
	confirmation := set.String("confirm", "", "exact confirmation required for signing or staging")
	clean := set.Bool("clean", false, "remove only the matching unstaged local run before preparing")
	set.SetOutput(os.Stderr)
	if err := set.Parse(os.Args[2:]); err != nil {
		fail(err)
	}
	if *configPath == "" {
		fail(fmt.Errorf("--config is required"))
	}
	cfg, err := release.LoadConfig(*configPath)
	if err != nil {
		fail(err)
	}
	engine := release.Engine{Runner: release.Runner{Out: os.Stdout}}

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
  %s validate      --config <file>
  %s plan          --config <file>
  %s prepare       --config <file> [--clean]
  %s sign          --config <file> --confirm <artifact-sha512>
  %s stage         --config <file> --confirm "STAGE RC<number>"
  %s verify-public --config <file>
  %s version

The current build intentionally supports only the Apache Casbin Go adapter.
`, name, name, name, name, name, name, name)
}

func fail(err error) {
	fmt.Fprintln(os.Stderr, "FAILED:", err)
	os.Exit(1)
}
