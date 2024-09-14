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

// bb = branch-and-bound

package solvers

import (
	"encoding/json"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/snow-abstraction/cover"
	"github.com/snow-abstraction/cover/internal/tree"
	"gotest.tools/v3/assert"
)

func TestCreateSubInstanceOnEmptyInstance(t *testing.T) {
	ins, err := MakeInstance(0, [][]int{}, []float64{})
	assert.NilError(t, err)
	subproblemIns := createSubInstance(ins, tree.CreateRoot())
	assert.DeepEqual(t, subproblemIns.ins, ins, cmp.AllowUnexported(instance{}))
}

func TestCreateSubInstanceCreatesInfeasibleBranch(t *testing.T) {
	ins, err := MakeInstance(2, [][]int{{0, 1}, {0}}, []float64{1, 2})
	assert.NilError(t, err)
	rootNode := tree.CreateRoot()
	_, diffNode := rootNode.Branch(0, 0, 1)

	actualDiffIns := createSubInstance(ins, diffNode)
	assert.Assert(t, actualDiffIns == nil)
}

func TestCreateSubInstancesSimple(t *testing.T) {
	ins, err := MakeInstance(3, [][]int{{0, 1}, {0}, {1}, {2}}, []float64{1, 2, 3, 4})
	assert.NilError(t, err)
	rootNode := tree.CreateRoot()
	bothNode, diffNode := rootNode.Branch(0, 0, 1)

	actualBothIns := createSubInstance(ins, bothNode)
	expectedBothIns, err := MakeInstance(3, [][]int{{0, 1}, {2}}, []float64{1, 4})
	assert.NilError(t, err)
	assert.DeepEqual(t, expectedBothIns, actualBothIns.ins, cmp.AllowUnexported(instance{}))

	actualDiffIns := createSubInstance(ins, diffNode)
	expectedDiffIns, err := MakeInstance(3, [][]int{{0}, {1}, {2}}, []float64{2, 3, 4})
	assert.NilError(t, err)
	assert.DeepEqual(t, expectedDiffIns, actualDiffIns.ins, cmp.AllowUnexported(instance{}))
}

func TestCreateSubInstancesTricker(t *testing.T) {
	ins, err := MakeInstance(
		3,
		[][]int{{0}, {1}, {2}, {0, 1}, {0, 2}, {1, 2}, {0, 1, 2}},
		[]float64{1, 2, 3, 4, 5, 6, 7})
	assert.NilError(t, err)
	rootNode := tree.CreateRoot()
	both01Node, diff01Node := rootNode.Branch(0, 0, 1)
	both01Both12Node, both01Diff12Node := both01Node.Branch(0, 1, 2)
	diff01both12Node, diff01Diff12Node := diff01Node.Branch(0, 1, 2)

	actualBoth01Both12NodeIns := createSubInstance(ins, both01Both12Node)
	expectedBoth01Both12NodeIns, err := MakeInstance(3, [][]int{{0, 1, 2}}, []float64{7})
	assert.NilError(t, err)
	assert.DeepEqual(t, expectedBoth01Both12NodeIns, actualBoth01Both12NodeIns.ins, cmp.AllowUnexported(instance{}))

	actualBoth01Diff12NodeIns := createSubInstance(ins, both01Diff12Node)
	expectedBoth01Diff12NodeIns, err := MakeInstance(3, [][]int{{2}, {0, 1}}, []float64{3, 4})
	assert.NilError(t, err)
	assert.DeepEqual(t, expectedBoth01Diff12NodeIns, actualBoth01Diff12NodeIns.ins, cmp.AllowUnexported(instance{}))

	actualDiff01bBoth12NodeIns := createSubInstance(ins, diff01both12Node)
	expectedDiff01Both12NodeIns, err := MakeInstance(3, [][]int{{0}, {1, 2}}, []float64{1, 6})
	assert.NilError(t, err)
	assert.DeepEqual(t, expectedDiff01Both12NodeIns, actualDiff01bBoth12NodeIns.ins, cmp.AllowUnexported(instance{}))

	actualDiff01Diff12NodeIns := createSubInstance(ins, diff01Diff12Node)
	expectedDiff01Diff12NodeIns, err := MakeInstance(3, [][]int{{0}, {1}, {2}, {0, 2}}, []float64{1, 2, 3, 5})
	assert.NilError(t, err)
	assert.DeepEqual(t, expectedDiff01Diff12NodeIns, actualDiff01Diff12NodeIns.ins, cmp.AllowUnexported(instance{}))
}

func testBBFindsEquallyGoodSolution(t *testing.T, spec cover.TestInstanceSpecification) {
	pythonResultBytes, err := os.ReadFile(filepath.Join("../..", spec.PythonSolutionPath))
	assert.NilError(t, err)
	var pythonResult map[string]interface{}
	err = json.Unmarshal(pythonResultBytes, &pythonResult)
	assert.NilError(t, err)

	instanceBytes, err := os.ReadFile(filepath.Join("../..", spec.InstancePath))
	assert.NilError(t, err)
	var ins cover.Instance
	err = json.Unmarshal(instanceBytes, &ins)
	assert.NilError(t, err)
	solverInstance, err := MakeInstance(ins.M, ins.Subsets, ins.Costs)
	assert.NilError(t, err)

	result, err := SolveByBranchAndBound(solverInstance)
	assert.NilError(t, err)

	// This is tightly coupled to JSON format of tools/solve_sc.py.
	if result.ExactlyCovered {
		assert.Equal(t, "optimal", pythonResult["status"].(string))
		pythonCost := pythonResult["cost"].(float64)
		costDiff := math.Abs(result.Cost - pythonCost)
		assert.Assert(
			t,
			costDiff < 0.000000000001,
			"brute found an optimal cost %f but the Python script found an optimal cost %f ",
			result.Cost,
			pythonCost,
		)
	} else {
		assert.Equal(
			t,
			"infeasible",
			pythonResult["status"].(string),
			"was found infeasible by brute but not by the Python script",
		)
	}

}

func TestBBOnInstances(t *testing.T) {
	instanceSpecifications := loadInstanceSpecifications(t)

	for _, spec := range instanceSpecifications {
		spec := spec
		name := fmt.Sprintf("instance %+v", spec)
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			testBBFindsEquallyGoodSolution(t, spec)
		})

	}
}

func BenchmarkBBOnRandomInstances(b *testing.B) {
	instanceSpecifications := loadInstanceSpecifications(b)

	instances := make([]instance, 0, len(instanceSpecifications))
	for _, spec := range instanceSpecifications {
		instanceBytes, err := os.ReadFile(filepath.Join("../..", spec.InstancePath))
		assert.NilError(b, err)
		var ins cover.Instance
		err = json.Unmarshal(instanceBytes, &ins)
		assert.NilError(b, err)
		solverInstance, err := MakeInstance(ins.M, ins.Subsets, ins.Costs)
		instances = append(instances, solverInstance)
		assert.NilError(b, err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for j := 0; j < len(instances); j++ {
			_, err := SolveByBranchAndBound(instances[j])
			assert.NilError(b, err)
		}
	}
}

func BenchmarkBBOnRandomScale1Instances(b *testing.B) {
	instanceSpecifications := loadInstanceSpecifications(b)

	instances := make([]instance, 0, len(instanceSpecifications))
	for _, spec := range instanceSpecifications {
		if spec.CostScale != 1.0 {
			continue
		}
		instanceBytes, err := os.ReadFile(filepath.Join("../..", spec.InstancePath))
		assert.NilError(b, err)
		var ins cover.Instance
		err = json.Unmarshal(instanceBytes, &ins)
		assert.NilError(b, err)
		solverInstance, err := MakeInstance(ins.M, ins.Subsets, ins.Costs)
		instances = append(instances, solverInstance)
		assert.NilError(b, err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for j := 0; j < len(instances); j++ {
			_, err := SolveByBranchAndBound(instances[j])
			assert.NilError(b, err)
		}
	}
}

func BenchmarkBBOnRandomScale1000Instances(b *testing.B) {
	instanceSpecifications := loadInstanceSpecifications(b)

	instances := make([]instance, 0, len(instanceSpecifications))
	for _, spec := range instanceSpecifications {
		if spec.CostScale != 1000.0 {
			continue
		}
		instanceBytes, err := os.ReadFile(filepath.Join("../..", spec.InstancePath))
		assert.NilError(b, err)
		var ins cover.Instance
		err = json.Unmarshal(instanceBytes, &ins)
		assert.NilError(b, err)
		solverInstance, err := MakeInstance(ins.M, ins.Subsets, ins.Costs)
		instances = append(instances, solverInstance)
		assert.NilError(b, err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for j := 0; j < len(instances); j++ {
			_, err := SolveByBranchAndBound(instances[j])
			assert.NilError(b, err)
		}
	}
}
