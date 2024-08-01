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
	"log"
	"math"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"

	"github.com/snow-abstraction/cover"
)

const (
	resultDelimiter = "solve_sc_result:"
)

// main creates some instance test data. See the `func usage()` or run with `-help` for more details.
func main() {
	flag.Usage = usage
	pythonSolverPath := flag.String("solver", "tools/solve_sc.py", "python solver path")
	outputDir := flag.String("output", "testdata/instances",
		"output directory for instances and python solution files")
	specificationsPath := flag.String("specifications", "testdata/instance_specifications.json",
		"instance specifications file")
	verbose := flag.Bool("verbose", false, "more verbose logging")
	flag.Parse()

	if *verbose {
		log.Println("Running with flags:")
		flag.VisitAll(func(f *flag.Flag) {
			log.Printf("%s: %s\n", f.Name, f.Value)
		})
	}

	specifications := createSpecifications(*outputDir, verbose)

	b, err := json.MarshalIndent(specifications, "", "  ")
	if err != nil {
		log.Panic(err)
	}
	if err := os.WriteFile(*specificationsPath, b, 0600); err != nil {
		log.Panic(err)
	}

	createInstanceFiles(specifications)
	solveInstances(specifications, *pythonSolverPath, verbose)
}

func usage() {
	w := flag.CommandLine.Output()
	fmt.Fprintf(
		w,
		`Usage: %s -verbose

%s creates some instance test data. Specifically it:
1. generates some set covering instances
2. these instances are saved as JSON
3. these instances are solved using an independent solver and solutions are saved as JSON
4. JSON information about all these instances and solutions are saved to the specifications file

Arguments:
`,
		os.Args[0],
		os.Args[0])
	flag.PrintDefaults()
}

func createSpecifications(outputDir string, verbose *bool) []cover.TestInstanceSpecification {
	specifications := make([]cover.TestInstanceSpecification, 0)
	var seed int64
	numberOfElements := []int{1, 2, 3, 4}

	for _, m := range numberOfElements {
		maxN := int(3*math.Exp2(float64(m))) / 4
		for n := 1; n <= maxN; n++ {
			// two instances for every (m, n)
			for j := 0; j < 2; j++ {
				instanceName := fmt.Sprintf("instance_%d_%d_%d.json", m, n, seed)
				instancePath := filepath.Join(outputDir, instanceName)

				solutionFileName := fmt.Sprintf("python_solution_%d_%d_%d.json", m, n, seed)
				solutionPath := filepath.Join(outputDir, solutionFileName)

				specifications = append(specifications,
					cover.TestInstanceSpecification{NumElements: m, NumSubSets: n, Seed: seed, InstancePath: instancePath, PythonSolutionPath: solutionPath})
				if *verbose {
					log.Printf("will generated instance using number of elements: %d, number of subsets: %d and random seed: %d\n",
						m, n, seed)
				}
				seed++
			}
		}
	}
	return specifications
}

func createInstanceFiles(specifications []cover.TestInstanceSpecification) {
	log.Printf("Creating %d test instance files", len(specifications))
	for _, spec := range specifications {
		ins := cover.MakeRandomInstance(spec.NumElements, spec.NumSubSets, spec.Seed)
		b, err := json.MarshalIndent(ins, "", "  ")
		if err != nil {
			log.Panic(err)
		}
		if err := os.WriteFile(spec.InstancePath, b, 0600); err != nil {
			log.Panic(err)
		}
	}
}

// Solve instance using the python script
// Extract the result from stdout
func solveInstances(specifications []cover.TestInstanceSpecification, pythonSolverPath string, verbose *bool) {
	log.Printf("Solving %d test instances", len(specifications))
	var wg sync.WaitGroup
	for _, spec := range specifications {
		wg.Add(1)
		instancePath, solutionPath := spec.InstancePath, spec.PythonSolutionPath
		go func() {
			defer wg.Done()

			cmd := exec.Command("python", pythonSolverPath, instancePath)
			if *verbose {
				log.Printf("running %s\n", cmd)
			}

			stdout, err := cmd.Output()
			if err != nil {
				log.Panicf("running '%s' resulted in error '%s'", cmd, err)
			}

			s := string(stdout)
			if !strings.Contains(s, resultDelimiter) {
				log.Panicf("output from running %s is missing the result delimiter %s",
					pythonSolverPath, resultDelimiter)

			}
			splitOutput := strings.Split(s, resultDelimiter)
			resultStr := splitOutput[len(splitOutput)-1]

			if err := os.WriteFile(solutionPath, []byte(resultStr), 0600); err != nil {
				log.Panic(err)
			}
		}()
	}

	wg.Wait()
}
