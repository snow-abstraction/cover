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
	"encoding/json"
	"flag"
	"fmt"
	"os"

	"github.com/snow-abstraction/cover"
)

func usage() {
	w := flag.CommandLine.Output()
	fmt.Fprintf(
		w,
		`Usage: %s -seed 1 -m 10 -n 100 -costScale 1.0

%s outputs a random instance to standard out. The instance generated may be
infeasible.

For certain m and n will take a long time because each
subset is generated randomly but must be unique. In fact if the number of
possible nonempty subsets (2^m-1) is less than n then the program will never
terminate.
		
Arguments:
`,
		os.Args[0],
		os.Args[0])
	flag.PrintDefaults()
}

func main() {
	// add empty lists to avoid "Null" text in JSON for zero Instance.
	ins := cover.Instance{Subsets: make([][]int, 0), Costs: make([]float64, 0)}

	flag.Usage = usage
	m := flag.Int("m", 0, "number of sets to be covered")
	n := flag.Int("n", 0, "number of subsets")
	scale := flag.Float64("costScale", 1.0, "scale factor for subset costs")
	seed := flag.Int64("seed", 1, "seed for the random generator")
	flag.Parse()

	if *m < 0 {
		fmt.Fprintln(os.Stderr, "m must be non-negative (0 <= m)")
		os.Exit(1)
	}

	if *n < 0 {
		fmt.Fprintln(os.Stderr, "n must be non-negative (0 <= n)")
		os.Exit(1)
	}

	if *scale <= 0.0 {
		fmt.Fprintln(os.Stderr, "costScale must be strictly positive (0.0 < costScale)")
		os.Exit(1)
	}

	if *n > 0 {
		ins = cover.MakeRandomInstance(*m, *n, *scale, *seed)
	}

	b, err := json.MarshalIndent(ins, "", "  ")
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	fmt.Println(string(b))
}
