/*
 Copyright (C) 2022 Douglas Wayne Potter

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

package math

import (
	"reflect"
	"testing"
)

func TestConvert(t *testing.T) {
	ccs := cCSMatrix{0, 1, 2, sen, sen, 1, sen}
	crs := ccs.Convert()
	want := cRSMatrix{0, sen, 0, 2, sen, 0, sen}
	if !reflect.DeepEqual(crs, want) {
		t.Fatalf("\n%v !=\n%v", crs, want)
	}

	ccs_again := crs.Convert()
	if !reflect.DeepEqual(ccs_again, ccs) {
		t.Fatalf("\n%v !=\n%v", ccs_again, ccs)
	}

}

func TestInvalid(t *testing.T) {
	l := 12
	data := make([]uint32, l)
	for i := 0; i < l; i++ {
		data[i] = sen
	}

	if _, err := makeCompressedRowMatrix(data); err == nil {
		t.Fatalf("Should be invalid.")
	}

	if _, err := makeCompressedColumnMatrix(data); err == nil {
		t.Fatalf("Should be invalid.")
	}

}

func TestMatrixVectorMultiply(t *testing.T) {
	tests := []struct {
		matrix []uint32
		vector []float64
		want   []float64
	}{
		{
			matrix: []uint32{0, sen, 1, sen, 2, sen},
			vector: []float64{1, 2, 3},
			want:   []float64{1, 2, 3},
		},
		{
			matrix: []uint32{0, sen, 0, 1, sen, 0, 1, 2, sen},
			vector: []float64{1, 1, 1},
			want:   []float64{1, 2, 3},
		},
		{
			matrix: []uint32{0, sen, 0, 1, sen, 0, 1, 2, sen},
			vector: []float64{1, -0.5, 3},
			want:   []float64{1, 0.5, 3.5},
		},
		{
			matrix: []uint32{0, sen, 0, 1, sen, sen, sen},
			vector: []float64{1, -0.5, 3, 4},
			want:   []float64{1, 0.5, 0, 0},
		},
	}

	for _, tc := range tests {
		m, _ := makeCompressedRowMatrix(tc.matrix)
		got := make([]float64, len(tc.want))
		m.MatrixVectorMultiply(tc.vector, got)
		if !reflect.DeepEqual(got, tc.want) {
			t.Fatalf("\n%v !=\n%v", got, tc.want)
		}
	}
}
