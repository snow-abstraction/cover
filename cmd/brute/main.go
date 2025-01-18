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

// A brute force solver for the "Weighted Exact Cover Problem".
package main

import (
	"flag"
	"fmt"
	"log"

	"github.com/snow-abstraction/cover"
	"github.com/snow-abstraction/cover/internal/util"
	"github.com/snow-abstraction/cover/solvers"
)

func main() {
	flag.Usage = util.CreateUsageFunc(`Usage: %s -instance instance.json

%s reads in a problem instance JSON file, solves it and outputs a solution
to standard out.

Arguments:
`)
	filename := flag.String("instance", "", "instance JSON filename")
	flag.Parse()

	if *filename == "" {
		log.Fatalln("Please supply the instance file name")
	}

	ins, err := cover.ReadJsonInstance(*filename)
	if err != nil {
		log.Fatalln(err)
	}

	sol, err := solvers.SolveByBruteForce(*ins)
	if err != nil {
		fmt.Printf("failed to optimal solution due to error: %s", err)
	}
	fmt.Printf("Solution: %+v\n", sol)
}
