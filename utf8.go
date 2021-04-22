package jsonvalue

//#include <string.h>
import "C"

import (
	"errors"
	"fmt"
	"unsafe"
)

// utf8Iter is used to iterate []byte text
type utf8Iter struct {
	b []byte
}

// parseStrFromBytes
func (it *utf8Iter) parseStrFromBytes(offset, length int) (resLen int, err error) {
	i := offset
	end := i + length
	sectEnd := offset

	shift := func(i *int, le int) {
		if end-*i < le {
			err = errors.New("illegal UTF8 string")
			return
		}
		it.memcpy(sectEnd, *i, le)
		sectEnd += le
		*i += le
	}

	// iterate every byte
	for i := offset; i < offset+length; {
		chr := it.b[i]

		// ACSII?
		if chr <= 0x7F {
			if chr == '\\' {
				err = it.handleEscapeStart(&i, end, &sectEnd)
			} else {
				i++
				sectEnd++
			}
		} else if runeIdentifyingBytes2(chr) {
			shift(&i, 2)
		} else if runeIdentifyingBytes3(chr) {
			shift(&i, 3)
		} else if runeIdentifyingBytes4(chr) {
			shift(&i, 4)
		} else {
			err = errors.New("illegal UTF8 string")
		}

		if err != nil {
			return -1, err
		}
	}

	return sectEnd - offset, nil
}

func (it utf8Iter) handleEscapeStart(i *int, end int, sectEnd *int) error {
	if end-*i < 1 {
		return errors.New("escape symbol not followed by another character")
	}
	chr := it.b[*i+1]
	switch chr {
	default:
		return fmt.Errorf("unreconized character 0x%02X after escape symbol", chr)
	case '"', '\'', '/', '\\':
		it.b[*sectEnd] = chr
		*sectEnd++
		*i += 2
	case 'b':
		it.b[*sectEnd] = '\b'
		*sectEnd++
		*i += 2
	case 'f':
		it.b[*sectEnd] = '\f'
		*sectEnd++
		*i += 2
	case 'r':
		it.b[*sectEnd] = '\r'
		*sectEnd++
		*i += 2
	case 'n':
		it.b[*sectEnd] = '\n'
		*sectEnd++
		*i += 2
	case 't':
		it.b[*sectEnd] = '\t'
		*sectEnd++
		*i += 2
	case 'u':
		return it.handleEscapeUnicodeStart(i, end, sectEnd)
	}
	return nil
}

func (it *utf8Iter) handleEscapeUnicodeStart(i *int, end int, sectEnd *int) (err error) {
	if end-*i <= 5 {
		return errors.New("escape symbol not followed by another character")
	}

	b3 := chrToHex(it.b[*i+2], &err)
	b2 := chrToHex(it.b[*i+3], &err)
	b1 := chrToHex(it.b[*i+4], &err)
	b0 := chrToHex(it.b[*i+5], &err)
	if err != nil {
		return
	}

	r := (rune(b3) << 12) + (rune(b2) << 8) + (rune(b1) << 4) + rune(b0)

	// this rune is smaller than 0x10000
	if r <= 0xD7FF || r >= 0xE000 {
		le := it.assignWideRune(*sectEnd, r)
		*i += 6
		*sectEnd += le
		return nil
	}

	// reference: [JSON 序列化中的转义和 Unicode 编码](https://cloud.tencent.com/developer/article/1625557/)
	// should get another unicode-escaped character
	if end-*i <= 11 {
		return fmt.Errorf("insufficient UTF-16 data at offset %d", *i)
	}
	if it.b[*i+6] != '\\' || it.b[*i+7] != 'u' {
		return fmt.Errorf("expect unicode escape character at position %d but not", *i+6)
	}

	ex3 := chrToHex(it.b[*i+8], &err)
	ex2 := chrToHex(it.b[*i+9], &err)
	ex1 := chrToHex(it.b[*i+10], &err)
	ex0 := chrToHex(it.b[*i+11], &err)
	if err != nil {
		return
	}

	ex := (rune(ex3) << 12) + (rune(ex2) << 8) + (rune(ex1) << 4) + rune(ex0)
	ex -= 0xDC00
	if ex > 0x03FF {
		return fmt.Errorf("expect second UTF-16 encoding but got 0x04%X at position %d", r, *i+8)
	}

	r = ((r - 0xD800) << 10) + ex + 0x10000

	le := it.assignWideRune(*sectEnd, r)
	*i += 12
	*sectEnd += le
	return nil
}

func chrToHex(chr byte, errOut *error) byte {
	if chr >= '0' && chr <= '9' {
		return chr - '0'
	}
	if chr >= 'a' && chr <= 'z' {
		return chr - 'a' + 10
	}
	if chr >= 'A' && chr <= 'Z' {
		return chr - 'A' + 10
	}
	*errOut = fmt.Errorf("invalid unicode value character: %c", rune(chr))
	return 0
}

func (it *utf8Iter) memcpy(dst, src, length int) {
	if dst == src {
		return
	}
	ptr := unsafe.Pointer(&it.b[0])
	C.memcpy(
		unsafe.Pointer(uintptr(ptr)+uintptr(dst)),
		unsafe.Pointer(uintptr(ptr)+uintptr(src)),
		C.size_t(length),
	)
}

func (it *utf8Iter) assignWideRune(dst int, r rune) (offset int) {
	// 00000yyy yyzzzzzz ==>
	// 110yyyyy 10zzzzzz
	if r <= 0x7FF {
		b0 := byte((r&0x7C0)>>6) + 0xC0
		b1 := byte((r&0x03F)>>0) + 0x80
		it.b[dst+0] = b0
		it.b[dst+1] = b1
		return 2
	}

	// xxxxyyyy yyzzzzzz ==>
	// 1110xxxx 10yyyyyy 10zzzzzz
	if r <= 0xFFFF {
		b0 := byte((r&0xF000)>>12) + 0xE0
		b1 := byte((r&0x0FC0)>>6) + 0x80
		b2 := byte((r&0x003F)>>0) + 0x80
		it.b[dst+0] = b0
		it.b[dst+1] = b1
		it.b[dst+2] = b2
		return 3
	}

	// 000wwwxx xxxxyyyy yyzzzzzz ==>
	// 11110www 10xxxxxx 10yyyyyy 10zzzzzz
	b0 := byte((r&0x1C0000)>>18) + 0xF0
	b1 := byte((r&0x03F000)>>12) + 0x80
	b2 := byte((r&0x000FC0)>>6) + 0x80
	b3 := byte((r&0x00003F)>>0) + 0x80
	it.b[dst+0] = b0
	it.b[dst+1] = b1
	it.b[dst+2] = b2
	it.b[dst+3] = b3
	return 4
}

func runeIdentifyingBytes2(chr byte) bool {
	return (chr & 0xC0) == 0xC0
}

func runeIdentifyingBytes3(chr byte) bool {
	return (chr & 0xE0) == 0xE0
}

func runeIdentifyingBytes4(chr byte) bool {
	return (chr & 0xF8) == 0xF8
}
