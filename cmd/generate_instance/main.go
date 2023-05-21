/*
 Copyright (C) 2023 Douglas Wayne Potter

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
	"log"
	"math/rand"
	"os"
	"sort"

	"github.com/snow-abstraction/cover"
	"golang.org/x/exp/slices"
)

func usage() {
	w := flag.CommandLine.Output()
	fmt.Fprintf(
		w,
		`Usage: %s -seed 1 -m 100 -n 10

%s outputs a random instance to standard out. The instance generated may be
infeasible.

For certain m and n will take a long time because each
subset is generated randomly but must be unique. In fact if the number of
possible nonempty subsets (2^n-1) is less than m then the program will never
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
	var seed int64
	flag.Int64Var(&seed, "seed", 1, "seed for the random generator")
	var m int
	flag.IntVar(&m, "m", 0, "number of subsets")
	flag.IntVar(&ins.N, "n", 0, "number of sets to be covered")
	flag.Parse()

	if m < 0 {
		log.Fatalln("m must be non-negative (0 <= m)")
	}

	if ins.N < 0 {
		log.Fatalln("n must be non-negative (0 <= n)")
	}

	if ins.N > 0 {
		populateInstance(&ins, seed, m)
	}

	b, err := json.MarshalIndent(ins, "", "  ")
	if err != nil {
		log.Fatalln(err)

	}
	fmt.Print(string(b))
}

func populateInstance(ins *cover.Instance, seed int64, m int) {
	gen := rand.New(rand.NewSource(seed))

	// universe of elements to be covered
	u := make([]int, ins.N)
	for i := 0; i < ins.N; i++ {
		u[i] = i
	}
	for j := 0; j < m; j++ {
		for {
			// make random subset u[:k]
			gen.Shuffle(len(u), func(i, j int) { u[i], u[j] = u[j], u[i] })
			k := gen.Intn(ins.N) + 1
			subset := u[:k]
			// sort subset to give it an unique representation
			sort.Ints(subset)

			// only add subset if unique
			// TODO: This introduces quadratic complexity. Ideally we would
			// do a binary search or use some hash table to check if the
			// subset has alredy been added.
			match := false
			for _, s := range ins.Subsets {
				if slices.Equal(subset, s) {
					match = true
					break
				}
			}

			if !match {
				ins.Subsets = append(ins.Subsets, make([]int, len(subset)))
				copy(ins.Subsets[len(ins.Subsets)-1], subset)
				break
			}
		}

		// generate random cost such that 1 <= cost < 10
		ins.Costs = append(ins.Costs, 10.0*(1-0.9*gen.Float64()))
	}

}
