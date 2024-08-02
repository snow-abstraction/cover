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
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/snow-abstraction/cover/internal/tree"
	"gotest.tools/v3/assert"
)

func TestCreateSubproblemOnEmptyInstance(t *testing.T) {
	ins, err := MakeInstance(0, [][]int{}, []float64{})
	assert.NilError(t, err)
	subproblemIns := createSubproblem(ins, tree.CreateRoot())
	assert.DeepEqual(t, *subproblemIns, ins, cmp.AllowUnexported(instance{}))
}

func TestCreateSubproblemCreatesInfeasibleBranch(t *testing.T) {
	ins, err := MakeInstance(2, [][]int{{0, 1}, {0}}, []float64{1, 2})
	assert.NilError(t, err)
	rootNode := tree.CreateRoot()
	_, diffNode := rootNode.Branch(0, 0, 1)

	actualDiffIns := createSubproblem(ins, diffNode)
	assert.Assert(t, actualDiffIns == nil)
}

func TestCreateSubproblemsSimple(t *testing.T) {
	ins, err := MakeInstance(3, [][]int{{0, 1}, {0}, {1}, {2}}, []float64{1, 2, 3, 4})
	assert.NilError(t, err)
	rootNode := tree.CreateRoot()
	bothNode, diffNode := rootNode.Branch(0, 0, 1)

	actualBothIns := createSubproblem(ins, bothNode)
	expectedBothIns, err := MakeInstance(3, [][]int{{0, 1}, {2}}, []float64{1, 4})
	assert.NilError(t, err)
	assert.DeepEqual(t, expectedBothIns, *actualBothIns, cmp.AllowUnexported(instance{}))

	actualDiffIns := createSubproblem(ins, diffNode)
	expectedDiffIns, err := MakeInstance(3, [][]int{{0}, {1}, {2}}, []float64{2, 3, 4})
	assert.NilError(t, err)
	assert.DeepEqual(t, expectedDiffIns, *actualDiffIns, cmp.AllowUnexported(instance{}))
}

func TestCreateSubproblemsTricker(t *testing.T) {
	ins, err := MakeInstance(
		3,
		[][]int{{0}, {1}, {2}, {0, 1}, {0, 2}, {1, 2}, {0, 1, 2}},
		[]float64{1, 2, 3, 4, 5, 6, 7})
	assert.NilError(t, err)
	rootNode := tree.CreateRoot()
	both01Node, diff01Node := rootNode.Branch(0, 0, 1)
	both01Both12Node, both01Diff12Node := both01Node.Branch(0, 1, 2)
	diff01both12Node, diff01Diff12Node := diff01Node.Branch(0, 1, 2)

	actualBoth01Both12NodeIns := createSubproblem(ins, both01Both12Node)
	expectedBoth01Both12NodeIns, err := MakeInstance(3, [][]int{{0, 1, 2}}, []float64{7})
	assert.NilError(t, err)
	assert.DeepEqual(t, expectedBoth01Both12NodeIns, *actualBoth01Both12NodeIns, cmp.AllowUnexported(instance{}))

	actualBoth01Diff12NodeIns := createSubproblem(ins, both01Diff12Node)
	expectedBoth01Diff12NodeIns, err := MakeInstance(3, [][]int{{2}, {0, 1}}, []float64{3, 4})
	assert.NilError(t, err)
	assert.DeepEqual(t, expectedBoth01Diff12NodeIns, *actualBoth01Diff12NodeIns, cmp.AllowUnexported(instance{}))

	actualDiff01bBoth12NodeIns := createSubproblem(ins, diff01both12Node)
	expectedDiff01Both12NodeIns, err := MakeInstance(3, [][]int{{0}, {1, 2}}, []float64{1, 6})
	assert.NilError(t, err)
	assert.DeepEqual(t, expectedDiff01Both12NodeIns, *actualDiff01bBoth12NodeIns, cmp.AllowUnexported(instance{}))

	actualDiff01Diff12NodeIns := createSubproblem(ins, diff01Diff12Node)
	expectedDiff01Diff12NodeIns, err := MakeInstance(3, [][]int{{0}, {1}, {2}, {0, 2}}, []float64{1, 2, 3, 5})
	assert.NilError(t, err)
	assert.DeepEqual(t, expectedDiff01Diff12NodeIns, *actualDiff01Diff12NodeIns, cmp.AllowUnexported(instance{}))
}
