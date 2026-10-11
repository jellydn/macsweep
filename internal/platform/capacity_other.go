//go:build !darwin && !linux

package platform

import "errors"

func Volume(string) (Capacity, error) {
	return Capacity{}, errors.New("volume capacity is unavailable on this platform")
}
