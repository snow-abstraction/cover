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
	"log/slog"
	"math"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"

	"github.com/snow-abstraction/cover"
)

const (
	resultDelimiter           = "solve_sc_result:"
	defaultSpecificationsPath = "testdata/[suite_name]_instance_specifications.json"
)

// main creates some instance test data. See the `func usage()` or run with `-help` for more details.
func main() {

	flag.Usage = usage
	pythonSolverPath := flag.String("solver", "tools/solve_sc.py", "python solver path")
	outputDir := flag.String("output", "testdata/instances",
		"output directory for instances and python solution files")
	suite := flag.String("suite", "tiny", "instance suite to generate (tiny)")
	specificationsPath := flag.String("specifications", defaultSpecificationsPath,
		"instance specifications file")
	logLevel := flag.String("logLevel", "Info", "log level (Debug, Info, Warn, Error)")
	flag.Parse()

	level := parseLogLevel(*logLevel)
	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
		AddSource: true,
		Level:     level,
	})))

	slog.Debug("Running with flags:")
	flag.VisitAll(func(f *flag.Flag) { slog.Debug("flag", f.Name, f.Value) })

	if err := checkTestSuiteName(*suite); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	if *specificationsPath == defaultSpecificationsPath {
		path := "testdata/" + *suite + "_instance_specifications.json"
		specificationsPath = &path
		slog.Debug("set specification path", "specificationsPath", *specificationsPath)
	}

	var specifications []cover.TestInstanceSpecification
	switch *suite {
	case "tiny":
		specifications = createTinySpecifications(*outputDir)
	}

	b, err := json.MarshalIndent(specifications, "", "  ")
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	if err := os.WriteFile(*specificationsPath, b, 0600); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	createInstanceFiles(specifications)
	solveInstances(specifications, *pythonSolverPath)
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

func createTinySpecifications(outputDir string) []cover.TestInstanceSpecification {
	specifications := make([]cover.TestInstanceSpecification, 0)
	numberOfElements := []int{1, 2, 3, 4}
	costScales := []float64{1.0, 1000.0}

	for _, costScale := range costScales {
		var seed int64
		for _, m := range numberOfElements {
			maxN := int(3*math.Exp2(float64(m))) / 4
			for n := 1; n <= maxN; n++ {
				// two instances for every (m, n)
				for j := 0; j < 2; j++ {
					instanceName := fmt.Sprintf("instance_%d_%d_%d_%d.json", m, n, int(costScale), seed)
					instancePath := filepath.Join(outputDir, instanceName)

					solutionFileName := fmt.Sprintf("python_solution_%d_%d_%d_%d.json", m, n, int(costScale), seed)
					solutionPath := filepath.Join(outputDir, solutionFileName)

					specifications = append(specifications,
						cover.TestInstanceSpecification{
							NumElements:        m,
							NumSubSets:         n,
							CostScale:          costScale,
							Seed:               seed,
							InstancePath:       instancePath,
							PythonSolutionPath: solutionPath,
						})
					slog.Debug("generating instance", "elements count", m, "subsets count", n, "random seed", seed)
					seed++
				}
			}
		}
	}
	return specifications
}

func createInstanceFiles(specifications []cover.TestInstanceSpecification) {
	slog.Info("Creating test instance files", "count", len(specifications))
	for _, spec := range specifications {
		ins := cover.MakeRandomInstance(spec.NumElements, spec.NumSubSets, spec.CostScale, spec.Seed)
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
func solveInstances(specifications []cover.TestInstanceSpecification, pythonSolverPath string) {
	slog.Info("Solving test instances", "count", len(specifications))
	// TODO: introduce "work pools" so we don't start so many Python processes
	var wg sync.WaitGroup
	for _, spec := range specifications {
		wg.Add(1)
		instancePath, solutionPath := spec.InstancePath, spec.PythonSolutionPath
		go func() {
			defer wg.Done()

			cmd := exec.Command("python", pythonSolverPath, instancePath)
			slog.Debug("running", "cmd", cmd)
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

func checkTestSuiteName(names string) error {
	suites := []string{"tiny"}
	for _, s := range suites {
		if names == s {
			return nil
		}
	}

	return fmt.Errorf("unknown test suite name %s", names)
}
