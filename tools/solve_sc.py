"""Solve exact set covering problem instance.

This a bare bones solver for the exact set covering problem in order
to test solvers in Golang. It reads the supplied instance file
and prints solution to the stdout.

Usage: python solve_sc instance_file.json


License:
 Copyright (C) 2023 Douglas Wayne Potter

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
"""
import sys
import json
import mip


def convert_optimization_status_to_string(status):
    if status == mip.OptimizationStatus.OPTIMAL:
        return "optimal"
    elif status in (
        mip.OptimizationStatus.INFEASIBLE,
        mip.OptimizationStatus.INT_INFEASIBLE,
    ):
        return "infeasible"

    return "unhandled OptimizationStatus"


if __name__ == "__main__":
    with open(sys.argv[1]) as f:
        instance = json.load(f)

    model = mip.Model()
    model.verbose = 0

    x_vars = [
        model.add_var(var_type=mip.BINARY, obj=cost, lb=0.0, ub=1.0)
        for cost in instance["Costs"]
    ]

    constraints = [[] for _ in range(instance["N"])]
    for var_index, subset in enumerate(instance["Subsets"]):
        for constr_index in subset:
            constraints[constr_index].append(x_vars[var_index])

    # It does not seem possible to add row with empty lhs. Thus
    # instances with elements that are not in any subsets won't
    # be represented correctly.
    all_nonempty_lhs = True
    for constraint in constraints:
        if not constraint:
            all_nonempty_lhs = False
            break
        model += mip.xsum(constraint) == 1

    results = {}
    status = "infeasible"
    if all_nonempty_lhs:
        status = convert_optimization_status_to_string(model.optimize())
    sys.stdout.flush()
    results["status"] = status

    if status == "optimal":
        results["cost"] = model.objective_value
        results["solution"] = [i for (i, var) in enumerate(x_vars) if var.x > 0.99]

    print("solve_sc_result:", json.dumps(results))
