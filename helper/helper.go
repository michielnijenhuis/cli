package helper

import (
	"fmt"
	"io"
	"strconv"
	"strings"
	"unicode/utf8"

	"golang.org/x/text/encoding"
	"golang.org/x/text/encoding/charmap"
	"golang.org/x/text/transform"
)

func Width(str string) int {
	length := 0

	for len(str) > 0 {
		r, size := utf8.DecodeRuneInString(str)
		cp := int(r)
		str = str[size:]

		switch {
		case 0x1100 <= cp && cp <= 0x115f,
			0x11a3 <= cp && cp <= 0x11a7,
			0x11fa <= cp && cp <= 0x11ff,
			0x2329 <= cp && cp <= 0x232a,
			0x2e80 <= cp && cp <= 0x2e99,
			0x2e9b <= cp && cp <= 0x2ef3,
			0x2f00 <= cp && cp <= 0x2fd5,
			0x2ff0 <= cp && cp <= 0x2ffb,
			0x3000 <= cp && cp <= 0x303e,
			0x3041 <= cp && cp <= 0x3096,
			0x3099 <= cp && cp <= 0x30ff,
			0x3105 <= cp && cp <= 0x312d,
			0x3131 <= cp && cp <= 0x318e,
			0x3190 <= cp && cp <= 0x31ba,
			0x31c0 <= cp && cp <= 0x31e3,
			0x31f0 <= cp && cp <= 0x321e,
			0x3220 <= cp && cp <= 0x3247,
			0x3250 <= cp && cp <= 0x32fe,
			0x3300 <= cp && cp <= 0x4dbf,
			0x4e00 <= cp && cp <= 0xa48c,
			0xa490 <= cp && cp <= 0xa4c6,
			0xa960 <= cp && cp <= 0xa97c,
			0xac00 <= cp && cp <= 0xd7a3,
			0xd7b0 <= cp && cp <= 0xd7c6,
			0xd7cb <= cp && cp <= 0xd7fb,
			0xf900 <= cp && cp <= 0xfaff,
			0xfe10 <= cp && cp <= 0xfe19,
			0xfe30 <= cp && cp <= 0xfe52,
			0xfe54 <= cp && cp <= 0xfe66,
			0xfe68 <= cp && cp <= 0xfe6b,
			0xff01 <= cp && cp <= 0xff60,
			0xffe0 <= cp && cp <= 0xffe6,
			0x1b000 <= cp && cp <= 0x1b001,
			0x1f200 <= cp && cp <= 0x1f202,
			0x1f210 <= cp && cp <= 0x1f23a,
			0x1f240 <= cp && cp <= 0x1f248,
			0x1f250 <= cp && cp <= 0x1f251,
			0x20000 <= cp && cp <= 0x2fffd,
			0x30000 <= cp && cp <= 0x3fffd:
			length += 2
		default:
			length += 1
		}
	}

	return length
}

func Len(str string) int {
	if utf8.ValidString(str) {
		return utf8.RuneCountInString(str)
	}

	encoding, err := detectEncoding(str)
	if err != nil || encoding == nil {
		return len(str)
	}

	decoded, _, _ := transform.String(encoding.NewDecoder(), str)
	return utf8.RuneCountInString(decoded)
}

func detectEncoding(s string) (encoding.Encoding, error) {
	for _, enc := range []encoding.Encoding{
		charmap.ISO8859_1, charmap.Windows1252,
	} {
		decoded, _, err := transform.String(enc.NewDecoder(), s)
		if err == nil && utf8.ValidString(decoded) {
			return enc, nil
		}
	}

	return nil, io.EOF
}

func Substring(s string, from int, length int) string {
	if s == "" {
		return s
	}

	return s[from:length]
}

func FormatMemory(memory int) string {
	format := fmt.Sprintf("%%.%df %%s", 1)

	if memory >= 1024*1024*1024 {
		return fmt.Sprintf(format, memory/1024/1024/1024, "GiB")
	}

	if memory >= 1024*1024 {
		return fmt.Sprintf(format, memory/1024/1024, "MiB")
	}

	if memory >= 1024 {
		return fmt.Sprintf(format, memory/1024, "KiB")
	}

	return fmt.Sprintf("%d B", memory)
}

var timeFormats = [][]string{
	{"1", "1 sec", "secs"},
	{"60", "1 min", "mins"},
	{"3600", "1 hr", "hrs"},
	{"86400", "1 day", "days"},
}

func FormatTime(secs int, precision int) string {
	if secs <= 0 {
		return "< 1 sec"
	}

	times := make(map[int]string)

	for i, format := range timeFormats {
		seconds := secs
		if i+1 < len(timeFormats) {
			conv, _ := strconv.Atoi(timeFormats[i+1][0])
			seconds = secs % conv
		}

		delete(times, i-precision)

		if seconds == 0 {
			continue
		}

		f, _ := strconv.Atoi(format[0])
		unitCount := seconds / f

		var t string
		if unitCount == 1 {
			t = format[1]
		} else {
			t = fmt.Sprintf("%d %s", unitCount, format[2])
		}
		times[i] = t

		if secs == seconds {
			continue
		}

		secs -= seconds
	}

	values := make([]string, len(times))
	i := 0
	for k := range times {
		values[i] = fmt.Sprint(k)
		i++
	}

	return strings.Join(values, ", ")
}
