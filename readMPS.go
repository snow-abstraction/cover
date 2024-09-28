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
	"bufio"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"strconv"
	"strings"
)

const (
	MPS_SECTION_NOT_SET int = iota
	MPS_SECTION_NAME
	MPS_SECTION_ROWS
	MPS_SECTION_COLUMNS
	MPS_SECTION_RHS
	MPS_SECTION_BOUNDS
	MPS_SECTION_ENDATA
)

// ReadMPSInstance reads an exact set covering problem from a MPS file.
//
// ReadMPSInstance has neither been tested systematically or programmatically.
// It has been tested successfully on a few exact cover (i.e. setting partition)
// files from the miplib2003 and miplip2010 problem collections.
func ReadMPSInstance(filename *string) (*Instance, error) {
	file, err := os.Open(*filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	ins := Instance{}
	scanner := bufio.NewScanner(file)

	prefix := "MPS reader"
	currentSection := MPS_SECTION_NOT_SET
	foundCostRow := false
	rhsCount := 0
	upperBoundCount := 0
	rows := make(map[string]int, 0)    // row name to row/element index
	columns := make(map[string]int, 0) // column name to column/subset index
	for scanner.Scan() {
		s := scanner.Text()
		if s == "" {
			continue
		} else if strings.HasPrefix(s, "*") {
			slog.Debug(prefix, "comment", s)
			continue
		}

		newSection := !strings.HasPrefix(s, " ") && !strings.HasPrefix(s, "\t")
		if newSection {
			slog.Info(prefix, "start section", s)
			currentSection, err = parseMPSSection(s)
			if err != nil {
				return nil, fmt.Errorf("unsupported MPS section '%s'", s)
			}
			if currentSection == MPS_SECTION_ENDATA {
				break
			}
			continue
		}

		switch currentSection {
		case MPS_SECTION_ROWS:
			if strings.Contains(s, "N") && strings.Contains(s, "COST") {
				foundCostRow = true
				continue
			}
			fields := strings.Fields(s)
			if len(fields) != 2 {
				return nil, fmt.Errorf("ROW entry should contain two fields but found '%s'", s)
			}
			if fields[0] != "E" {
				return nil, fmt.Errorf("only ROW sense E supported but found sense '%s'", fields[0])
			}
			if _, found := rows[fields[1]]; found {
				return nil, fmt.Errorf("ROW name '%s' duplicated", fields[0])
			}
			rows[fields[1]] = len(rows)
			ins.M++
		case MPS_SECTION_COLUMNS:
			if !foundCostRow {
				return nil, fmt.Errorf("expected cost row to be found before COLUMN section")

			}
			if strings.Contains(s, "MARKER") {

				continue
			}
			fields := strings.Fields(s)
			if len(fields) == 0 {
				return nil, fmt.Errorf("expected column name in COLUMN entry but found '%s'", s)
			}
			if len(fields) != 1 && len(fields) != 3 && len(fields) != 5 {
				return nil, fmt.Errorf(
					"expected COLUMN entry a column name with 0, 1 or 2 row entries but found '%s'", s)

			}
			colIdx, found := columns[fields[0]]
			if !found {
				colIdx = len(columns)
				columns[fields[0]] = colIdx
			}

			for len(ins.Subsets) <= colIdx {
				ins.Subsets = append(ins.Subsets, make([]int, 0))
			}
			for len(ins.Costs) <= colIdx {
				ins.Costs = append(ins.Costs, 0)
			}

			for i := 1; i < len(fields); i = i + 2 {
				if fields[i] == "COST" {
					cost, err := strconv.ParseFloat(fields[i+1], 64)
					if err != nil {
						return nil, fmt.Errorf("unable to parse cost '%s' from COLUMN entry '%s'", fields[i+1], s)
					}

					ins.Costs[colIdx] = cost
				} else {
					rowIdx, found := rows[fields[i]]
					if !found {
						return nil, fmt.Errorf("unknown row '%s' in COLUMN entry '%s'", fields[i], s)
					}
					one, err := strconv.ParseFloat(fields[i+1], 64)
					if err != nil {
						return nil, fmt.Errorf("unable to parse expected '1' from '%s' from COLUMN entry '%s'", fields[i+1], s)
					}
					if one != 1.0 {
						return nil, fmt.Errorf("expect all constraint values to be exactly 1.0 in COLUMN entry '%s'", s)
					}
					ins.Subsets[colIdx] = append(ins.Subsets[colIdx], rowIdx)
				}
			}
		case MPS_SECTION_RHS:
			fields := strings.Fields(s)
			if len(fields) != 3 && len(fields) != 5 {
				return nil, fmt.Errorf("expect RHS entry to contain 1 or 2 entries '%s'", s)
			}
			for i := 1; i < len(fields); i = i + 2 {

				_, found := rows[fields[i]]
				if !found {
					return nil, fmt.Errorf("unknown row '%s' in RHS entry '%s'", fields[i], s)
				}
				rhs, err := strconv.ParseFloat(fields[i+1], 64)
				if err != nil {
					return nil, fmt.Errorf("unable to parse rhs '%s' from RHS entry '%s'", fields[i+1], s)
				}
				if rhs != 1.0 {
					return nil, fmt.Errorf("expect all rhs values to be exactly 1.0 in RHS entry '%s'", s)
				}
				rhsCount++
			}
		case MPS_SECTION_BOUNDS:

			fields := strings.Fields(s)
			errMsg := fmt.Sprintf("expect UP BND followed by column name and 1 for all BOUND entries but found '%s'", s)
			if len(fields) != 4 {
				return nil, errors.New(errMsg)
			}
			up := strings.ToUpper(fields[0])
			if fields[0] != up {
				return nil, errors.New(errMsg)
			}
			bnd := strings.ToUpper(fields[1])
			if bnd != "BND" {
				return nil, errors.New(errMsg)
			}
			_, found := columns[fields[2]]
			if !found {
				return nil, fmt.Errorf("unknown column '%s' in BOUNDS entry '%s'", fields[2], s)
			}
			upperBound, err := strconv.ParseFloat(fields[3], 64)
			if err != nil {
				return nil, fmt.Errorf(
					"unable to parse upper bound '%s' for column '%s' from BOUNDS entry '%s'",
					fields[3], fields[2], s)
			}
			if upperBound != 1.0 {
				return nil, fmt.Errorf("expect all upper bounds values to be exactly 1.0 in RHS entry '%s'", s)
			}
			upperBoundCount++

		default:
			return nil, fmt.Errorf("mps section processing error")
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}
	if rhsCount != ins.M {
		return nil, fmt.Errorf("rhsCount (%d) != ins.m (%d)", rhsCount, ins.M)
	}
	if upperBoundCount != len(ins.Subsets) {
		return nil, fmt.Errorf("upperBoundCount (%d) != len(ins.Subsets) (%d)", upperBoundCount, len(ins.Subsets))
	}

	return &ins, nil
}

func parseMPSSection(s string) (int, error) {
	// TODO: check sections occur in expected order
	switch {
	case strings.HasPrefix(s, "NAME"):
		return MPS_SECTION_NAME, nil
	case strings.HasPrefix(s, "ROWS"):
		return MPS_SECTION_ROWS, nil
	case strings.HasPrefix(s, "COLUMNS"):
		return MPS_SECTION_COLUMNS, nil
	case strings.HasPrefix(s, "RHS"):
		return MPS_SECTION_RHS, nil
	case strings.HasPrefix(s, "BOUNDS"):
		return MPS_SECTION_BOUNDS, nil
	case strings.HasPrefix(s, "ENDATA"):
		return MPS_SECTION_ENDATA, nil
	}
	return MPS_SECTION_NOT_SET, fmt.Errorf("unsupported MPS section '%s'", s)
}
