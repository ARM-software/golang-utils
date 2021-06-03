package serialization

import (
	"path"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ARMmbed/golang-utils/utils/filesystem"
)

func TestUnmarshalAttr(t *testing.T) {
	type ParamVal struct {
		Int int `xml:"int,attr"`
	}

	type ParamPtr struct {
		Int *int `xml:"int,attr"`
	}

	type ParamStringPtr struct {
		Int *string `xml:"int,attr"`
	}

	x := []byte(`<Param int="1" />`)

	paramPtr := &ParamPtr{}
	err := UnmarshallXML(x, paramPtr)
	assert.Nil(t, err)
	assert.Equal(t, 1, *paramPtr.Int)

	paramVal := &ParamVal{}
	err = UnmarshallXML(x, paramVal)
	assert.Nil(t, err)
	assert.Equal(t, 1, paramVal.Int)

	paramStrPtr := &ParamStringPtr{}
	err = UnmarshallXML(x, paramStrPtr)
	assert.Nil(t, err)
	assert.Equal(t, "1", *paramStrPtr.Int)
}

func TestUnmarshalUTF8(t *testing.T) {
	type TestStruct struct {
		Attr string `xml:",attr"`
	}

	const inputData = `
	<?xml version="1.0" charset="utf-8"?>
	<Test Attr="Hello, 世界" />`

	expected := "Hello, 世界"

	var x TestStruct
	err := UnmarshallXML([]byte(inputData), &x)
	assert.Nil(t, err)
	assert.Equal(t, expected, x.Attr)
}

func TestUnmarshal_UTF8_GB2312(t *testing.T) {
	type Pack struct {
		URL     string
		DataChn string
		Name    string `xml:",attr"`
		Vendor  string `xml:",attr"`
	}

	var pack Pack
	byteValue, err := filesystem.ReadFile(path.Join("testdata", "testfile_GB2312.xml"))
	require.Nil(t, err)
	require.Nil(t, UnmarshallXML(byteValue, &pack))

	bytes := []byte(pack.DataChn)
	expectedbytes := []byte("世界")

	assert.Equal(t, expectedbytes, bytes)
	assert.Equal(t, "ARM", pack.Vendor)
	assert.Equal(t, "CMSIS", pack.Name)
	assert.Equal(t, "http://www.keil.com/pack/", pack.URL)
}
