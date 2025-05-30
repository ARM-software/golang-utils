package headers

import "net/http"

//nolint:goimport
//go:generate go tool mockgen -destination=../../mocks/mock_$GOPACKAGE.go -package=mocks github.com/ARM-software/golang-utils/utils/http/$GOPACKAGE IHTTPHeaders

// IHTTPHeaders defines an HTTP header.
type IHTTPHeaders interface {
	AppendHeader(key, value string)
	Append(h *Header)
	Has(h *Header) bool
	HasHeader(key string) bool
	Empty() bool
	AppendToResponse(w http.ResponseWriter)
}
