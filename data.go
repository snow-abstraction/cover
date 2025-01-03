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

package cover

import (
	"encoding/json"
	"math"
	"math/rand"
	"os"
	"slices"
	"sort"
)

type Instance struct {
	// The number of elements in the set X to be covered, indexed
	// 0 ... ElementCount-1.
	ElementCount int
	// Subsets of X. The inner slices must only contain element indices in
	// [0, M-1]. The indices must be sorted and each subset must be include
	// at most once. Empty Subsets are not allowed.
	Subsets [][]int
	// The cost of each subset. Each cost must be strictly positive.
	// The length of subsets and Costs must be equal.
	// The restrictions on the Costs reasonable for many problems and
	// suit certain algorithms.
	Costs []float64
}

// MakeRandomInstance makes a random Instance with m elements and n subsets
// using a PRG (Pseudo-Random Generator) initialized with the seed. The random
// cost of each subset is scaled by costScale.
//
// Note: as n approaches 2^m, this function will run slowly.
func MakeRandomInstance(m int, n int, costScale float64, seed int64) Instance {
	gen := rand.New(rand.NewSource(seed))

	ins := Instance{ElementCount: m, Subsets: make([][]int, 0), Costs: make([]float64, 0)}

	// universe of elements to be covered
	u := make([]int, ins.ElementCount)
	for i := 0; i < ins.ElementCount; i++ {
		u[i] = i
	}
	for j := 0; j < n; j++ {
		for {
			// make random subset u[:k]
			gen.Shuffle(len(u), func(i, j int) { u[i], u[j] = u[j], u[i] })
			k := gen.Intn(ins.ElementCount) + 1
			subset := u[:k]
			// sort subset to give it an unique representation
			sort.Ints(subset)

			// only add subset if unique
			// TODO: This introduces quadratic complexity. Ideally we would
			// do a binary search or use some hash table to check if the
			// subset has already been added.
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

				// generate random cost such that:
				// costScale < cost <= costScale*k^smallSubsetPreference + 1.
				// This biases instances where optimal solutions
				// consist of several small subsets.
				const smallSubsetPreference = 1.1
				f := math.Pow(float64(k), smallSubsetPreference) + 1
				s := gen.Float64()
				ins.Costs = append(ins.Costs, costScale*(f*(1-s)+s))
				break
			}
		}

	}

	return ins
}

func ReadJsonInstance(filename string) (*Instance, error) {
	b, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	var ins Instance
	if err := json.Unmarshal(b, &ins); err != nil {
		return nil, err
	}
	return &ins, nil
}

// Subsets with an evaluation of them w.r.t. some instance.
type SubsetsEval struct {
	SubsetsIndices []int
	// For the instance, do the subsets exactly cover each element.
	// If false, the subsets either undercover or overcover the set.
	ExactlyCovered bool
	// The sum of the subsets' costs.
	Cost float64
	// If the SubsetsIndices constitute a proven optimum. This can only be true if
	// ExactlyCovered is true.
	Optimal bool
}

// For specifying data related to test instances
type TestInstanceSpecification struct {
	NumElements        int
	NumSubSets         int
	CostScale          float64
	Seed               int64 // random seed used to generate instance
	InstancePath       string
	PythonSolutionPath string
}
