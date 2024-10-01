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
	"os"
	"testing"

	"github.com/snow-abstraction/cover"
	"gotest.tools/v3/assert"
)

func loadTinyInstanceSpecifications(t testing.TB) []cover.TestInstanceSpecification {
	var result []cover.TestInstanceSpecification
	b, err := os.ReadFile("../../testdata/tiny_instance_specifications.json")
	assert.NilError(t, err)
	err = json.Unmarshal(b, &result)
	assert.NilError(t, err)
	return result
}

func loadSmallInstanceSpecifications(t testing.TB) []cover.TestInstanceSpecification {
	var result []cover.TestInstanceSpecification
	b, err := os.ReadFile("../../testdata/small_instance_specifications.json")
	assert.NilError(t, err)
	err = json.Unmarshal(b, &result)
	assert.NilError(t, err)
	return result
}

func loadSolverInstance(t testing.TB, jsonInstancePath string) instance {
	instanceBytes, err := os.ReadFile(jsonInstancePath)
	assert.NilError(t, err)
	var ins cover.Instance
	err = json.Unmarshal(instanceBytes, &ins)
	assert.NilError(t, err)
	solverInstance, err := MakeInstance(ins.M, ins.Subsets, ins.Costs)
	assert.NilError(t, err)
	return solverInstance
}
