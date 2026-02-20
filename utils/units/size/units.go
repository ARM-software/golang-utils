// Package size define common size units
package size

import (
	"fmt"
	"math"
	"regexp"
	"strconv"
	"strings"

	"github.com/ARM-software/golang-utils/utils/commonerrors"
	"github.com/ARM-software/golang-utils/utils/safecast"
	"github.com/ARM-software/golang-utils/utils/units/multiplication"
)

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

var (
	DecimalSIUnits = map[string]float64{
		"B":  B,
		"KB": KB,
		"MB": MB,
		"GB": GB,
		"TB": TB,
		"PB": PB,
		"EB": EB,
		"ZB": ZB,
		"YB": YB,
	}
	BinarySIUnits = map[string]float64{
		"B":   B,
		"KiB": KiB,
		"MiB": MiB,
		"GiB": GiB,
		"TiB": TiB,
		"PiB": PiB,
		"EiB": EiB,
		"ZiB": ZiB,
		"YiB": YiB,
	}
)

var sizeRegex = regexp.MustCompile(`\s*(?P<value>[+-]?\s*[0-9]+[0-9,]*[.]?[0-9]*)\s*(?P<unit>[KkMmGgTtPpEeZzYy]i?[Bb]?)?`)

func ParseSize(s string) (value float64, err error) {
	matches := sizeRegex.FindStringSubmatch(s)
	if len(matches) == 0 {
		return 0, commonerrors.New(commonerrors.ErrInvalid, "string does not represent a size")
	}
	valueIndex := sizeRegex.SubexpIndex("value")
	valuePart, err := strconv.ParseFloat(matches[valueIndex], 64)
	if err != nil {
		return 0, commonerrors.WrapError(commonerrors.ErrMarshalling, err, "failed parsing the size")
	}
	unitIndex := sizeRegex.SubexpIndex("unit")
	unitValue := 1.0
	if unitIndex >= 0 {
		unitValue, _, err = FindUnit(matches[unitIndex])
		if err != nil {
			err = commonerrors.WrapError(commonerrors.ErrMarshalling, err, "failed parsing the size")
			return
		}
	}
	value = valuePart * unitValue
	return
}

func FindUnit(unitStr string) (value float64, canonicalUnit string, err error) {
	canonicalUnit = strings.ToUpper(strings.TrimSpace(unitStr))
	if canonicalUnit == "" {
		canonicalUnit = "B"
		value = 1
		return
	}
	if !strings.HasSuffix(unitStr, "B") {
		canonicalUnit += "B"
	}
	canonicalUnit = strings.ReplaceAll(canonicalUnit, "I", "i")
	value, ok := DecimalSIUnits[canonicalUnit]
	if ok {
		return
	}
	value, ok = BinarySIUnits[canonicalUnit]
	if ok {
		return
	}
	err = commonerrors.Newf(commonerrors.ErrNotFound, "unknown unit [%v]", canonicalUnit)
	canonicalUnit = ""
	return
}

// FormatSizeAsDecimalSI formats a size into a decimal SI string (https://en.wikipedia.org/wiki/Binary_prefix)
// scale corresponds to the number of decimal places. For no limits, set to a negative number
func FormatSizeAsDecimalSI(value float64, scale int) (string, error) {
	return formatSize(value, decimalFormatFunc, scale)
}

// FormatSizeAsBinarySI formats a size into a binary SI string (https://en.wikipedia.org/wiki/Binary_prefix)
// scale corresponds to the number of decimal places. For no limits, set to a negative number
func FormatSizeAsBinarySI(value float64, scale int) (string, error) {
	return formatSize(value, binaryFormatFunc, scale)
}

func formatSize(value float64, formatValueFunc func(value float64) (valueInUnit float64, unit string), scale int) (str string, err error) {
	builder := strings.Builder{}
	if value < 0 {
		_, err = builder.WriteString("-")
		if err != nil {
			err = commonerrors.WrapError(commonerrors.ErrMarshalling, err, "failed formatting the size")
			return
		}
		value = math.Abs(value)
	}

	valueInUnit, unit := formatValueFunc(value)
	formatDecimal := ""
	switch {
	case scale == 0:
		formatDecimal = "%d"
	case scale > 0:
		formatDecimal = "%" + fmt.Sprintf(".%vf", scale)
	default:
		formatDecimal = "%v"
	}
	if scale == 0 {
		_, err = fmt.Fprintf(&builder, formatDecimal, safecast.ToUint64(math.Round(valueInUnit)))
	} else {
		_, err = fmt.Fprintf(&builder, formatDecimal, valueInUnit)
	}
	if err != nil {
		err = commonerrors.WrapError(commonerrors.ErrMarshalling, err, "failed formatting the size")
		return
	}
	_, err = builder.WriteString(unit)
	if err != nil {
		err = commonerrors.WrapError(commonerrors.ErrMarshalling, err, "failed formatting the size")
		return
	}
	str = strings.TrimSpace(builder.String())
	return
}

func decimalFormatFunc(value float64) (valueInUnit float64, unit string) {
	switch {
	case value < KB:
		unit = "B"
		valueInUnit = value
	case value < MB:
		unit = "KB"
		valueInUnit = value / KB
	case value < GB:
		unit = "MB"
		valueInUnit = value / MB
	case value < TB:
		unit = "GB"
		valueInUnit = value / GB
	case value < PB:
		unit = "TB"
		valueInUnit = value / TB
	case value < EB:
		unit = "PB"
		valueInUnit = value / PB
	case value < ZB:
		unit = "EB"
		valueInUnit = value / EB
	case value < YB:
		unit = "ZB"
		valueInUnit = value / ZB
	default:
		unit = "YB"
		valueInUnit = value / YB
	}
	return
}

func binaryFormatFunc(value float64) (valueInUnit float64, unit string) {
	switch {
	case value < KiB:
		unit = "B"
		valueInUnit = value
	case value < MiB:
		unit = "KiB"
		valueInUnit = value / KiB
	case value < GiB:
		unit = "MiB"
		valueInUnit = value / MiB
	case value < TiB:
		unit = "GiB"
		valueInUnit = value / GiB
	case value < PiB:
		unit = "TiB"
		valueInUnit = value / TiB
	case value < EiB:
		unit = "PiB"
		valueInUnit = value / PiB
	case value < ZiB:
		unit = "EiB"
		valueInUnit = value / EiB
	case value < YiB:
		unit = "ZiB"
		valueInUnit = value / ZiB
	default:
		unit = "YiB"
		valueInUnit = value / YiB
	}
	return
}
