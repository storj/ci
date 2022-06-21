// Copyright (C) 2022 Storj Labs, Inc.
// See LICENSE for copying information.

package gerrit

// LabelMax finds the maximum value of a label (eg. what is the highest Verified label?)
func (c Change) LabelMax(name string) int {
	res := 0
	for label, values := range c.Labels {
		if label == name {
			for _, vote := range values.All {
				if vote.Value > res {
					res = vote.Value
				}
			}
		}
	}
	return res
}

// LabelMin finds the minimum value of a label (eg. what is the lowest Code-Review feedback?)
func (c Change) LabelMin(name string) int {
	res := 0
	for label, values := range c.Labels {
		if label == name {
			for _, vote := range values.All {
				if vote.Value < res {
					res = vote.Value
				}
			}
		}
	}
	return res
}

// LabelCount counts the number of labels with specific value (eg. Verified=+1).
func (c Change) LabelCount(name string, value int) int {
	res := 0
	for label, values := range c.Labels {
		if label == name {
			for _, vote := range values.All {
				if vote.Value == value {
					res++
				}
			}
		}
	}
	return res
}
