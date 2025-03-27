package memlimit

import (
	"errors"
)

const checkerName = "mem_limit_checker"

var ErrLowMemory = errors.New("low memory")

type MemoryGetter interface {
	GetUsedBytes() (int64, error)
	GetTotalBytes() (int64, error)
}

func IsInLowMemory(getter MemoryGetter, enabledCheck bool, percentageThreshold int) (bool, error) {
	if !enabledCheck {
		return false, nil
	}

	availableMemory, err := getter.GetTotalBytes()
	if err != nil {
		if errors.Is(err, ErrNotSupported) {
			return false, nil
		}
		return false, err
	}
	availableMemoryMB := availableMemory / 1024 / 1024
	lowMemoryThresholdMB := availableMemoryMB * int64(percentageThreshold) / 100
	used, err := getter.GetUsedBytes()
	if err != nil {
		return false, err
	}
	usageMB := used / 1024 / 1024
	return usageMB >= lowMemoryThresholdMB, nil
}

func LowMemoryHealthAdapter(isLowMemory func() (bool, error)) func() error {
	return func() error {
		out, err := isLowMemory()
		if err != nil {
			return err
		}
		if out {
			return ErrLowMemory
		}
		return nil
	}
}
