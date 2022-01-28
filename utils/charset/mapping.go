package charset

import (
	"fmt"
	"strings"
	"sync"

	"github.com/ARM-software/golang-utils/utils/commonerrors"
)

var (
	mapping     charsetEncodingMapping
	mappingOnce sync.Once
)

type ICharsetEncodingMapping interface {
	GetCanonicalName(alias string) (string, error)
}

type charsetEncodingMapping struct {
	mapping map[string]string
}

func (m *charsetEncodingMapping) GetCanonicalName(alias string) (name string, err error) {
	name, found := m.mapping[strings.ToLower(strings.TrimSpace(alias))]
	if !found {
		err = fmt.Errorf("%w: charset alias [%v] was not found in the list of supported Charsets", commonerrors.ErrNotFound, alias)
	}
	return
}

func initialiseMapping() {
	mappingOnce.Do(func() {
		mapping = charsetEncodingMapping{mapping: map[string]string{
			"iso-ir-6":         "US-ASCII",
			"ansi_x3.4-1968":   "US-ASCII",
			"ansi_x3.4-1986":   "US-ASCII",
			"iso_646.irv:1991": "US-ASCII",
			"iso646-US":        "US-ASCII",
			"us-ascii":         "US-ASCII",
			"us":               "US-ASCII",
			"ibm367":           "US-ASCII",
			"cp367":            "US-ASCII",
			"csascii":          "US-ASCII",

			"iso-ir-100":  "ISO_8859-1:1987",
			"iso_8859-1":  "ISO_8859-1:1987",
			"iso-8859-1":  "ISO_8859-1:1987",
			"latin1":      "ISO_8859-1:1987",
			"l1":          "ISO_8859-1:1987",
			"ibm819":      "ISO_8859-1:1987",
			"cp819":       "ISO_8859-1:1987",
			"csisolatin1": "ISO_8859-1:1987",

			"iso-ir-101":  "ISO_8859-2:1987",
			"iso_8859-2":  "ISO_8859-2:1987",
			"iso-8859-2":  "ISO_8859-2:1987",
			"latin2":      "ISO_8859-2:1987",
			"l2":          "ISO_8859-2:1987",
			"csisolatin2": "ISO_8859-2:1987",

			"iso-ir-109":  "ISO_8859-3:1988",
			"iso_8859-3":  "ISO_8859-3:1988",
			"iso-8859-3":  "ISO_8859-3:1988",
			"latin3":      "ISO_8859-3:1988",
			"l3":          "ISO_8859-3:1988",
			"csisolatin3": "ISO_8859-3:1988",

			"iso-ir-110":  "ISO_8859-4:1988",
			"iso_8859-4":  "ISO_8859-4:1988",
			"iso-8859-4":  "ISO_8859-4:1988",
			"latin4":      "ISO_8859-4:1988",
			"l4":          "ISO_8859-4:1988",
			"csisolatin4": "ISO_8859-4:1988",

			"iso-ir-144":         "ISO_8859-5:1988",
			"iso_8859-5":         "ISO_8859-5:1988",
			"iso-8859-5":         "ISO_8859-5:1988",
			"cyrillic":           "ISO_8859-5:1988",
			"csisolatincyrillic": "ISO_8859-5:1988",

			"iso-ir-127":       "ISO_8859-6:1987",
			"iso_8859-6":       "ISO_8859-6:1987",
			"iso-8859-6":       "ISO_8859-6:1987",
			"ecma-114":         "ISO_8859-6:1987",
			"asmo-708":         "ISO_8859-6:1987",
			"arabic":           "ISO_8859-6:1987",
			"csisolatinarabic": "ISO_8859-6:1987",

			"iso-ir-126":      "ISO_8859-7:1987",
			"iso_8859-7":      "ISO_8859-7:1987",
			"iso-8859-7":      "ISO_8859-7:1987",
			"elot_928":        "ISO_8859-7:1987",
			"ecma-118":        "ISO_8859-7:1987",
			"greek":           "ISO_8859-7:1987",
			"greek8":          "ISO_8859-7:1987",
			"csisolatingreek": "ISO_8859-7:1987",
			"latingreek":      "ISO_8859-7:1987",
			"latin-greek":     "ISO_8859-7:1987",

			"iso-ir-138":       "ISO_8859-8:1988",
			"iso_8859-8":       "ISO_8859-8:1988",
			"iso-8859-8":       "ISO_8859-8:1988",
			"hebrew":           "ISO_8859-8:1988",
			"csisolatinhebrew": "ISO_8859-8:1988",

			"iso-ir-148":  "ISO_8859-9:1989",
			"iso_8859-9":  "ISO_8859-9:1989",
			"iso-8859-9":  "ISO_8859-9:1989",
			"latin5":      "ISO_8859-9:1989",
			"l5":          "ISO_8859-9:1989",
			"csisolatin5": "ISO_8859-9:1989",

			"iso-ir-157":       "ISO-8859-10",
			"l6":               "ISO-8859-10",
			"iso_8859-10:1992": "ISO-8859-10",
			"csisolatin6":      "ISO-8859-10",
			"latin6":           "ISO-8859-10",

			"ms_kanji":   "Shift_JIS",
			"csshiftjis": "Shift_JIS",

			"iso-ir-149":     "KS_C_5601-1987",
			"ks_c_5601-1989": "KS_C_5601-1987",
			"ksc_5601":       "KS_C_5601-1987",
			"korean":         "KS_C_5601-1987",
			"csksc56011987":  "KS_C_5601-1987",

			"csiso2022kr": "ISO-2022-KR",

			"cseuckr": " EUC-KR",

			"csiso2022jp": "ISO-2022-JP",

			"iso-ir-58":       "GB_2312-80",
			"chinese":         "GB_2312-80",
			"csiso58gb231280": "GB_2312-80",

			"csiso2022cn": "ISO-2022-CN",

			"csiso2022cnext": "ISO-2022-CN-EXT",

			"csutf8": "UTF-8",

			"csiso885913": "ISO-8859-13",

			"iso-ir-199":       "ISO-8859-14",
			"iso_8859-14:1998": "ISO-8859-14",
			"iso_8859-14":      "ISO-8859-14",
			"latin8":           "ISO-8859-14",
			"iso-celtic":       "ISO-8859-14",
			"l8":               "ISO-8859-14",
			"csiso885914":      "ISO-8859-14",

			"iso_8859-15": "ISO-8859-15",
			"latin-9":     "ISO-8859-15",
			"csiso885915": "ISO-8859-15",

			"iso-ir-226":       "ISO-8859-16",
			"iso_8859-16:2001": "ISO-8859-16",
			"iso_8859-16":      "ISO-8859-16",
			"latin10":          "ISO-8859-16",
			"l10":              "ISO-8859-16",
			"csiso885916":      "ISO-8859-16",

			"cp936":       "GBK",
			"ms936":       "GBK",
			"windows-936": "GBK",
			"csgbk":       "GBK",

			"csgb18030": "GB18030",

			"csutf16be": "UTF-16BE",

			"csutf16le": "UTF-16LE",

			"csutf16": "UTF-16",

			"cswindows31j": "Windows-31J",

			"csgb2312": "GB2312",

			"csbig5":   "Big5",
			"big-5":    "Big5",
			"big-five": "Big5",
			"bigfive":  "Big5",

			"mac":         "macintosh",
			"csmacintosh": "macintosh",

			"cskoi8r": "KOI8-R",

			"cp866":    "IBM866",
			"866":      "IBM866",
			"csibm866": "IBM866",

			"cskoi8u": "KOI8-U",

			"csbig5hkscs": "Big5-HKSCS",
			"big5hkscs":   "Big5-HKSCS",

			"cswindows874": "windows-874",

			"cswindows1250": "windows-1250",

			"cswindows1251": "windows-1251",

			"cswindows1252": "windows-1252",

			"cswindows1253": "windows-1253",

			"cswindows1254": "windows-1254",

			"cswindows1255": "windows-1255",

			"cswindows1256": "windows-1256",

			"cswindows1257": "windows-1257",

			"cswindows1258": "windows-1258",

			"cstis620":    "TIS-620",
			"iso-8859-11": "TIS-620",
		},
		}
	})
}

func getEncodingMapping() ICharsetEncodingMapping {
	initialiseMapping()
	fmt.Println(mapping)
	return &mapping
}
