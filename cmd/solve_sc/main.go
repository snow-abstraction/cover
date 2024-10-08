/*
 Copyright (C) 2024 Douglas Wayne Potter

 This program is free software: you can redistribute it and/or modify
 it under the terms of the GNU Affero General Public License as
 published by the Free Software Foundation, either version 3 of the
 License, or (at your option) any later version.

 This program is distributed in the hope that it will be useful,
 but WITHOUT ANY WARRANTY; without even the implied warranty of
 MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 GNU Affero General Public License for more details.

 You should have received a copy of the GNU Affero General Public License
 along with this program.  If not, see <https://www.gnu.org/licenses/>.
*/

// A Branch-and-Bound solver for the "Weighted Exact Cover Problem".
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	"github.com/snow-abstraction/cover"
	"github.com/snow-abstraction/cover/internal/solvers"
)

func usage() {
	w := flag.CommandLine.Output()
	fmt.Fprintf(
		w,
		`Usage: %s -instance instance.json

%s reads in a problem instance JSON file, solves it and outputs a solution
to standard out.

Arguments:
`,
		os.Args[0],
		os.Args[0])
	flag.PrintDefaults()
}

func main() {

	flag.Usage = usage
	filename := flag.String("instance", "",
		"instance filename. The file should end in .json (or .JSON) or .mps (or .MPS). MPS support is experimental.")
	logLevel := flag.String("logLevel", "Info", "log level (Debug, Info, Warn, Error)")
	flag.Parse()

	level := parseLogLevel(*logLevel)
	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
		AddSource: true,
		Level:     level,
	})))

	if *filename == "" {
		fmt.Fprintln(os.Stderr, "Please supply the instance file name")
		os.Exit(1)
	}

	ins, err := readInstance(filename)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to read instance due to error: %s\n", err)
		os.Exit(1)
	}

	solverInstance, err := solvers.MakeInstance(ins.M, ins.Subsets, ins.Costs)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	sol, err := solvers.SolveByBranchAndBound(solverInstance)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to optimal solution due to error: %s\n", err)
		os.Exit(1)
	}
	fmt.Printf("Solution: %+v\n", sol)
}

func readInstance(filename *string) (*cover.Instance, error) {
	ext := filepath.Ext(*filename)
	lowerExt := strings.ToLower(ext)
	switch lowerExt {
	case ".json":
		return readJsonInstance(filename)
	case ".mps":
		return cover.ReadMPSInstance(filename)
	}

	return nil, fmt.Errorf(
		"the file extension should be .JSON, .json, .MPS or .mps and not %s", ext)

}

func readJsonInstance(filename *string) (*cover.Instance, error) {
	b, err := os.ReadFile(*filename)
	if err != nil {
		return nil, err
	}

	var ins *cover.Instance
	if err := json.Unmarshal(b, ins); err != nil {
		return nil, err
	}
	return ins, nil
}

func parseLogLevel(level string) slog.Level {
	switch level {
	case "Debug":
		return slog.LevelDebug
	case "Info":
		return slog.LevelInfo
	case "Warn":
		return slog.LevelWarn
	case "Error":
		return slog.LevelError
	}
	slog.Error("unknown log level. defaulting to Info")

	return slog.LevelInfo
}
