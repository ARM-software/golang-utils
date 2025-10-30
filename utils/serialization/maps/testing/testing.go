package testing

import (
	"github.com/ARM-software/golang-utils/utils/commonerrors"
)

type TestEnumWithUnmarshal int

const (
	TestEnumStringVer0 = "test0"
	TestEnumStringVer1 = "test1"
)

func (i *TestEnumWithUnmarshal) UnmarshalText(text []byte) error {
	v, ok := map[string]TestEnumWithUnmarshal{
		TestEnumStringVer0: TestEnumWithUnmarshal0,
		TestEnumStringVer1: TestEnumWithUnmarshal1,
	}[string(text)]
	if !ok {
		return commonerrors.ErrInvalid
	}
	*i = v
	return nil
}

func ValidationFunc(value any) error {
	e, ok := value.(TestEnumWithUnmarshal)
	if !ok || (e != TestEnumWithUnmarshal0 && e != TestEnumWithUnmarshal1) {
		return commonerrors.ErrInvalid
	}
	return nil
}

const (
	TestEnumWithUnmarshal0 TestEnumWithUnmarshal = iota
	TestEnumWithUnmarshal1
)

type TestEnumWithoutUnmarshal int

const (
	TestEnumWithoutUnmarshal0 TestEnumWithoutUnmarshal = iota
	TestEnumWithoutUnmarshal1
)
