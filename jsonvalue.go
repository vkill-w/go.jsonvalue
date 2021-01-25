// Package jsonvalue is for JSON parsing and setting. It is used in situations those Go structures cannot achieve, or "map[string]interface{}" could not do properbally.
//
// As a quick start:
// 	v := jsonvalue.NewObject()
// 	v.SetString("Hello, JSON").At("someObject", "someObject", "someObject", "message")  // automatically create sub objects
// 	fmt.Println(v.MustMarshalString())                                                  // marshal to string type
// 	// Output:
// 	// {"someObject":{"someObject":{"someObject":{"message":"Hello, JSON!"}}}
// If you want to parse raw JSON data, use Unmarshal()
// 	raw := []byte(`{"message":"hello, world"}`)
// 	v, err := jsonvalue.Unmarshal(raw)
// 	s, _ := v.GetString("message")
// 	fmt.Println(s)
// 	// Output:
// 	// hello, world
package jsonvalue

import (
	"bytes"
	"container/list"
	"fmt"
	"reflect"
	"unsafe"

	"github.com/buger/jsonparser"
)

// V is the main type of jsonvalue, representing a JSON value.
type V struct {
	valueType  jsonparser.ValueType
	valueBytes []byte

	status struct {
		parsed   bool
		negative bool
		floated  bool
	}

	value struct {
		str     string
		i64     int64
		u64     uint64
		boolean bool
		f64     float64
	}

	children struct {
		object map[string]*V
		array  *list.List
	}
}

func new() *V {
	v := V{}
	v.valueType = jsonparser.NotExist
	return &v
}

func newObject() *V {
	v := V{}
	v.valueType = jsonparser.Object
	v.children.object = make(map[string]*V)
	return &v
}

func newArray() *V {
	v := V{}
	v.valueType = jsonparser.Array
	v.children.array = list.New()
	return &v
}

// UnmarshalString is equavilent to Unmarshal(string(b))
func UnmarshalString(s string) (*V, error) {
	// reference: https://stackoverflow.com/questions/41591097/slice-bounds-out-of-range-when-using-unsafe-pointer
	sh := (*reflect.StringHeader)(unsafe.Pointer(&s))
	bh := reflect.SliceHeader{
		Data: sh.Data,
		Len:  sh.Len,
		Cap:  sh.Len,
	}
	b := *(*[]byte)(unsafe.Pointer(&bh))
	return Unmarshal(b)
}

// Unmarshal parse raw bytes(encoded in UTF-8 or pure AscII) and returns a *V instance.
func Unmarshal(b []byte) (ret *V, err error) {
	if nil == b || 0 == len(b) {
		return nil, ErrNilParameter
	}

	for i, c := range b {
		switch c {
		case ' ', '\r', '\n', '\t', '\b':
			// continue
		case '{':
			// object start
			return newFromObject(b[i:])
		case '[':
			return newFromArray(b[i:])
		case '0', '1', '2', '3', '4', '5', '6', '7', '8', '9', '-':
			ret, err = newFromNumber(b[i:])
			if err != nil {
				return
			}
			err = ret.parseNumber()
			if err != nil {
				return nil, err
			}
			return ret, nil

		case '"':
			ret = new()
			ret.valueType = jsonparser.String
			ret.value.str, ret.valueBytes, err = parseString(b[i:])
			if err != nil {
				return nil, err
			}
			ret.status.parsed = true
			return ret, nil

		case 't':
			return newFromTrue(b[i:])
		case 'f':
			return newFromFalse(b[i:])
		case 'n':
			return newFromNull(b[i:])
		default:
			return nil, ErrRawBytesUnrecignized
		}
	}

	return nil, ErrRawBytesUnrecignized
}

func (v *V) parseNumber() (err error) {
	b := v.valueBytes

	if bytes.Contains(b, []byte(".")) {
		v.status.floated = true
		v.value.f64, err = parseFloat(b)
		if err != nil {
			return
		}

		v.status.parsed = true
		v.status.negative = (v.value.f64 < 0)
		v.value.i64 = int64(v.value.f64)
		v.value.u64 = uint64(v.value.f64)

	} else if '-' == b[0] {
		v.status.negative = true
		v.value.i64, err = parseInt(b)
		if err != nil {
			return
		}

		v.status.parsed = true
		v.value.u64 = uint64(v.value.i64)
		v.value.f64 = float64(v.value.i64)

	} else {
		v.status.negative = false
		v.value.u64, err = parseUint(b)
		if err != nil {
			return
		}

		v.status.parsed = true
		v.value.i64 = int64(v.value.u64)
		v.value.f64 = float64(v.value.u64)
	}

	return nil
}

// ==== simple object parsing ====
func newFromNumber(b []byte) (ret *V, err error) {
	v := new()
	v.valueType = jsonparser.Number
	v.valueBytes = b
	return v, nil
}

// func newFromString(b []byte) (ret *V, err error) {
// 	v := new()
// 	v.valueType = jsonparser.String
// 	v.valueBytes = b
// 	return v, nil
// }

func newFromTrue(b []byte) (ret *V, err error) {
	if len(b) != 4 || string(b) != "true" {
		return nil, ErrNotValidBoolValue
	}
	v := new()
	v.status.parsed = true
	v.valueType = jsonparser.Boolean
	v.valueBytes = []byte{'t', 'r', 'u', 'e'}
	v.value.boolean = true
	return v, nil
}

func newFromFalse(b []byte) (ret *V, err error) {
	if len(b) != 5 || string(b) != "false" {
		return nil, ErrNotValidBoolValue
	}
	v := new()
	v.status.parsed = true
	v.valueType = jsonparser.Boolean
	v.valueBytes = []byte{'f', 'a', 'l', 's', 'e'}
	v.value.boolean = false
	return v, nil
}

func newFromBool(b []byte) (ret *V, err error) {
	v := new()
	v.valueType = jsonparser.Boolean

	switch string(b) {
	case "true":
		v.status.parsed = true
		v.valueBytes = []byte{'t', 'r', 'u', 'e'}
		v.value.boolean = true
	case "false":
		v.status.parsed = true
		v.valueBytes = []byte{'f', 'a', 'l', 's', 'e'}
		v.value.boolean = false
	default:
		return nil, ErrNotValidBoolValue
	}

	return v, nil
}

func newFromNull(b []byte) (ret *V, err error) {
	if len(b) != 4 || string(b) != "null" {
		return nil, ErrNotValidBoolValue
	}
	v := new()
	v.status.parsed = true
	v.valueType = jsonparser.Null
	return v, nil
}

// ====
func newFromArray(b []byte) (ret *V, err error) {
	o := newArray()

	jsonparser.ArrayEach(b, func(v []byte, t jsonparser.ValueType, _ int, _ error) {
		if err != nil {
			return
		}

		var child *V

		switch t {
		default:
			err = fmt.Errorf("invalid value type: %v", t)
		case jsonparser.Object:
			child, err = newFromObject(v)
		case jsonparser.Array:
			child, err = newFromArray(v)
		case jsonparser.Number:
			child, err = newFromNumber(v)
		case jsonparser.Boolean:
			child, err = newFromBool(v)
		case jsonparser.Null:
			child, err = newFromNull(v)
		case jsonparser.String:
			s, err := parseStringNoQuote(v)
			if err != nil {
				return
			}
			child = new()
			child.status.parsed = true
			child.valueType = jsonparser.String
			child.value.str = s
		}

		if err != nil {
			return
		}
		o.children.array.PushBack(child)
		return
	})

	// done
	if err != nil {
		return
	}
	return o, nil
}

// ==== object parsing ====
func newFromObject(b []byte) (ret *V, err error) {
	o := newObject()

	err = jsonparser.ObjectEach(b, func(k, v []byte, t jsonparser.ValueType, _ int) error {
		// key
		var child *V
		key, err := parseStringNoQuote(k)
		if err != nil {
			return err
		}

		switch t {
		default:
			return fmt.Errorf("invalid value type: %v", t)
		case jsonparser.Object:
			child, err = newFromObject(v)
		case jsonparser.Array:
			child, err = newFromArray(v)
		case jsonparser.Number:
			child, err = newFromNumber(v)
		case jsonparser.Boolean:
			child, err = newFromBool(v)
		case jsonparser.Null:
			child, err = newFromNull(v)
		case jsonparser.String:
			s, err := parseStringNoQuote(v)
			if err != nil {
				return err
			}
			child = new()
			child.status.parsed = true
			child.valueType = jsonparser.String
			child.value.str = s
		}

		if err != nil {
			return err
		}
		o.children.object[key] = child
		return nil
	})

	// done
	if err != nil {
		return
	}
	return o, nil
}

// ==== type access ====

// IsObject tells whether value is an object
func (v *V) IsObject() bool {
	return v.valueType == jsonparser.Object
}

// IsArray tells whether value is an array
func (v *V) IsArray() bool {
	return v.valueType == jsonparser.Array
}

// IsString tells whether value is a string
func (v *V) IsString() bool {
	return v.valueType == jsonparser.String
}

// IsNumber tells whether value is a number
func (v *V) IsNumber() bool {
	return v.valueType == jsonparser.Number
}

// IsFloat tells whether value is a float point number
func (v *V) IsFloat() bool {
	if v.valueType != jsonparser.Number {
		return false
	}
	if false == v.status.parsed {
		v.parseNumber()
	}
	return v.status.floated
}

// IsInteger tells whether value is a fix point interger
func (v *V) IsInteger() bool {
	if v.valueType != jsonparser.Number {
		return false
	}
	if false == v.status.parsed {
		err := v.parseNumber()
		if err != nil {
			return false
		}
	}
	return !(v.status.floated)
}

// IsNegative tells whether value is a negative number
func (v *V) IsNegative() bool {
	if v.valueType != jsonparser.Number {
		return false
	}
	if false == v.status.parsed {
		v.parseNumber()
	}
	return v.status.negative
}

// IsPositive tells whether value is a positive number
func (v *V) IsPositive() bool {
	if v.valueType != jsonparser.Number {
		return false
	}
	if false == v.status.parsed {
		err := v.parseNumber()
		if err != nil {
			return false
		}
	}
	return !(v.status.negative)
}

// GreaterThanInt64Max return true when ALL conditions below are met:
// 	1. It is a number value.
// 	2. It is a positive interger.
// 	3. Its value is greater than 0x7fffffffffffffff.
func (v *V) GreaterThanInt64Max() bool {
	if v.valueType != jsonparser.Number {
		return false
	}
	if false == v.status.parsed {
		v.parseNumber()
	}
	if v.status.negative {
		return false
	}
	return v.value.u64 > 0x7fffffffffffffff
}

// IsBoolean tells whether value is a boolean
func (v *V) IsBoolean() bool {
	return v.valueType == jsonparser.Boolean
}

// IsNull tells whether value is a null
func (v *V) IsNull() bool {
	return v.valueType == jsonparser.Null
}

// ==== value access ====

// Bool returns represented bool value. If value is not boolean, returns false.
func (v *V) Bool() bool {
	return v.value.boolean
}

// Int returns represented int value. If value is not a number, returns zero.
func (v *V) Int() int {
	if v.valueType != jsonparser.Number {
		return 0
	}
	if false == v.status.parsed {
		v.parseNumber()
	}
	return int(v.value.i64)
}

// Uint returns represented uint value. If value is not a number, returns zero.
func (v *V) Uint() uint {
	if v.valueType != jsonparser.Number {
		return 0
	}
	if false == v.status.parsed {
		v.parseNumber()
	}
	return uint(v.value.u64)
}

// Int64 returns represented int64 value. If value is not a number, returns zero.
func (v *V) Int64() int64 {
	if v.valueType != jsonparser.Number {
		return 0
	}
	if false == v.status.parsed {
		v.parseNumber()
	}
	return int64(v.value.i64)
}

// Uint64 returns represented uint64 value. If value is not a number, returns zero.
func (v *V) Uint64() uint64 {
	if v.valueType != jsonparser.Number {
		return 0
	}
	if false == v.status.parsed {
		v.parseNumber()
	}
	return uint64(v.value.u64)
}

// Int32 returns represented int32 value. If value is not a number, returns zero.
func (v *V) Int32() int32 {
	if v.valueType != jsonparser.Number {
		return 0
	}
	if false == v.status.parsed {
		v.parseNumber()
	}
	return int32(v.value.i64)
}

// Uint32 returns represented uint32 value. If value is not a number, returns zero.
func (v *V) Uint32() uint32 {
	if v.valueType != jsonparser.Number {
		return 0
	}
	if false == v.status.parsed {
		v.parseNumber()
	}
	return uint32(v.value.u64)
}

// Float64 returns represented float64 value. If value is not a number, returns zero.
func (v *V) Float64() float64 {
	if v.valueType != jsonparser.Number {
		return 0.0
	}
	if false == v.status.parsed {
		v.parseNumber()
	}
	return v.value.f64
}

// Float32 returns represented float32 value. If value is not a number, returns zero.
func (v *V) Float32() float32 {
	if v.valueType != jsonparser.Number {
		return 0.0
	}
	if false == v.status.parsed {
		v.parseNumber()
	}
	return float32(v.value.f64)
}

// String returns represented string value or the description for the jsonvalue.V instance if it is not a string.
func (v *V) String() string {
	switch v.valueType {
	default:
		return ""
	case jsonparser.Null:
		return "null"
	case jsonparser.Number:
		return string(v.valueBytes)
	case jsonparser.String:
		if false == v.status.parsed {
			var e error
			v.value.str, v.valueBytes, e = parseString(v.valueBytes)
			if nil == e {
				v.status.parsed = true
			}
		}
		return v.value.str
	case jsonparser.Boolean:
		return formatBool(v.value.boolean)
	case jsonparser.Object:
		return v.packObjChildren()
	case jsonparser.Array:
		return v.packArrChildren()
	}
}

func (v *V) packObjChildren() string {
	buf := bytes.Buffer{}
	v.bufObjChildren(&buf)
	return buf.String()
}

func (v *V) bufObjChildren(buf *bytes.Buffer) {
	buf.WriteByte('{')
	i := 0
	for k, v := range v.children.object {
		if i > 0 {
			buf.WriteString(", ")
		}
		buf.WriteString(k)
		buf.WriteString(": ")
		buf.WriteString(v.String())
		i++
	}
	buf.WriteByte('}')
}

func (v *V) packArrChildren() string {
	buf := bytes.Buffer{}
	v.bufArrChildren(&buf)
	return buf.String()
}

func (v *V) bufArrChildren(buf *bytes.Buffer) {
	buf.WriteByte('[')
	v.RangeArray(func(i int, v *V) bool {
		if i > 0 {
			buf.WriteString(", ")
		}
		buf.WriteString(v.String())
		i++
		return true
	})
	buf.WriteByte(']')
}
