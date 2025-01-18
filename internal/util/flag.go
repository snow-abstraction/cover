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

// createUsageFunc creates a usage function for the flag.Usage.
// The usage string should contain exactly 2 "%s" for the command name. Example:
// `Usage: %s -instance instance.json
//
// %s reads in a problem instance JSON file, solves it and outputs a solution
// to standard out.
//
// Arguments:
// `
func CreateUsageFunc(usage string) func() {
	w := flag.CommandLine.Output()
	return func() {
		fmt.Fprintf(
			w,
			usage,
			os.Args[0],
			os.Args[0])
		flag.PrintDefaults()
	}
}
