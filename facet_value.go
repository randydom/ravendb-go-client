package ravendb

import (
	"fmt"
	"strconv"
	"strings"
)

type FacetValue struct {
	Range   string
	Count   int
	Sum     *float64
	Max     *float64
	Min     *float64
	Average *float64
}

func (v *FacetValue) SetSum(sum float64) {
	v.Sum = &sum
}

func (v *FacetValue) SetMax(max float64) {
	v.Max = &max
}

func (v *FacetValue) SetMin(min float64) {
	v.Min = &min
}

func (v *FacetValue) SetAverage(average float64) {
	v.Average = &average
}

func (v *FacetValue) String() string {
	msg := v.Range + " - Count: " + strconv.Itoa(v.Count) + ", "
	if v.Sum != nil {
		msg += fmt.Sprintf("Sum: %f,", *v.Sum)
	}
	if v.Max != nil {
		msg += fmt.Sprintf("Max: %f,", *v.Max)
	}
	if v.Min != nil {
		msg += fmt.Sprintf("Min: %f,", *v.Min)
	}
	if v.Average != nil {
		msg += fmt.Sprintf("Average: %f,", *v.Average)
	}

	// TODO: this makes no sense but is in Java code
	return strings.TrimSuffix(msg, ";")
}
