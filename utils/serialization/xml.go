package serialization

import (
	"bytes"
	"encoding/xml"

	"golang.org/x/net/html/charset"
)

// UnmarshalXml was introduced instead
// of using xml.Unmarshal() as this only supports UTF8
// But its been noticed that UnmarshalXml doesn't support UTF16
func UnmarshallXML(data []byte, value interface{}) error {
	// Read the XML file and create an in-memory model constructed from the
	// elements in the data
	reader := bytes.NewReader(data)
	decoder := xml.NewDecoder(reader)

	decoder.CharsetReader = charset.NewReaderLabel
	return decoder.Decode(&value)
}
