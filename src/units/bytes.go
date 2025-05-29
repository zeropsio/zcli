package units

import (
	"fmt"

	"golang.org/x/exp/constraints"
)

const (
	Bi  = uint64(1)
	KiB = 1024 * Bi
	MiB = 1024 * KiB
	GiB = 1024 * MiB
	TiB = 1024 * GiB
)

const (
	B  = uint64(1)
	KB = 1000 * B
	MB = 1000 * KB
	GB = 1000 * MB
	TB = 1000 * GB
)

func ByteCount[T constraints.Unsigned](b T, unit uint64, suffix string) string {
	if uint64(b) < unit {
		return fmt.Sprintf("%d %s", b, suffix)
	}
	div, exp := unit, 0
	for n := uint64(b) / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	suffixes := "KMGTPE"
	if unit == 1000 {
		suffixes = "kMGTPE"
	}
	return fmt.Sprintf("%.1f %c%s", float64(b)/float64(div), suffixes[exp], suffix)
}

func ByteCountSI[T constraints.Unsigned](b T) string {
	return ByteCount[T](b, 1000, "B")
}

func ByteCountIEC[T constraints.Unsigned](b T) string {
	return ByteCount[T](b, 1024, "iB")
}
