package utils

import (
	"math"
	"strconv"
	"strings"
)

// ParseDuration parses a string in the format "[dd:mm]mm:ss" and returns the duration in seconds
func ParseDuration(s string) (uint32, error) {
	var seconds uint32
	slices := strings.Split(s, ":")
	multiplier := math.Pow(60, float64(len(slices)-1)) // stores whether we are parsing days, hours, ...
	for _, v := range slices {
		if n, err := strconv.Atoi(v); err != nil {
			return 0, err
		} else {
			seconds += uint32(n) * uint32(multiplier) // add the parsed number to the total
			multiplier /= 60                          // move to the next unit
		}
	}
	return seconds, nil
}
