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
package solvers

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/snow-abstraction/cover"
	"gotest.tools/v3/assert"
)

func TestLowerBoundCalcOnTinyInstances(t *testing.T) {
	instanceSpecifications := loadTinyInstanceSpecifications(t)

	for _, spec := range instanceSpecifications {
		spec := spec
		name := fmt.Sprintf("instance %+v", spec)
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			testLowerBound(t, spec)
		})

	}
}

func testLowerBound(t *testing.T, spec cover.TestInstanceSpecification) {
	pythonResultBytes, err := os.ReadFile(filepath.Join("../..", spec.PythonSolutionPath))
	assert.NilError(t, err)
	var pythonResult map[string]interface{}
	err = json.Unmarshal(pythonResultBytes, &pythonResult)
	assert.NilError(t, err)

	// Only check lower bound for feasible instances
	if pythonResult["status"].(string) == "infeasible" {
		return
	}
	assert.Equal(t, "optimal", pythonResult["status"].(string))

	instanceFile := filepath.Join("../..", spec.InstancePath)
	ins, err := cover.ReadJsonInstance(instanceFile)
	assert.NilError(t, err)

	m, err := convertSubsetsToMatrix(ins.Subsets)
	assert.NilError(t, err)

	lowerBound, err := CalcScLb(m, ins.Costs)
	assert.NilError(t, err)

	pythonCost := pythonResult["cost"].(float64)

	// Due to floating point, use lowerBound-pythonCost <= 1e-10 to check lowerBound <= pythonCost
	// This is a hack and more sophisticated check might be needed later.
	assert.Assert(t, lowerBound-pythonCost <= 1e-10, "%f <= %f is false", lowerBound, pythonCost)
}
