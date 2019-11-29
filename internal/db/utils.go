package db

import "strconv"

// IsMacLocallyAdministered expects the mac in the format e.g. "20:c9:d0:7a:fa:31"
// https://en.wikipedia.org/wiki/MAC_address
func IsMacLocallyAdministered(mac string) bool {
	// 00000010
	const mask = 1 << 1

	first2chars := mac[:2]
	decimal, _ := strconv.ParseInt(first2chars, 16, 64)
	return (decimal & mask) == mask
}
