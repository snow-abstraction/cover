/*
 Copyright (C) 2025 Douglas Wayne Potter

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

package util

import (
	"flag"
	"fmt"
	"os"
)

// Embedding of flag.FlatSet to have a connivent Parse()
// receiver.
type FlagSet struct {
	*flag.FlagSet
}

// createUsageFunc creates a new *Flagset using the supplied usage string.
//
// The usage string should contain exactly 2 "%s" for the command name. Example:
// `Usage: %s -instance instance.json
//
// %s reads in a problem instance JSON file, solves it and outputs a solution
// to standard out.
//
// Arguments:
// `
func NewFlagSet(usage string) *FlagSet {
	fs := flag.NewFlagSet(os.Args[0], flag.ExitOnError)
	fs.Usage = func() {
		fmt.Fprintf(
			flag.CommandLine.Output(),
			usage,
			os.Args[0],
			os.Args[0])
		fs.PrintDefaults()
	}

	return &FlagSet{fs}
}

// Parse parses the command-line flags from os.Args[1:].
// Must be called after all flags are defined and before flags are accessed by the program.
// Note: this documentation was copied from flags.Parse()
func (fs *FlagSet) Parse() {
	fs.FlagSet.Parse(os.Args[1:])
}
