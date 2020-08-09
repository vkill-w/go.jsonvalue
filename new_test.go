package jsonvalue

import (
	"os"
	"strconv"
	"testing"
)

func TestNewString(t *testing.T) {
	s := "你好，世界"
	v := NewString(s)
	if s != v.String() {
		t.Errorf("test NewString Failed")
	}
}

func TestNewBool(t *testing.T) {
	v := NewBool(true)
	if v.Bool() != true {
		t.Errorf("test NewBool - true failed")
		return
	}

	v = NewBool(false)
	if v.Bool() != false {
		t.Errorf("test NewBool - false failed")
		return
	}
}

func TestNewNull(t *testing.T) {
	v := NewNull()
	if false == v.IsNull() {
		t.Errorf("test NewNull failed")
		return
	}
}

func TestNewInteger(t *testing.T) {
	i := int64(-1234567)

	v := NewInt(int(i))
	if v.Int() != int(i) {
		t.Errorf("Test NewInt failed")
		return
	}

	v = NewUint(uint(i))
	if v.Uint() != uint(i) {
		t.Errorf("Test NewUint failed")
		return
	}

	v = NewInt32(int32(i))
	if v.Int32() != int32(i) {
		t.Errorf("Test NewInt32 failed")
		return
	}

	v = NewUint32(uint32(i))
	if v.Uint32() != uint32(i) {
		t.Errorf("Test NewUint32 failed")
		return
	}

	v = NewInt64(int64(i))
	if v.Int64() != int64(i) {
		t.Errorf("Test NewInt64 failed")
		return
	}

	v = NewUint64(uint64(i))
	if v.Uint64() != uint64(i) {
		t.Errorf("Test NewUint64 failed")
		return
	}
}

func TestNewFloat(t *testing.T) {
	s := "3.1415926535"
	f := 3.1415926535

	v := NewFloat64(f, -1)
	if v.String() != s {
		t.Errorf("Test NewFloat64 failed")
		return
	}

	v = NewFloat64(f, 2)
	if v.String() != "3.14" {
		t.Errorf("Test NewFloat64 failed")
		return
	}

	s = "3.1415927"
	v = NewFloat32(float32(f), -1)
	if v.String() != s {
		t.Errorf("Test NewFloat32 failed: %s", v.String())
		return
	}

	v = NewFloat32(float32(f), 5)
	if v.String() != "3.14159" {
		t.Errorf("Test NewFloat32 failed")
		return
	}
}

func TestEmptyObject(t *testing.T) {
	v := NewObject()
	b, _ := v.Marshal()
	if string(b) != "{}" {
		t.Errorf("Test NewObject failed")
		return
	}
}

func TestEmptyArray(t *testing.T) {
	v := NewArray()
	b := v.MustMarshal()
	if string(b) != "[]" {
		t.Errorf("TestNewObject failed")
		return
	}
}

func TestMiscValue(t *testing.T) {
	var err error
	var v *V
	var c *V
	checkErrMark := 0
	raw := ""
	topic := ""
	checkErr := func() {
		checkErrMark++
		if err != nil {
			t.Errorf("%02d - %s - error: %v", checkErrMark, topic, err)
			os.Exit(-1)
		}
	}
	checkCond := func(b bool) {
		if false == b {
			t.Errorf("%02d - %s - failed, object: %v", checkErrMark, topic, v)
			os.Exit(-1)
		}
	}

	topic = "parse array"
	raw = "\r\n[1, 2, 3 ]\t\b"
	v, err = UnmarshalString(raw)
	checkErr()
	checkCond(v.IsArray())

	topic = "parse object"
	raw = `{ }`
	v, err = UnmarshalString(raw)
	checkErr()
	checkCond(v.IsObject())

	topic = "parse string"
	raw = ` "hello, world"  `
	v, err = UnmarshalString(raw)
	checkErr()
	checkCond(v.IsString())
	checkCond(v.Int() == 0)
	checkCond(v.Uint() == 0)
	checkCond(v.Int64() == 0)
	checkCond(v.Uint64() == 0)
	checkCond(v.Int32() == 0)
	checkCond(v.Uint32() == 0)
	checkCond(v.Float64() == 0)
	checkCond(v.Float32() == 0)
	checkCond(v.IsFloat() == false)
	checkCond(v.IsInteger() == false)

	topic = "parse string with special character"
	raw = `"\"\\\/\f\t\r\n\b\u0030\uD87E\uDC04"`
	v, err = UnmarshalString(raw)
	checkErr()
	checkCond(v.IsString())

	topic = "parse null"
	raw = `null`
	v, err = UnmarshalString(raw)
	checkErr()
	checkCond(v.IsNull())

	topic = "parse bool (true)"
	raw = `true`
	v, err = UnmarshalString(raw)
	checkErr()
	checkCond(v.IsBoolean())
	checkCond(true == v.Bool())

	topic = "parse bool (true)"
	raw = `false`
	v, err = UnmarshalString(raw)
	checkErr()
	checkCond(v.IsBoolean())
	checkCond(false == v.Bool())

	topic = "parse negative float"
	raw = `-12345.12345`
	v, err = UnmarshalString(raw)
	checkErr()
	checkCond(v.IsNumber())
	checkCond(v.IsFloat())
	checkCond(v.IsNegative())
	checkCond(false == v.GreaterThanInt64Max())

	topic = "parse positive float"
	raw = `12345.12345`
	v, err = UnmarshalString(raw)
	checkErr()
	checkCond(v.IsNumber())
	checkCond(v.IsFloat())
	checkCond(v.IsPositive())
	checkCond(false == v.GreaterThanInt64Max())

	topic = "parse negative integer"
	raw = `-12345`
	v, err = UnmarshalString(raw)
	checkErr()
	checkCond(v.IsNumber())
	checkCond(v.IsInteger())
	checkCond(v.IsNegative())
	checkCond(false == v.GreaterThanInt64Max())

	topic = "parse positive integer"
	raw = `12345`
	v, err = UnmarshalString(raw)
	checkErr()
	checkCond(v.IsNumber())
	checkCond(v.IsInteger())
	checkCond(v.IsPositive())
	checkCond(false == v.GreaterThanInt64Max())

	topic = "parse big uint64"
	raw = `18446744073709551615` // 0xFFFFFFFFFFFFFFFF
	v, err = UnmarshalString(raw)
	checkErr()
	checkCond(v.IsNumber())
	checkCond(v.IsInteger())
	checkCond(v.IsPositive())
	checkCond(v.GreaterThanInt64Max())

	topic = "parse object in array"
	raw = `[{}]`
	v, err = UnmarshalString(raw)
	checkErr()
	checkCond(v.IsArray())
	c, err = v.Get(0)
	checkErr()
	checkCond(c.IsObject())

	topic = "parse array in object"
	raw = `{"array":[]}`
	v, err = UnmarshalString(raw)
	checkErr()
	checkCond(v.IsObject())
	c, err = v.Get("array")
	checkErr()
	checkCond(c.IsArray())

	topic = "parse float in object"
	raw = `{"float":123.4567}`
	v, err = UnmarshalString(raw)
	checkErr()
	v, err = v.Get("float")
	checkErr()
	checkCond(v.IsFloat())

	topic = "parse integer in object"
	raw = `{"int":123}`
	v, err = UnmarshalString(raw)
	checkErr()
	v, err = v.Get("int")
	checkErr()
	checkCond(v.IsInteger())

	topic = "parse int in object"
	raw = `{"int":123}`
	v, err = UnmarshalString(raw)
	checkErr()
	v, err = v.Get("int")
	checkErr()
	checkCond(v.Int() == 123)

	topic = "parse uint in object"
	raw = `{"uint":123}`
	v, err = UnmarshalString(raw)
	checkErr()
	v, err = v.Get("uint")
	checkErr()
	checkCond(v.Uint() == 123)

	topic = "parse int64 in object"
	raw = `{"int":123}`
	v, err = UnmarshalString(raw)
	checkErr()
	v, err = v.Get("int")
	checkErr()
	checkCond(v.Int64() == 123)

	topic = "parse uint64 in object"
	raw = `{"uint":123}`
	v, err = UnmarshalString(raw)
	checkErr()
	v, err = v.Get("uint")
	checkErr()
	checkCond(v.Uint64() == 123)

	topic = "parse int32 in object"
	raw = `{"int":123}`
	v, err = UnmarshalString(raw)
	checkErr()
	v, err = v.Get("int")
	checkErr()
	checkCond(v.Int32() == 123)

	topic = "parse uint32 in object"
	raw = `{"uint":123}`
	v, err = UnmarshalString(raw)
	checkErr()
	v, err = v.Get("uint")
	checkErr()
	checkCond(v.Uint32() == 123)

	topic = "parse float32 in object"
	raw = `{"float":123.456}`
	v, err = UnmarshalString(raw)
	checkErr()
	v, err = v.Get("float")
	checkErr()
	checkCond(v.Float32() == 123.456)

	topic = "parse float64 in object"
	raw = `{"float":123.456}`
	v, err = UnmarshalString(raw)
	checkErr()
	v, err = v.Get("float")
	checkErr()
	checkCond(v.Float64() == 123.456)

	topic = "parse negative in object"
	raw = `{"negative":-123}`
	v, err = UnmarshalString(raw)
	checkErr()
	v, err = v.Get("negative")
	checkErr()
	checkCond(v.IsNegative())

	topic = "parse positive in object"
	raw = `{"positive":123}`
	v, err = UnmarshalString(raw)
	checkErr()
	v, err = v.Get("positive")
	checkErr()
	checkCond(v.IsPositive())

	topic = "parse greater than int64 in object"
	raw = `{"int":9223372036854775808}`
	v, err = UnmarshalString(raw)
	checkErr()
	v, err = v.Get("int")
	checkErr()
	checkCond(v.GreaterThanInt64Max())
}

func TestRange(t *testing.T) {
	var checkCount int
	var topic string
	var err error
	var raw string
	var v *V
	var rangeCount int
	checkErr := func() {
		checkCount++
		if err != nil {
			t.Errorf("%02d - %s - error occurred: %v", checkCount, topic, err)
			return
		}
	}

	topic = "invalid array range"
	v = NewString("")
	v.RangeArray(func(i int, c *V) bool {
		return true
	}) // just do not panic

	topic = "nil array callback"
	v = NewArray()
	v.AppendNull().InTheEnd()
	v.RangeArray(nil) // just do not panic

	topic = "unmarshal array"
	raw = `[1,2,3,4,5,6,7,8,9,10,"11","12",13.0]`
	v, err = UnmarshalString(raw)
	checkErr()

	topic = "complete array range"
	rangeCount = 0
	v.RangeArray(func(i int, c *V) bool {
		rangeCount++
		element := 0
		if c.IsNumber() {
			element = c.Int()
		} else if c.IsString() {
			s := c.String()
			element, err = strconv.Atoi(s)
			checkErr()
		} else {
			t.Errorf("invalid jsonvalue type")
			return false
		}

		if element != i+1 {
			t.Errorf("unexpected element %d, index %d", element, i)
			return false
		}

		return true
	})

	if rangeCount != 13 {
		t.Errorf("unexpected rangeCount: %d", rangeCount)
		return
	}

	topic = "broken array range"
	rangeCount = 0
	brokenCount := 4
	v.RangeArray(func(i int, c *V) bool {
		if i < brokenCount {
			rangeCount++
			return true
		}
		return false
	})
	if rangeCount != brokenCount {
		t.Errorf("expected rangeCount %d but %d got", brokenCount, rangeCount)
		return
	}

	// object
	topic = "invalid object range"
	v = NewString("")
	v.RangeObjects(func(k string, c *V) bool {
		return true
	}) // just do not panic

	topic = "nil object callback"
	v = NewObject()
	v.SetString("world").At("hello")
	v.RangeObjects(nil) // just do not panic

	topic = "unmarshal object"
	raw = `{"number":12345,"bool":true,"string":"hello, world","null":null}`
	v, err = UnmarshalString(raw)
	checkErr()

	topic = "complete object range"
	checkKeys := map[string]bool{
		"number": true,
		"bool":   true,
		"string": true,
		"null":   true,
	}
	v.RangeObjects(func(k string, v *V) bool {
		delete(checkKeys, k)
		return true
	})
	if len(checkKeys) > 0 {
		t.Errorf("not all key checked, remains: %+v", checkKeys)
		return
	}

	topic = "broken object range"
	caughtValue := ""
	caughtKey := ""
	v.RangeObjects(func(k string, v *V) bool {
		caughtKey = k
		caughtValue = v.String()
		if k == "string" {
			return false
		}
		return true
	})
	if caughtKey == "string" && caughtValue == "hello, world" {
		// OK
	} else {
		t.Errorf("unexpected K-V: %s - %s", caughtKey, caughtValue)
		return
	}
}

func TestMustMarshalError(t *testing.T) {
	defer func() {
		if err := recover(); err == nil {
			t.Errorf("error expected but not caught")
		}
	}()

	v := &V{}
	v.MustMarshal()
	v.MustMarshalString()
	return
}

func TestMustMarshalStringError(t *testing.T) {
	defer func() {
		if err := recover(); err == nil {
			t.Errorf("error expected but not caught")
		}
	}()

	v := &V{}
	v.MustMarshalString()
	return
}

func TestValueError(t *testing.T) {
	var checkCount int
	var topic string
	var err error
	var raw string
	var v *V
	shouldError := func() {
		checkCount++
		if err == nil {
			s, _ := v.MarshalString()
			t.Errorf("%02d - %s - error expected but not caught, marshaled: %s", checkCount, topic, s)
			return
		}
		t.Logf("%02d - %s - expected error string: %v", checkCount, topic, err)
		return
	}

	topic = "invalid json"
	v = &V{}
	if v.String() != "" {
		t.Errorf("uninitizlized object should be empty")
		return
	}

	topic = "nil string input"
	raw = ""
	v, err = UnmarshalString(raw)
	shouldError()

	topic = "nil bytes input"
	v, err = Unmarshal(nil)
	shouldError()

	topic = "illegal char"
	raw = `\\`
	v, err = UnmarshalString(raw)
	shouldError()

	topic = "no start chars"
	raw = `     `
	v, err = UnmarshalString(raw)
	shouldError()

	topic = "illegal float number"
	raw = `1.a`
	v, err = UnmarshalString(raw)
	shouldError()

	topic = "illegal negative interger"
	raw = `-1a`
	v, err = UnmarshalString(raw)
	shouldError()

	topic = "illegal positive interger"
	raw = `1a`
	v, err = UnmarshalString(raw)
	shouldError()

	topic = "illegal true"
	raw = `trUE`
	v, err = UnmarshalString(raw)
	shouldError()

	topic = "illegal false"
	raw = `fAlse`
	v, err = UnmarshalString(raw)
	shouldError()

	topic = "illegal null"
	raw = `nUll`
	v, err = UnmarshalString(raw)
	shouldError()

	topic = "illegal string"
	raw = `"too many quote""`
	v, err = UnmarshalString(raw)
	shouldError()

	topic = "too short string"
	raw = `"`
	v, err = UnmarshalString(raw)
	shouldError()

	topic = "illegal escaping"
	raw = `"\"`
	v, err = UnmarshalString(raw)
	shouldError()

	// following error detections are not supported in jsonparser
	// should fixed later

	// topic = "illegal bool in object"
	// raw = `{"bool":tRue}`
	// v, err = UnmarshalString(raw)
	// shouldError()

	// topic = "illegal bool in array"
	// raw = `[tRue]`
	// v, err = UnmarshalString(raw)
	// shouldError()

	// topic = "illegal array"
	// raw = `["incompleteString]`
	// v, err = UnmarshalString(raw)
	// shouldError()

	topic = "marshaling uninitialized value"
	v = &V{}
	_, err = v.MarshalString()
	shouldError()
	_, err = v.Marshal()
	shouldError()

	topic = "marshaling with option"
	v = NewObject()
	v.SetNull().At("null")
	raw, _ = v.MarshalString()
	if raw != `{"null":null}` {
		t.Errorf("null object is omitted ('%s')", raw)
	}
	raw, _ = v.MarshalString(Opt{
		OmitNull: true,
	})
	if raw != "{}" {
		t.Errorf("null object is not omitted ('%s')", raw)
	}
	rawB, _ := v.Marshal(Opt{
		OmitNull: true,
	})
	if string(rawB) != "{}" {
		t.Errorf("null object is not omitted ('%s')", raw)
	}
}
