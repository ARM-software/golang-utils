/*
 * Copyright (C) 2020-2022 Arm Limited or its affiliates and Contributors. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */
package charset

import (
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
		err = commonerrors.Newf(commonerrors.ErrNotFound, "charset alias [%v] was not found in the list of supported Charsets", alias)
	}
	return
}

func initialiseMapping() {
	mappingOnce.Do(func() {
		// This mapping list was created based on the following indexes:
		// - https://www.iana.org/assignments/character-sets/character-sets.xhtml
		// - https://encoding.spec.whatwg.org/encodings.json
		// - https://www.forcs.com/en/documentation-query/reference_charset_enabled_alias.htm
		mapping = charsetEncodingMapping{mapping: map[string]string{
			"iso-ir-6":         "US-ASCII",
			"ansi_x3.4-1968":   "US-ASCII",
			"ansi_x3.4-1986":   "US-ASCII",
			"iso_646.irv:1991": "US-ASCII",
			"iso646-US":        "US-ASCII",
			"us-ascii":         "US-ASCII",
			"us":               "US-ASCII",
			"ibm367":           "US-ASCII",
			"csibm367":         "US-ASCII",
			"ibm-367":          "US-ASCII",
			"cp367":            "US-ASCII",
			"csascii":          "US-ASCII",

			"iso-ir-100":  "ISO_8859-1:1987",
			"iso_8859-1":  "ISO_8859-1:1987",
			"iso-8859-1":  "ISO_8859-1:1987",
			"latin1":      "ISO_8859-1:1987",
			"l1":          "ISO_8859-1:1987",
			"ibm819":      "ISO_8859-1:1987",
			"csibm819":    "ISO_8859-1:1987",
			"cp819":       "ISO_8859-1:1987",
			"csisolatin1": "ISO_8859-1:1987",
			"8859_1":      "ISO_8859-1:1987",
			"iso8859-1":   "ISO_8859-1:1987",
			"ibm-819":     "ISO_8859-1:1987",
			"819":         "ISO_8859-1:1987",

			"iso-ir-101":  "ISO_8859-2:1987",
			"iso_8859-2":  "ISO_8859-2:1987",
			"iso-8859-2":  "ISO_8859-2:1987",
			"latin2":      "ISO_8859-2:1987",
			"l2":          "ISO_8859-2:1987",
			"csisolatin2": "ISO_8859-2:1987",
			"8859_2":      "ISO_8859-2:1987",
			"iso8859-2":   "ISO_8859-2:1987",
			"ibm912":      "ISO_8859-2:1987",
			"csibm912":    "ISO_8859-2:1987",
			"ibm-912":     "ISO_8859-2:1987",
			"cp912":       "ISO_8859-2:1987",
			"912":         "ISO_8859-2:1987",

			"iso-ir-109":  "ISO_8859-3:1988",
			"iso_8859-3":  "ISO_8859-3:1988",
			"iso-8859-3":  "ISO_8859-3:1988",
			"latin3":      "ISO_8859-3:1988",
			"l3":          "ISO_8859-3:1988",
			"csisolatin3": "ISO_8859-3:1988",
			"8859_3":      "ISO_8859-3:1988",
			"iso8859-3":   "ISO_8859-3:1988",
			"ibm913":      "ISO_8859-3:1988",
			"csibm913":    "ISO_8859-3:1988",
			"ibm-913":     "ISO_8859-3:1988",
			"cp913":       "ISO_8859-3:1988",
			"913":         "ISO_8859-3:1988",

			"iso-ir-110":  "ISO_8859-4:1988",
			"iso_8859-4":  "ISO_8859-4:1988",
			"iso-8859-4":  "ISO_8859-4:1988",
			"latin4":      "ISO_8859-4:1988",
			"l4":          "ISO_8859-4:1988",
			"csisolatin4": "ISO_8859-4:1988",
			"8859_4":      "ISO_8859-4:1988",
			"iso8859-4":   "ISO_8859-4:1988",
			"ibm914":      "ISO_8859-4:1988",
			"csibm914":    "ISO_8859-4:1988",
			"ibm-914":     "ISO_8859-4:1988",
			"cp914":       "ISO_8859-4:1988",
			"914":         "ISO_8859-4:1988",

			"iso-ir-144":         "ISO_8859-5:1988",
			"iso_8859-5":         "ISO_8859-5:1988",
			"iso-8859-5":         "ISO_8859-5:1988",
			"cyrillic":           "ISO_8859-5:1988",
			"csisolatincyrillic": "ISO_8859-5:1988",
			"8859_5":             "ISO_8859-5:1988",
			"iso8859-5":          "ISO_8859-5:1988",
			"ibm915":             "ISO_8859-5:1988",
			"csibm915":           "ISO_8859-5:1988",
			"ibm-915":            "ISO_8859-5:1988",
			"cp915":              "ISO_8859-5:1988",
			"915":                "ISO_8859-5:1988",

			"iso-ir-127":       "ISO_8859-6:1987",
			"iso_8859-6":       "ISO_8859-6:1987",
			"iso-8859-6":       "ISO_8859-6:1987",
			"ecma-114":         "ISO_8859-6:1987",
			"asmo-708":         "ISO_8859-6:1987",
			"arabic":           "ISO_8859-6:1987",
			"csisolatinarabic": "ISO_8859-6:1987",
			"8859_6":           "ISO_8859-6:1987",
			"iso8859-6":        "ISO_8859-6:1987",
			"ibm1089":          "ISO_8859-6:1987",
			"csibm1089":        "ISO_8859-6:1987",
			"ibm-1089":         "ISO_8859-6:1987",
			"cp1089":           "ISO_8859-6:1987",
			"1089":             "ISO_8859-6:1987",

			"iso-ir-126":      "ISO_8859-7:1987",
			"iso_8859-7":      "ISO_8859-7:1987",
			"iso-8859-7":      "ISO_8859-7:1987",
			"elot_928":        "ISO_8859-7:1987",
			"ecma-118":        "ISO_8859-7:1987",
			"greek":           "ISO_8859-7:1987",
			"greek8":          "ISO_8859-7:1987",
			"csisolatingreek": "ISO_8859-7:1987",
			"8859_7":          "ISO_8859-7:1987",
			"iso8859-7":       "ISO_8859-7:1987",
			"ibm813":          "ISO_8859-7:1987",
			"csibm813":        "ISO_8859-7:1987",
			"ibm-813":         "ISO_8859-7:1987",
			"cp813":           "ISO_8859-7:1987",
			"813":             "ISO_8859-7:1987",

			"iso-ir-138":       "ISO_8859-8:1988",
			"iso_8859-8":       "ISO_8859-8:1988",
			"iso-8859-8":       "ISO_8859-8:1988",
			"hebrew":           "ISO_8859-8:1988",
			"csisolatinhebrew": "ISO_8859-8:1988",
			"8859_8":           "ISO_8859-8:1988",
			"iso8859-8":        "ISO_8859-8:1988",
			"ibm916":           "ISO_8859-8:1988",
			"csibm916":         "ISO_8859-8:1988",
			"ibm-916":          "ISO_8859-8:1988",
			"cp916":            "ISO_8859-8:1988",
			"916":              "ISO_8859-8:1988",

			"iso-ir-148":  "ISO_8859-9:1989",
			"iso_8859-9":  "ISO_8859-9:1989",
			"iso-8859-9":  "ISO_8859-9:1989",
			"iso8859-9":   "ISO_8859-9:1989",
			"iso88599":    "ISO_8859-9:1989",
			"iso8859-9e":  "ISO_8859-9:1989",
			"iso-8859-9e": "ISO_8859-9:1989",
			"iso_8859-9e": "ISO_8859-9:1989",
			"iso88599e":   "ISO_8859-9:1989",
			"latin5":      "ISO_8859-9:1989",
			"l5":          "ISO_8859-9:1989",
			"csisolatin5": "ISO_8859-9:1989",
			"8859_9":      "ISO_8859-9:1989",
			"ibm920":      "ISO_8859-9:1989",
			"csibm920":    "ISO_8859-9:1989",
			"ibm-920":     "ISO_8859-9:1989",
			"cp920":       "ISO_8859-9:1989",
			"920":         "ISO_8859-9:1989",

			"iso-ir-157":       "ISO-8859-10",
			"l6":               "ISO-8859-10",
			"iso_8859-10:1992": "ISO-8859-10",
			"iso_8859-10":      "ISO-8859-10",
			"csisolatin6":      "ISO-8859-10",
			"latin6":           "ISO-8859-10",

			"latin7": "ISO-8859-13",
			"l7":     "ISO-8859-13",

			"ms_kanji":       "Shift_JIS",
			"csshiftjis":     "Shift_JIS",
			"sjis":           "Shift_JIS",
			"pck":            "Shift_JIS",
			"windows-31j":    "Shift_JIS",
			"ibm943":         "Shift_JIS",
			"csibm943":       "Shift_JIS",
			"ibm-943":        "Shift_JIS",
			"cp943":          "Shift_JIS",
			"943":            "Shift_JIS",
			"shift_jisx0213": "Shift_JIS",

			"cseucpkdfmtjapanese": "EUC-JP",
			"euc-jp":              "EUC-JP",
			"eucjp":               "EUC-JP",
			"eucjis":              "EUC-JP",
			"euc-jis":             "EUC-JP",
			"eucjisx0213":         "EUC-JP",
			"euc-jisx0213":        "EUC-JP",
			"extended_unix_code_packed_format_for_japanese": "EUC-JP",
			"x-euc-jp": "EUC-JP",
			"x-eucjp":  "EUC-JP",

			"iso-ir-149":     "KS_C_5601-1987",
			"ks_c_5601-1989": "KS_C_5601-1987",
			"ksc_5601":       "KS_C_5601-1987",
			"korean":         "KS_C_5601-1987",
			"csksc56011987":  "KS_C_5601-1987",
			"windows-949":    "KS_C_5601-1987",
			"ks_c_5601-1987": "KS_C_5601-1987",
			"ksc5601-1987":   "KS_C_5601-1987",
			"ksc5601_1987":   "KS_C_5601-1987",
			"5601":           "KS_C_5601-1987",
			"ibm949":         "KS_C_5601-1987",
			"csibm949":       "KS_C_5601-1987",
			"ibm-949":        "KS_C_5601-1987",
			"cp949":          "KS_C_5601-1987",
			"949":            "KS_C_5601-1987",

			"csiso2022kr": "ISO-2022-KR",
			"iso2022kr":   "ISO-2022-KR",
			"iso_2022_kr": "ISO-2022-KR",

			"cseuckr": " EUC-KR",
			"ksc5601": " EUC-KR",
			"euckr":   " EUC-KR",

			"csiso2022jp":   "ISO-2022-JP",
			"jis":           "ISO-2022-JP",
			"iso-2022-jp":   "ISO-2022-JP",
			"iso-2022-jp-3": "ISO-2022-JP",
			"iso2022jp":     "ISO-2022-JP",
			"iso2022jp2":    "ISO-2022-JP",
			"jis_encoding":  "ISO-2022-JP",
			"csjisencoding": "ISO-2022-JP",

			"latingreek1":        "Latin-greek-1",
			"iso-ir-27":          "Latin-greek-1",
			"csiso27latingreek1": "Latin-greek-1",

			"iso-ir-58":        "GB_2312-80",
			"chinese":          "GB_2312-80",
			"csiso58gb231280":  "GB_2312-80",
			"gb2312 gb2312-80": "GB_2312-80",
			"gb2312-1980":      "GB_2312-80",
			"euc-cn":           "GB_2312-80",
			"euccn":            "GB_2312-80",

			"iso-ir-111":           "ECMA-cyrillic",
			"koi8-e":               "ECMA-cyrillic",
			"csiso111ecmacyrillic": "ECMA-cyrillic",
			"ecmacyrillic":         "ECMA-cyrillic",

			"csiso2022cn": "ISO-2022-CN",
			"iso2022cn":   "ISO-2022-CN",

			"csiso2022cnext": "ISO-2022-CN-EXT",
			"iso2022cnext":   "ISO-2022-CN-EXT",

			"csutf8":            "UTF-8",
			"unicode-1-1-utf-8": "UTF-8",
			"unicode11utf8":     "UTF-8",
			"unicode20utf8":     "UTF-8",
			"utf8":              "UTF-8",
			"x-unicode20utf8":   "UTF-8",

			"csiso885913": "ISO-8859-13",

			"iso-ir-199":       "ISO-8859-14",
			"iso_8859-14:1998": "ISO-8859-14",
			"iso_8859-14":      "ISO-8859-14",
			"latin8":           "ISO-8859-14",
			"iso-celtic":       "ISO-8859-14",
			"l8":               "ISO-8859-14",
			"csiso885914":      "ISO-8859-14",

			"iso_8859-15":      "ISO-8859-15",
			"iso_8859-15:1998": "ISO-8859-15",
			"latin-9":          "ISO-8859-15",
			"latin9":           "ISO-8859-15",
			"csiso885915":      "ISO-8859-15",

			"iso-ir-226":       "ISO-8859-16",
			"iso_8859-16:2001": "ISO-8859-16",
			"iso_8859-16":      "ISO-8859-16",
			"latin10":          "ISO-8859-16",
			"l10":              "ISO-8859-16",
			"csiso885916":      "ISO-8859-16",
			"iso885916":        "ISO-8859-16",
			"iso8859-16":       "ISO-8859-16",

			"cp936":       "GBK",
			"ms936":       "GBK",
			"windows-936": "GBK",
			"csgbk":       "GBK",

			"csgb18030": "GB18030",

			"csutf7": "UTF-7",
			"utf7":   "UTF-7",

			"csutf7imap": "UTF-7-IMAP",

			"unicodefffe": "UTF-16BE",
			"csutf16be":   "UTF-16BE",
			"utf16be":     "UTF-16BE",
			"unicodebig":  "UTF-16BE",
			"unicode-1-1": "UTF-16BE",
			"x-utf-16be":  "UTF-16BE",

			"ucs-2":         "UTF-16LE",
			"unicodefeff":   "UTF-16LE",
			"csutf16le":     "UTF-16LE",
			"utf16le":       "UTF-16LE",
			"unicodelittle": "UTF-16LE",
			"x-utf-16le":    "UTF-16LE",

			"csutf16": "UTF-16",
			"utf16":   "UTF-16",

			"csutf32": "UTF-32",
			"utf32":   "UTF-32",

			"csutf32be": "UTF-32BE",
			"utf32be":   "UTF-32BE",

			"csutf32le": "UTF-32LE",
			"utf32le":   "UTF-32LE",

			"cswindows31j": "Windows-31J",

			"csgb2312":   "GB2312",
			"hz-gb-2312": "GB2312",

			"csbig5":   "Big5",
			"big-5":    "Big5",
			"big-five": "Big5",
			"bigfive":  "Big5",
			"ibm950":   "Big5",
			"csibm950": "Big5",
			"ibm-950":  "Big5",
			"cp950":    "Big5",
			"950":      "Big5",

			"mac":         "macintosh",
			"csmacintosh": "macintosh",

			"cskoi8r": "KOI8-R",
			"koi8r":   "KOI8-R",

			"cp866":    "IBM866",
			"866":      "IBM866",
			"csibm866": "IBM866",
			"ibm-866":  "IBM866",
			"ibm866":   "IBM866",

			"cskoi8u": "KOI8-U",
			"koi8u":   "KOI8-U",

			"csbig5hkscs": "Big5-HKSCS",
			"big5hkscs":   "Big5-HKSCS",

			"ami1251":     "Amiga-1251",
			"amiga1251":   "Amiga-1251",
			"ami-1251":    "Amiga-1251",
			"csamiga1251": "Amiga-1251",

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

			"cstis620":      "TIS-620",
			"iso-8859-11":   "TIS-620",
			"tis620.2533":   "TIS-620",
			"tis620.2533-0": "TIS-620",
			"ibm874":        "TIS-620",
			"csibm874":      "TIS-620",
			"ibm-874":       "TIS-620",
			"cp874":         "TIS-620",
			"874":           "TIS-620",
			"windows-28601": "TIS-620",
			"tis620-0":      "TIS-620",
			"tis620":        "TIS-620",

			"iso5428":        "ISO_5428:1980",
			"iso-5428":       "ISO_5428:1980",
			"iso_5428":       "ISO_5428:1980",
			"iso-ir-55":      "ISO_5428:1980",
			"csiso5428greek": "ISO_5428:1980",

			"iso_2033":  "ISO_2033-1983",
			"iso2033":   "ISO_2033-1983",
			"iso-2033":  "ISO_2033-1983",
			"iso-ir-98": "ISO_2033-1983",
			"e13b":      "ISO_2033-1983",
			"csiso2033": "ISO_2033-1983",

			"iso11548-1":     "ISO-11548-1",
			"iso_11548-1":    "ISO-11548-1",
			"iso_tr_11548-1": "ISO-11548-1",
			"csiso115481":    "ISO-11548-1",
			"iso/tr_11548-1": "ISO-11548-1",

			"csiso10646utf1":  "ISO-10646-UTF-1",
			"iso-10646/utf-8": "ISO-10646-UTF-1",
			"iso-10646/utf8":  "ISO-10646-UTF-1",

			"csucs4":            "ISO-10646-UCS-4",
			"iso-10646/ucs4":    "ISO-10646-UCS-4",
			"10646-1:1993/ucs4": "ISO-10646-UCS-4",
			"10646-1:1993":      "ISO-10646-UCS-4",
			"ucs4":              "ISO-10646-UCS-4",
			"ucs-4":             "ISO-10646-UCS-4",

			"csunicode":      "ISO-10646-UCS-2",
			"unicode":        "ISO-10646-UCS-2",
			"iso-10646/ucs2": "ISO-10646-UCS-2",
			"ucs2":           "ISO-10646-UCS-2",

			"iso-ir-18":        "greek7-old",
			"csiso18greek7old": "greek7-old",
			"greek7old":        "greek7-old",

			"iso-ir-19":         "latin-greek",
			"csiso19latingreek": "latin-greek",
			"latingreek":        "latin-greek",

			"iso-ir-57":     "GB_1988-80",
			"cn":            "GB_1988-80",
			"iso646-cn":     "GB_1988-80",
			"csiso57gb1988": "GB_1988-80",
			"gb_198880":     "GB_1988-80",
			"cn-gb":         "GB_1988-80",

			"iso-ir-37":             "ISO_5427",
			"csiso5427cyrillic":     "ISO_5427",
			"csiso5427cyrillic1981": "ISO_5427",

			"iso-ir-155":    "ISO_10367-box",
			"csiso10367box": "ISO_10367-box",
			"iso10367box":   "ISO_10367-box",
			"iso_10367box":  "ISO_10367-box",

			"iso-ir-14":         "JIS_C6220-1969-ro",
			"jp":                "JIS_C6220-1969-ro",
			"iso646-jp":         "JIS_C6220-1969-ro",
			"csiso14jisc6220ro": "JIS_C6220-1969-ro",
			"jis_c6220ro":       "JIS_C6220-1969-ro",

			"iso-ir-92":            "JIS_C6229-1984-b",
			"iso646-jp-ocr-b":      "JIS_C6229-1984-b",
			"jp-ocr-b":             "JIS_C6229-1984-b",
			"csiso92jisc62991984b": "JIS_C6229-1984-b",
			"jis_c62991984b":       "JIS_C6229-1984-b",
		},
		}
	})
}

func getEncodingMapping() ICharsetEncodingMapping {
	initialiseMapping()
	return &mapping
}
