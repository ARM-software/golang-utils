// Package size define common size units
package size

import "github.com/ARM-software/golang-utils/utils/units/multiplication"

const (
	B = 1

	KB = multiplication.Kilo * B
	MB = multiplication.Mega * B
	GB = multiplication.Giga * B
	TB = multiplication.Tera * B
	PB = multiplication.Peta * B
	EB = multiplication.Exa * B
	ZB = multiplication.Zeta * B
	YB = multiplication.Yotta * B

	KiB = float64(1 << 10)
	MiB = float64(1 << 20)
	GiB = float64(1 << 30)
	TiB = float64(1 << 40)
	PiB = float64(1 << 50)
	EiB = float64(1 << 60)
	ZiB = float64(1 << 70)
	YiB = float64(1 << 80)
)
