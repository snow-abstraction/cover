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
package main

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	"github.com/snow-abstraction/cover"
	"github.com/snow-abstraction/cover/internal/util"
)

func main() {

	flags := util.NewFlagSet(`Usage: %s -instance instance.json

%s reads in a problem instance JSON or MPS file and outputs it to standard out using
Go debug formatting.

Arguments:
`)
	filename := flags.String("instance", "",
		"instance filename. The file should end in .json (or .JSON) or .mps (or .MPS). MPS support is experimental.")
	logLevel := flags.String("logLevel", "Info", "log level (Debug, Info, Warn, Error)")
	flags.Parse()

	level := parseLogLevel(*logLevel)
	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
		AddSource: true,
		Level:     level,
	})))

	if *filename == "" {
		fmt.Fprintln(os.Stderr, "Please supply the instance file name")
		os.Exit(1)
	}

	ins, err := readInstance(*filename)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to read instance due to error: %s\n", err)
		os.Exit(1)
	}

	fmt.Printf("Instance: %#v\n", ins)
}

func readInstance(filename string) (*cover.Instance, error) {
	ext := filepath.Ext(filename)
	lowerExt := strings.ToLower(ext)
	switch lowerExt {
	case ".json":
		return cover.ReadJsonInstance(filename)
	case ".mps":
		return cover.ReadMPSInstance(filename)
	}

	return nil, fmt.Errorf(
		"the file extension should be .JSON, .json, .MPS or .mps and not %s", ext)

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
