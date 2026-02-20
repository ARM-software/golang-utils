package maps

import (
	"testing"
	"time"

	"github.com/go-faker/faker/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ARM-software/golang-utils/utils/commonerrors"
	"github.com/ARM-software/golang-utils/utils/commonerrors/errortest"
	mapstest "github.com/ARM-software/golang-utils/utils/serialization/maps/testing" //nolint:misspell
)

type TestStruct0 struct {
	Number             int
	BigNumber          int64
	Float64            float64
	Uint               uint
	LongString         string
	OtherString        string
	Domain             string `faker:"domain_name"`
	Array              []string
	Bool               bool
	Latitude           float32           `faker:"lat"`
	Longitude          float32           `faker:"long"`
	RealAddress        faker.RealAddress `faker:"real_address"`
	CreditCardNumber   string            `faker:"cc_number"`
	CreditCardType     string            `faker:"cc_type"`
	Email              string            `faker:"email"`
	DomainName         string            `faker:"domain_name"`
	IPV4               string            `faker:"ipv4"`
	IPV6               string            `faker:"ipv6"`
	Password           string            `faker:"password"` //nolint:gosec //G117: Exported struct field
	Jwt                string            `faker:"jwt"`      //nolint:gosec //G117: Exported struct field
	PhoneNumber        string            `faker:"phone_number"`
	MacAddress         string            `faker:"mac_address"`
	URL                string            `faker:"url"`
	UserName           string            `faker:"username"`
	TollFreeNumber     string            `faker:"toll_free_number"`
	E164PhoneNumber    string            `faker:"e_164_phone_number"`
	TitleMale          string            `faker:"title_male"`
	TitleFemale        string            `faker:"title_female"`
	FirstName          string            `faker:"first_name"`
	FirstNameMale      string            `faker:"first_name_male"`
	FirstNameFemale    string            `faker:"first_name_female"`
	LastName           string            `faker:"last_name"`
	Name               string            `faker:"name"`
	UnixTime           int64             `faker:"unix_time"`
	Date               string            `faker:"date"`
	Time               string            `faker:"time"`
	MonthName          string            `faker:"month_name"`
	Year               string            `faker:"year"`
	DayOfWeek          string            `faker:"day_of_week"`
	DayOfMonth         string            `faker:"day_of_month"`
	Timestamp          string            `faker:"timestamp"`
	Century            string            `faker:"century"`
	TimeZone           string            `faker:"timezone"`
	TimePeriod         string            `faker:"time_period"`
	Word               string            `faker:"word"`
	Sentence           string            `faker:"sentence"`
	Paragraph          string            `faker:"paragraph"`
	Currency           string            `faker:"currency"`
	Amount             float64           `faker:"amount"`
	AmountWithCurrency string            `faker:"amount_with_currency"`
	UUIDHypenated      string            `faker:"uuid_hyphenated"`
	UUID               string            `faker:"uuid_digit"`
	PaymentMethod      string            `faker:"oneof: cc, paypal, check, money order"` // oneof will randomly pick one of the comma-separated values supplied in the tag
	AccountID          int               `faker:"oneof: 15, 27, 61"`                     // use commas to separate the values for now. Future support for other separator characters may be added
	Price32            float32           `faker:"oneof: 4.95, 9.99, 31997.97"`
	Price64            float64           `faker:"oneof: 47463.9463525, 993747.95662529, 11131997.978767990"`
	NumS64             int64             `faker:"oneof: 1, 2"`
	NumS32             int32             `faker:"oneof: -3, 4"`
	NumS16             int16             `faker:"oneof: -5, 6"`
	NumS8              int8              `faker:"oneof: 7, -8"`
	NumU64             uint64            `faker:"oneof: 9, 10"`
	NumU32             uint32            `faker:"oneof: 11, 12"`
	NumU16             uint16            `faker:"oneof: 13, 14"`
	NumU8              uint8             `faker:"oneof: 15, 16"`
	NumU               uint              `faker:"oneof: 17, 18"`
}

type TestStruct1 struct {
	Name        string
	Number      int
	BigNumber   int64
	Float64     float64
	Duration    int `mapstructure:"time_duration"`
	Uint        uint
	LongString  string
	OtherString string
	Domain      string `faker:"domain_name"`
	UUID        string
	Array       []int
	Bool        bool
	Struct      TestStruct0
}

type TestStruct2WithTime struct {
	Time     time.Time     `mapstructure:"some_time"`
	Duration time.Duration `mapstructure:"some_duration"`
}

type TestStruct3WithTime struct {
	Time     time.Time     `mapstructure:"some_time"`
	Duration time.Duration `mapstructure:"some_duration"`
	Struct   TestStruct2WithTime
}

type TestStructWithEnum struct {
	Time     time.Time                      `mapstructure:"some_time"`
	TestEnum mapstest.TestEnumWithUnmarshal `mapstructure:"test_enum"`
	Struct   TestStruct2WithTime
}

type TestStructWithEnumInvalid struct {
	Time     time.Time                         `mapstructure:"some_time"`
	TestEnum mapstest.TestEnumWithoutUnmarshal `mapstructure:"test_enum"`
	Struct   TestStruct2WithTime
}

type TestStruct4 struct {
	Field1 string `mapstructure:"field_1"`
	Field2 string `mapstructure:"field_2"`
}

type TestStruct5WithEmbeddedStruct struct {
	TestStruct4
	Field3 string `mapstructure:"field_3"`
}

type TestStruct5WithEmbeddedStructAndSquashTag struct {
	TestStruct4 `mapstructure:",squash"`
	Field3      string `mapstructure:"field_3"`
}

type TestStruct6WithEmbeddedStruct struct {
	TestStruct5WithEmbeddedStruct
	Field4 string `mapstructure:"field_4"`
}

type TestStruct6WithEmbeddedStructAndSquashTag struct {
	TestStruct5WithEmbeddedStructAndSquashTag `mapstructure:",squash"`
	Field4                                    string `mapstructure:"field_4"`
}

func TestToMap(t *testing.T) {
	t.Run("generic", func(t *testing.T) {
		testStruct := TestStruct1{}
		require.NoError(t, faker.FakeData(&testStruct))
		if len(testStruct.Array) == 0 {
			// This is to avoid the case where the slice is empty and so the comparison may differ because the slice could be set to nil instead of an empty slice
			testStruct.Array = []int{1212, 544}
		}

		structMap, err := ToMap[TestStruct1](&testStruct)
		require.NoError(t, err)
		newStruct := TestStruct1{}
		require.NoError(t, FromMap[TestStruct1](structMap, &newStruct))
		assert.Equal(t, testStruct, newStruct)
	})
	t.Run("with time", func(t *testing.T) {
		random, err := faker.RandomInt(0, 1000, 2)
		require.NoError(t, err)
		testStruct := TestStruct3WithTime{
			Time:     time.Now().UTC(),
			Duration: time.Duration(random[0]) * time.Minute,
			Struct: TestStruct2WithTime{
				Time:     time.Unix(faker.RandomUnixTime(), 0),
				Duration: time.Duration(random[1]) * time.Second,
			},
		}
		structMap, err := ToMap[TestStruct3WithTime](&testStruct)
		require.NoError(t, err)
		_, err = ToMapFromPointer[TestStruct3WithTime](testStruct)
		errortest.AssertError(t, err, commonerrors.ErrInvalid)
		newStruct := TestStruct3WithTime{}
		require.NoError(t, FromMap[TestStruct3WithTime](structMap, &newStruct))
		errortest.AssertError(t, FromMapToPointer[TestStruct3WithTime](structMap, newStruct), commonerrors.ErrInvalid)
		assert.WithinDuration(t, testStruct.Time, newStruct.Time, 0)
		assert.Equal(t, testStruct.Duration, newStruct.Duration)
		assert.WithinDuration(t, testStruct.Struct.Time, newStruct.Struct.Time, 0)
		assert.Equal(t, testStruct.Struct.Duration, newStruct.Struct.Duration)
	})
	t.Run("with custom type (has UnmarshalText)", func(t *testing.T) {
		random, err := faker.RandomInt(0, 1000, 2)
		require.NoError(t, err)
		testStruct := TestStructWithEnum{
			Time:     time.Now().UTC(),
			TestEnum: mapstest.TestEnumWithUnmarshal1,
			Struct: TestStruct2WithTime{
				Time:     time.Unix(faker.RandomUnixTime(), 0),
				Duration: time.Duration(random[1]) * time.Second,
			},
		}
		structMap, err := ToMap[TestStructWithEnum](&testStruct)
		structMap["test_enum"] = mapstest.TestEnumStringVer1 // change to the alternate version for unmarshalling
		require.NoError(t, err)
		_, err = ToMapFromPointer[TestStructWithEnum](testStruct)
		errortest.AssertError(t, err, commonerrors.ErrInvalid)
		newStruct := TestStructWithEnum{}
		require.NoError(t, FromMap[TestStructWithEnum](structMap, &newStruct))
		errortest.AssertError(t, FromMapToPointer[TestStructWithEnum](structMap, newStruct), commonerrors.ErrInvalid)
		assert.WithinDuration(t, testStruct.Time, newStruct.Time, 0)
		assert.Equal(t, testStruct.TestEnum, newStruct.TestEnum)
		assert.WithinDuration(t, testStruct.Struct.Time, newStruct.Struct.Time, 0)
		assert.Equal(t, testStruct.Struct.Duration, newStruct.Struct.Duration)
	})
	t.Run("with custom type (no UnmarshalText)", func(t *testing.T) {
		random, err := faker.RandomInt(0, 1000, 2)
		require.NoError(t, err)
		testStruct := TestStructWithEnumInvalid{
			Time:     time.Now().UTC(),
			TestEnum: mapstest.TestEnumWithoutUnmarshal1,
			Struct: TestStruct2WithTime{
				Time:     time.Unix(faker.RandomUnixTime(), 0),
				Duration: time.Duration(random[1]) * time.Second,
			},
		}
		structMap, err := ToMap[TestStructWithEnumInvalid](&testStruct)
		require.NoError(t, err)
		_, err = ToMapFromPointer[TestStructWithEnumInvalid](testStruct)
		errortest.AssertError(t, err, commonerrors.ErrInvalid)
		newStruct := TestStructWithEnumInvalid{}
		require.NoError(t, FromMap[TestStructWithEnumInvalid](structMap, &newStruct))
		errortest.AssertError(t, FromMapToPointer[TestStructWithEnumInvalid](structMap, newStruct), commonerrors.ErrInvalid)
		assert.WithinDuration(t, testStruct.Time, newStruct.Time, 0)
		assert.Equal(t, testStruct.TestEnum, newStruct.TestEnum)
		assert.WithinDuration(t, testStruct.Struct.Time, newStruct.Struct.Time, 0)
		assert.Equal(t, testStruct.Struct.Duration, newStruct.Struct.Duration)
	})
	t.Run("invalid", func(t *testing.T) {
		var testMap map[string]string
		testStruct := TestStruct3WithTime{}
		_, err := ToMapFromPointer[TestStruct3WithTime](testStruct)
		errortest.AssertError(t, err, commonerrors.ErrInvalid, commonerrors.ErrUndefined)
		_, err = ToMapFromPointer[any](testStruct)
		errortest.AssertError(t, err, commonerrors.ErrInvalid, commonerrors.ErrUndefined)
		_, err = ToMapFromPointer[any](nil)
		errortest.AssertError(t, err, commonerrors.ErrInvalid, commonerrors.ErrUndefined)
		_, err = ToMapFromPointer[*TestStruct3WithTime](nil)
		errortest.AssertError(t, err, commonerrors.ErrInvalid, commonerrors.ErrUndefined)
		errortest.AssertError(t, FromMapToPointer[TestStruct3WithTime](testMap, testStruct), commonerrors.ErrInvalid, commonerrors.ErrUndefined)
		errortest.AssertError(t, FromMapToPointer[any](testMap, testStruct), commonerrors.ErrInvalid, commonerrors.ErrUndefined)
		errortest.AssertError(t, FromMapToPointer[any](testMap, nil), commonerrors.ErrInvalid, commonerrors.ErrUndefined)
		errortest.AssertError(t, FromMapToPointer[*TestStruct3WithTime](testMap, nil), commonerrors.ErrInvalid, commonerrors.ErrUndefined)
	})
	t.Run("embedded struct mapping with `mapstructure:\",squash\"`", func(t *testing.T) {
		testStruct := TestStruct6WithEmbeddedStructAndSquashTag{}
		require.NoError(t, faker.FakeData(&testStruct))

		m := make(map[string]string)
		m["field_1"] = faker.Word()
		m["field_2"] = faker.Word()
		m["field_3"] = faker.Word()
		m["field_4"] = faker.Word()

		// Should correctly set embedded struct fields from map m, even on nested embedded structs
		err := FromMap[TestStruct6WithEmbeddedStructAndSquashTag](m, &testStruct)
		require.NoError(t, err)
		assert.Equal(t, testStruct.Field1, m["field_1"])
		assert.Equal(t, testStruct.Field2, m["field_2"])
		assert.Equal(t, testStruct.Field3, m["field_3"])
		assert.Equal(t, testStruct.Field4, m["field_4"])

		structMap, err := ToMap[TestStruct6WithEmbeddedStructAndSquashTag](&testStruct)
		require.NoError(t, err)
		assert.Equal(t, m, structMap)

		newStruct := TestStruct6WithEmbeddedStructAndSquashTag{}
		require.NoError(t, FromMap[TestStruct6WithEmbeddedStructAndSquashTag](structMap, &newStruct))
		assert.Equal(t, testStruct, newStruct)
	})
	t.Run("embedded struct mapping without `mapstructure:\",squash\"`", func(t *testing.T) {
		testStruct := TestStruct6WithEmbeddedStruct{}
		require.NoError(t, faker.FakeData(&testStruct))

		m := make(map[string]string)
		m["field_1"] = faker.Word()
		m["field_2"] = faker.Word()
		m["field_3"] = faker.Word()
		m["field_4"] = faker.Word()

		// Should only set the parent struct fields from map m and ignore any embedded struct fields
		err := FromMap[TestStruct6WithEmbeddedStruct](m, &testStruct)
		require.NoError(t, err)
		assert.NotEqual(t, testStruct.Field1, m["field_1"])
		assert.NotEqual(t, testStruct.Field2, m["field_2"])
		assert.NotEqual(t, testStruct.Field3, m["field_3"])
		assert.Equal(t, testStruct.Field4, m["field_4"])

		structMap, err := ToMap[TestStruct6WithEmbeddedStruct](&testStruct)
		require.NoError(t, err)
		assert.NotEqual(t, m, structMap)

		newStruct := TestStruct6WithEmbeddedStruct{}
		require.NoError(t, FromMap[TestStruct6WithEmbeddedStruct](structMap, &newStruct))
		assert.Equal(t, testStruct, newStruct)
	})
}
