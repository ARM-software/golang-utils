package useragent

import (
	"fmt"
	"strings"

	"github.com/ARM-software/golang-utils/utils/commonerrors"
	"github.com/ARM-software/golang-utils/utils/reflection"
)

// AddValuesToUserAgent extends a user agent string with new elements. See https://en.wikipedia.org/wiki/User-Agent_header#Format_for_human-operated_web_browsers
func AddValuesToUserAgent(userAgent string, elements ...string) (newUserAgent string) {
	if len(elements) == 0 {
		newUserAgent = userAgent
		return
	}
	newUserAgent = strings.Join(elements, " ")
	newUserAgent = strings.TrimSpace(newUserAgent)
	if newUserAgent == "" {
		newUserAgent = userAgent
		return
	}
	if !reflection.IsEmpty(userAgent) {
		newUserAgent = fmt.Sprintf("%v %v", userAgent, newUserAgent)
	}
	return
}

// GenerateUserAgentValue generates a user agent value. See https://en.wikipedia.org/wiki/User-Agent_header#Format_for_human-operated_web_browsers
func GenerateUserAgentValue(product string, productVersion string, comment string) (userAgent string, err error) {
	if reflection.IsEmpty(product) {
		err = commonerrors.UndefinedVariable("product")
		return
	}
	if reflection.IsEmpty(productVersion) {
		err = commonerrors.UndefinedVariable("product version")
		return
	}
	userAgent = fmt.Sprintf("%v/%v", product, productVersion)
	if !reflection.IsEmpty(comment) {
		userAgent = fmt.Sprintf("%v (%v)", userAgent, comment)
	}
	return
}
