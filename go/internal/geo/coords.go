package geo

import (
	"fmt"
	"math"
	"strconv"
)

// ParseCoords validates lat/lon and rounds to 2 decimals (~1 km) so cache
// entries are shared between nearby visitors.
func ParseCoords(latStr, lonStr string) (lat, lon string, ok bool) {
	latF, latErr := strconv.ParseFloat(latStr, 64)
	lonF, lonErr := strconv.ParseFloat(lonStr, 64)

	if latErr != nil || lonErr != nil ||
		latF < -90 || latF > 90 ||
		lonF < -180 || lonF > 180 {
		return "", "", false
	}

	lat = fmt.Sprintf("%.2f", math.Round(latF*100)/100)
	lon = fmt.Sprintf("%.2f", math.Round(lonF*100)/100)
	return lat, lon, true
}
