package solvers

import (
	"fmt"
)

type instance struct {
	// The number of elements in the set X to be partitioned, indexed
	// 0 ... n-1.
	n int
	// subsets of X. The inner slices must only contain element indices in
	// [0, n-1]. The indices must be sorted and each subset must be include
	// at most once. Empty subsets are not allowed.
	subsets [][]int
	// The cost of each subset. Each cost must be strictly positive.
	// The length of subsets and costs must be equal.
	// The restrictions on the costs reasonable for many problems and
	// suit certain algorithms.
	costs []float64
}

// Make an Instance and check the constraints that an Instance should satisfy.
func MakeInstance(n int, subsets [][]int, costs []float64) instance {
	if n < 0 {
		panic(fmt.Sprintf(
			"The number of elements n must be nonnegative. %d was supplied.", n))
	} else if n == 0 {
		if len(subsets) != 0 && len(costs) != 0 {
			panic(
				"When the set is empty (n=0), then both subsets and costs must be empty.")
		}
		return instance{}

	}

	for i, subset := range subsets {
		if len(subset) == 0 {
			panic(fmt.Sprintf(
				"Subset index %d was empty. Empty subsets are not allowed.", i))
		}

		for _, element := range subset {
			if element < 0 || element >= n {
				panic(fmt.Sprintf(
					"The subset %v with index %d is invalid since it contains element %d which is not a member of [0, %d).",
					subset, i, element, n))
			}
		}

		usedElements := make(map[int]bool)
		for _, element := range subset {
			usedElements[element] = true
		}
		if len(subset) > len(usedElements) {
			panic(fmt.Sprintf(
				"The subset %v with index %d is invalid since it contains duplicate elements.",
				subset, i))

		}

		prevElement := subset[0]
		for _, element := range subset[1:] {
			if prevElement > element {
				panic(fmt.Sprintf(
					"The subset %v with index %d is invalid since it is not sorted",
					subset, i))

			}
			prevElement = element
		}

	}

	if len(subsets) != len(costs) {
		panic("There must be exactly one cost per subset.")
	}

	for i, cost := range costs {
		if cost <= 0 {
			panic(fmt.Sprintf(
				"The cost %f with index %d is invalid since it is only (strictly) positive costs are supported.",
				cost, i))

		}
	}

	return instance{n: n, subsets: subsets, costs: costs}
}

// Subsets with an evaluation of them w.r.t. some instance.
type subsetsEval struct {
	subsetsIndices []int
	// For some instance, do the subsets exactly cover each element.
	exactlyCovered bool
	// The sum of the subsets' costs.
	cost float64
}
