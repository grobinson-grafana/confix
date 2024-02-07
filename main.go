package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/golang/protobuf/proto"
	"github.com/grafana/mimir/pkg/alertmanager/alertspb"
)

func parseFlags(args []string) (string, string) {
	var readFile, writeFile string
	if len(args) <= 1 {
		exitWithErr(fmt.Errorf("An input file must be specified"))
	}
	readFile = args[1]
	if len(args) <= 2 {
		dir, file := filepath.Split(readFile)
		writeFile = filepath.Join(dir, file+".out")
	}
	return readFile, writeFile
}

func exitWithErr(err error) {
	fmt.Fprintln(os.Stderr, err)
	os.Exit(1)
}

func main() {
	readFile, writeFile := parseFlags(os.Args)
	b, err := os.ReadFile(readFile)
	if err != nil {
		exitWithErr(fmt.Errorf("failed to read file: %w", err))
	}
	tmp1 := alertspb.AlertConfigDesc{}
	if err = proto.Unmarshal(b, &tmp1); err != nil {
		exitWithErr(fmt.Errorf("failed to unmarshal proto: %w", err))
	}
	tmp2, err := fix(tmp1)
	if err != nil {
		exitWithErr(fmt.Errorf("failed to fix proto: %w", err))
	}
	ok, diff, err := isEqual(tmp1, *tmp2)
	if err != nil {
		exitWithErr(fmt.Errorf("failed to check if before and after proto are equal: %w", err))
	}
	if !ok {
		fmt.Printf("before and after proto are not equal: %s", diff)
	}
	b, err = proto.Marshal(tmp2)
	if err != nil {
		exitWithErr(fmt.Errorf("failed to marshal proto: %w", err))
	}
	if err = os.WriteFile(writeFile, b, 0644); err != nil {
		exitWithErr(fmt.Errorf("failed to write file: %w", err))
	}
}
