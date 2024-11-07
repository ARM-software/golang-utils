package safecast

import (
	"math"
	"testing"
)

func FuzzToInt(f *testing.F) {
	f.Add(0)
	f.Add(math.MinInt)
	f.Add(math.MaxInt)
	f.Fuzz(func(t *testing.T, from int) {
		defer func() {
			if r := recover(); r != nil {
				t.Fatalf("panic: %v", r)
			}
		}()
		_ = ToInt(from)
	})
}

func FuzzToInt8(f *testing.F) {
	f.Add(int8(0))
	f.Add(int8(math.MinInt8))
	f.Add(int8(math.MaxInt8))
	f.Fuzz(func(t *testing.T, from int8) {
		defer func() {
			if r := recover(); r != nil {
				t.Fatalf("panic: %v", r)
			}
		}()
		_ = ToInt8(from)
	})
}

func FuzzToInt16(f *testing.F) {
	f.Add(int16(0))
	f.Add(int16(math.MinInt16))
	f.Add(int16(math.MaxInt16))
	f.Fuzz(func(t *testing.T, from int16) {
		defer func() {
			if r := recover(); r != nil {
				t.Fatalf("panic: %v", r)
			}
		}()
		_ = ToInt16(from)
	})
}

func FuzzToInt32(f *testing.F) {
	f.Add(int32(0))
	f.Add(int32(math.MinInt32))
	f.Add(int32(math.MaxInt32))
	f.Fuzz(func(t *testing.T, from int32) {
		defer func() {
			if r := recover(); r != nil {
				t.Fatalf("panic: %v", r)
			}
		}()
		_ = ToInt32(from)
	})
}

func FuzzToInt64(f *testing.F) {
	f.Add(int64(0))
	f.Add(int64(math.MinInt64))
	f.Add(int64(math.MaxInt64))
	f.Fuzz(func(t *testing.T, from int64) {
		defer func() {
			if r := recover(); r != nil {
				t.Fatalf("panic: %v", r)
			}
		}()
		_ = ToInt64(from)
	})
}

func FuzzToUint(f *testing.F) {
	f.Add(uint(0))
	f.Add(uint(math.MaxUint))
	f.Fuzz(func(t *testing.T, from uint) {
		defer func() {
			if r := recover(); r != nil {
				t.Fatalf("panic: %v", r)
			}
		}()
		_ = ToUint(from)
	})
}

func FuzzToUint8(f *testing.F) {
	f.Add(uint8(0))
	f.Add(uint8(math.MaxUint8))
	f.Fuzz(func(t *testing.T, from uint8) {
		defer func() {
			if r := recover(); r != nil {
				t.Fatalf("panic: %v", r)
			}
		}()
		_ = ToUint8(from)
	})
}

func FuzzToUint16(f *testing.F) {
	f.Add(uint16(0))
	f.Add(uint16(math.MaxUint16))
	f.Fuzz(func(t *testing.T, from uint16) {
		defer func() {
			if r := recover(); r != nil {
				t.Fatalf("panic: %v", r)
			}
		}()
		_ = ToUint16(from)
	})
}

func FuzzToUint32(f *testing.F) {
	f.Add(uint32(0))
	f.Add(uint32(math.MaxUint32))
	f.Fuzz(func(t *testing.T, from uint32) {
		defer func() {
			if r := recover(); r != nil {
				t.Fatalf("panic: %v", r)
			}
		}()
		_ = ToUint32(from)
	})
}

func FuzzToUint64(f *testing.F) {
	f.Add(uint64(0))
	f.Add(uint64(math.MaxUint64))
	f.Fuzz(func(t *testing.T, from uint64) {
		defer func() {
			if r := recover(); r != nil {
				t.Fatalf("panic: %v", r)
			}
		}()
		_ = ToUint64(from)
	})
}
