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
	"testing"

	"github.com/snow-abstraction/cover"
	"gotest.tools/v3/assert"
)

func BenchmarkMakeInstance(b *testing.B) {
	seed := int64(1)
	ins := cover.MakeRandomInstance(3000, 20000, seed)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {

		_, err := MakeInstance(ins.M, ins.Subsets, ins.Costs)
		assert.NilError(b, err)
	}
}
