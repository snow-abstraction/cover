# cover

Using the SCP "Set Cover Problem" as a context to play with Go.

The initial plan is to mix some algorithms to build a solver for
weighted covering problems with strictly positive costs. I haven't
decided if the focus will be on general covers or exact covers (also known
as the "set partitioning problem").

# License

This project is under [AGPL-3.0-only](LICENSE) license.

# Example



```go
import (
	"testing"

	"gotest.tools/v3/assert"

	"github.com/snow-abstraction/cover"
	"github.com/snow-abstraction/cover/solvers"
)

func TestReadMeExample(t *testing.T) {
	instance := cover.Instance{
		ElementCount: 4, // This means set of elements to cover is {0, 1, 2 and 3}
		Subsets:      [][]int{{0}, {0, 1}, {1, 2}, {1}, {0, 1, 2, 3}, {2, 3}, {0, 1, 3}, {2}},
		Costs:        []float64{1.8, 1.7, 2.4, 1.4, 5.4, 2.7, 1.9, 1.6}}

	result, err := solvers.SolveByBranchAndBound(instance)
	assert.NilError(t, err)
	assert.Assert(t, result.Optimal)
	assert.Equal(t, result.Cost, 3.5)
	assert.DeepEqual(t, result.SubsetsIndices, []int{6, 7})
}
```

This example shows a trivial instance with 4 elements where the two last subsets (indices
6 and 7) are an optimal exact cover with total cost 3.5. This example is tested
[here](internal/doctest/doc_test.go).

# Dev Note

While this is a Go project, a Python program is used to generate test data.
This program independently solves SCP instances so we can verify that equally
good solutions are found by our code. For my Ubuntu system, here is a simple
way to get started:

```
cd tools
sudo apt install libffi-dev # install requirement needed by next line
pip install -i requirements.txt
cd ..
go run cmd/generate_test_instances_and_solutions/main.go -verbose
```

(A less hacky setup would be use a container or Python virtual environment.)

# Bigger TODOs / Project Ideas

- [ ] better branching
- [ ] better step length, use some upper bound to calculate?
- [ ] smart subgradient iteration termination criteria instead of only detecting zero
      subgradient or fixed iteration limit
- [ ] parallelization
- [ ] smart warm starts. Naive warm starts did not improve performance. 
      These warmed started using the last dual vector found from the previously
	  processed node. Maybe the result would be better if the dual vector was
	  from a close ancestor node.
- [ ] in the Lagrangian relaxation, exploit that only m columns can be chosen
      in a primal feasible solution
- [ ] visualize the branch-and-bound tree
- [ ] support relative and absolute optimality gap termination criteria

# Project Note

As of August 2023, I have focused little on Go and thus I am unsure of the
point of this project since then. The primary reason for this project was to
become better at Go for a project at work. But that project has been paused
and my focus returned a Java code base.

Both algorithm and software engineering choices are motivated mainly
by what is fun as a hobby project for me.
