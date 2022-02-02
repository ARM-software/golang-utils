package charset

import (
	"fmt"
	"golang.org/x/text/encoding"
	"strings"

	"github.com/ARM-software/golang-utils/utils/commonerrors"
)

var (
	unsupportedCharsets = []string{"UTF-7-IMAP"}
)

// GetUnsupported gets valid IANA charset encoding we know are not supported by golang but not reported as such.
func GetUnsupported(name string) (encoding.Encoding, error) {
	for i := range unsupportedCharsets {
		if strings.EqualFold(unsupportedCharsets[i], name) {
			return nil, nil
		}
	}
	return nil, fmt.Errorf("%w encoding name", commonerrors.ErrInvalid)
}
