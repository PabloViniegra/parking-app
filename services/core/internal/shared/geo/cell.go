package geo

import (
	"fmt"

	h3 "github.com/uber/h3-go/v4"
)

func CellR10(latitude, longitude float64) (int64, error) {
	cell, err := h3.LatLngToCell(h3.NewLatLng(latitude, longitude), 10)
	if err != nil {
		return 0, fmt.Errorf("calculate H3 cell: %w", err)
	}
	return int64(cell), nil
}
