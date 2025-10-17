package tus

import (
	"context"
	"fmt"
	"net/url"
	"regexp"
	"strings"

	"github.com/ARM-software/golang-utils/utils/collection"
	"github.com/ARM-software/golang-utils/utils/commonerrors"
	"github.com/ARM-software/golang-utils/utils/encoding/base64"
	"github.com/ARM-software/golang-utils/utils/field"
	"github.com/ARM-software/golang-utils/utils/hashing"
	"github.com/ARM-software/golang-utils/utils/reflection"
)

const KeyTUSMetadata = "filename"

// GenerateTUSChecksumHeader generate the checksum header value
// See https://tus.io/protocols/resumable-upload#upload-checksum
func GenerateTUSChecksumHeader(hashAlgo, hash string) (header string, err error) {
	hashAlgoC, err := hashing.DetermineHashingAlgorithmCanonicalReference(hashAlgo)
	if err != nil {
		err = commonerrors.WrapError(commonerrors.ErrUnsupported, err, "hashing algorithm is not supported")
		return
	}
	base64Hash := base64.EncodeString(hash)
	header = fmt.Sprintf("%v %v", strings.ToLower(hashAlgoC), base64Hash)
	return
}

// ParseTUSHash parses the checksum header value and tries to determine the different elements it contains.
// See https://tus.io/protocols/resumable-upload#upload-checksum
func ParseTUSHash(checksum string) (hashAlgo, hash string, err error) {
	if reflection.IsEmpty(checksum) {
		err = commonerrors.UndefinedVariable("checksum")
		return
	}
	match := regexp.MustCompile(`^\s*([a-zA-Z0-9-_]+)\s+([-A-Za-z0-9+/]*={0,3})$`).FindStringSubmatch(checksum)
	if match == nil || len(match) != 3 {
		err = commonerrors.Newf(commonerrors.ErrInvalid, "invalid checksum format")
		return
	}

	hashAlgo, err = hashing.DetermineHashingAlgorithmCanonicalReference(match[1])
	if err != nil {
		err = commonerrors.WrapError(commonerrors.ErrUnsupported, err, "hashing algorithm is not supported")
		return
	}

	h := strings.TrimSpace(match[2])
	if !base64.IsEncoded(h) {
		err = commonerrors.Newf(commonerrors.ErrInvalid, "checksum is not base64 encoded")
		return
	}
	hash, err = base64.DecodeString(context.Background(), h)
	if err != nil {
		err = commonerrors.WrapError(commonerrors.ErrMarshalling, err, "failed decoding checksum")
	}
	return
}

// GenerateTUSConcatFinalHeader generates the `Concat` header value https://tus.io/protocols/resumable-upload#upload-concat
func GenerateTUSConcatFinalHeader(partials []*url.URL) (header string, err error) {
	header = fmt.Sprintf("final;%v", strings.Join(collection.Map[*url.URL, string](partials, func(u *url.URL) string {
		if u == nil {
			return ""
		}
		return u.EscapedPath()
	}), " "))
	return
}

// ParseTUSConcatHeader parses the `Concat` header value https://tus.io/protocols/resumable-upload#upload-concat
func ParseTUSConcatHeader(concat string) (isPartial bool, partials []*url.URL, err error) {
	header := strings.TrimSpace(concat)
	if reflection.IsEmpty(header) {
		err = commonerrors.UndefinedVariable("concat header")
		return
	}
	if strings.EqualFold(header, "partial") {
		isPartial = true
		return
	}
	h := strings.TrimPrefix(header, "final;")
	if header == h {
		err = commonerrors.New(commonerrors.ErrInvalid, "invalid header value")
		return
	}
	partials, err = collection.MapWithError[string, *url.URL](collection.ParseListWithCleanup(h, " "), url.Parse)
	if err != nil {
		err = commonerrors.WrapError(commonerrors.ErrInvalid, commonerrors.New(commonerrors.ErrMarshalling, "invalid partial url value"), "invalid header value")
	}
	if len(partials) == 0 {
		err = commonerrors.New(commonerrors.ErrInvalid, "no partial url found")
	}
	return
}

// GenerateTUSMetadataHeader generates the `metadata` header value https://tus.io/protocols/resumable-upload#upload-metadata
func GenerateTUSMetadataHeader(filename *string, elements map[string]any) (header string, err error) {
	newMap := make(map[string]string, len(elements))
	for key, value := range elements {
		valueB, ok := value.(bool)
		if ok && valueB {
			newMap[key] = ""
		} else {
			newMap[key] = base64.EncodeString(fmt.Sprintf("%v", value))
		}
	}
	if !reflection.IsEmpty(filename) {
		newMap[KeyTUSMetadata] = base64.EncodeString(field.OptionalString(filename, ""))
	}
	header = strings.Join(collection.ConvertMapToPairSlice(newMap, " "), ",")
	return
}

// ParseTUSMetadataHeader parses the `metadata` header value https://tus.io/protocols/resumable-upload#upload-metadata
func ParseTUSMetadataHeader(header string) (filename *string, elements map[string]any, err error) {
	h := strings.TrimSpace(header)
	if reflection.IsEmpty(h) {
		err = commonerrors.UndefinedVariable("metadata header")
		return
	}
	e := collection.ParseCommaSeparatedList(h)
	if len(e) == 0 {
		err = commonerrors.UndefinedVariable("metadata header")
		return
	}
	elements = make(map[string]any, len(e)/2)
	for i := range e {
		subElem := collection.ParseListWithCleanup(e[i], " ")
		switch len(subElem) {
		case 1:
			elements[subElem[0]] = true
		case 2:
			key := subElem[0]
			value := subElem[1]
			_, has := elements[key]
			if has {
				err = commonerrors.WrapError(commonerrors.ErrInvalid, commonerrors.Newf(commonerrors.ErrInvalid, "duplicated key [%v]", key), "invalid metadata element")
				return
			}
			if !base64.IsEncoded(value) {
				err = commonerrors.WrapError(commonerrors.ErrInvalid, commonerrors.New(commonerrors.ErrMarshalling, "value is not base64 encoded"), "invalid metadata element")
				return
			}
			v, subErr := base64.DecodeString(context.Background(), value)
			if subErr != nil {
				err = commonerrors.WrapError(commonerrors.ErrInvalid, commonerrors.New(commonerrors.ErrMarshalling, "value is not base64 encoded"), "invalid metadata element")
				return
			}
			elements[key] = v
			if strings.EqualFold(key, KeyTUSMetadata) {
				filename = field.ToOptionalString(v)
			}

		default:
			err = commonerrors.New(commonerrors.ErrInvalid, "invalid metadata header element")
			return
		}
	}

	return
}
