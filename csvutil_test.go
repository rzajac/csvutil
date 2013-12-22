package csvutil

import (
	"github.com/rzajac/goassert/assert"
	"io"
	"reflect"
	"strings"
	"testing"
)

// Stuff to help testing

var testCsvLines = []string{"Tony|23|123.456|Y", "John|34|234.567|N|"}

type person struct {
	Name       string
	Age        int
	Balance    float32
	Skipped    string `csv:"-"`
	LowBalance bool
}

type person2 struct {
	Name    string
	Balance float32
}

type A struct {
	Field1, Field2 string
}
type B struct {
	A
	Field3 string
}

type StringReaderCloser struct {
	sr *strings.Reader
}

func (s *StringReaderCloser) Read(p []byte) (n int, err error) {
	return s.sr.Read(p)
}

func (s *StringReaderCloser) Close() error {
	return nil
}

func StringReader(s string) *StringReaderCloser {
	src := &StringReaderCloser{sr: strings.NewReader(s)}
	return src
}

// Actual tests

func Test_NewReader(t *testing.T) {
	// Prepare test
	sr := StringReader(strings.Join(testCsvLines, "\n"))
	c := NewCsvUtil(sr).Comma('|').TrailingComma(true).FieldsPerRecord(-1)

	// Start test
	assert.NotNil(t, c.csvr)
	l, err := c.read()
	assert.NotError(t, err)
	assert.Equal(t, []string{"Tony", "23", "123.456", "Y"}, l)
	assert.Equal(t, "Tony|23|123.456|Y", c.LastCsvLine())

	l, err = c.read()
	assert.NotError(t, err)
	assert.Equal(t, []string{"John", "34", "234.567", "N", ""}, l)
	assert.Equal(t, "John|34|234.567|N|", c.LastCsvLine())
}

func Test_Header(t *testing.T) {
	// Prepare test
	c := NewCsvUtil(nil)

	// Start test
	exp := CsvHeader{"Name": 0, "Age": 1, "Balance": 2, "LowBalance": 3}
	c.Header(exp)
	assert.Equal(t, exp, c.header)
}

func Test_getFields(t *testing.T) {
	// Prepare test
	p := &person{}

	// Start test
	fields, structName := getFields(p)
	assert.Equal(t, 4, len(fields))
	assert.Equal(t, "csvutil.person", structName)

	assert.Equal(t, "Name", fields[0].name)
	assert.Equal(t, reflect.String, fields[0].typ.Kind())

	assert.Equal(t, "Age", fields[1].name)
	assert.Equal(t, reflect.Int, fields[1].typ.Kind())

	assert.Equal(t, "Balance", fields[2].name)
	assert.Equal(t, reflect.Float32, fields[2].typ.Kind())

	assert.Equal(t, "LowBalance", fields[3].name)
	assert.Equal(t, reflect.Bool, fields[3].typ.Kind())
}

func Test_getFields_panic(t *testing.T) {
	fn := func() {
		p := make(map[int]string)
		getFields(p)
	}

	assert.Panic(t, fn, "Expected pointer")

	fn = func() {
		p := 1
		getFields(p)
	}

	assert.Panic(t, fn, "Expected pointer")

	fn = func() {
		p := []string{"aaa"}
		getFields(p)
	}

	assert.Panic(t, fn, "Expected pointer")

	fn = func() {
		p := []string{"aaa"}
		getFields(&p)
	}

	assert.Panic(t, fn, "Expected pointer to")

	fn = func() {
		p := new(person)
		getFields(&p)
	}

	assert.Panic(t, fn, "Expected pointer to")
}

func Test_getHeaders(t *testing.T) {
	// Prepare test
	p := &person{}
	fields, structName := getFields(p)
	assert.Equal(t, "csvutil.person", structName)

	// Start test
	headers := getHeaders(fields)
	assert.Equal(t, CsvHeader{"Name": 0, "Age": 1, "Balance": 2, "LowBalance": 3}, headers)
}

func Test_SetData(t *testing.T) {
	// Prepare test
	sr := StringReader(strings.Join(testCsvLines, "\n"))
	c := NewCsvUtil(sr).Comma('|').
		TrailingComma(true).
		FieldsPerRecord(-1).
		CustomBool([]string{"Y"}, []string{"N"})

	// Start test
	p := &person{Skipped: "aaa"}
	err := c.SetData(p)
	assert.NotError(t, err)
	assert.Equal(t, "Tony", p.Name)
	assert.Equal(t, 23, p.Age)
	assert.Equal(t, float32(123.456), p.Balance)
	assert.Equal(t, "aaa", p.Skipped)
	assert.Equal(t, true, p.LowBalance)

	err = c.SetData(p)
	assert.NotError(t, err)
	assert.Equal(t, "John", p.Name)
	assert.Equal(t, 34, p.Age)
	assert.Equal(t, float32(234.567), p.Balance)
	assert.Equal(t, "aaa", p.Skipped)
	assert.Equal(t, false, p.LowBalance)

	err = c.SetData(p)
	assert.Equal(t, io.EOF, err)
	// The previous data stays intact
	assert.Equal(t, "John", p.Name)
	assert.Equal(t, 34, p.Age)
	assert.Equal(t, float32(234.567), p.Balance)
	assert.Equal(t, "aaa", p.Skipped)
	assert.Equal(t, false, p.LowBalance)
}

func Test_Comma(t *testing.T) {
	csvu := NewCsvUtil(nil)
	assert.Equal(t, ',', csvu.csvr.Comma)
	csvu.Comma('|')
	assert.Equal(t, '|', csvu.csvr.Comma)
}

func Test_TrailingComma(t *testing.T) {
	csvu := NewCsvUtil(nil)
	assert.Equal(t, false, csvu.csvr.TrailingComma)
	csvu.TrailingComma(true)
	assert.Equal(t, true, csvu.csvr.TrailingComma)
}

func Test_Comment(t *testing.T) {
	csvu := NewCsvUtil(nil)
	assert.Equal(t, '\000', csvu.csvr.Comment)
	csvu.Comment('#')
	assert.Equal(t, '#', csvu.csvr.Comment)
}

func Test_FieldsPerRecord(t *testing.T) {
	csvu := NewCsvUtil(nil)
	assert.Equal(t, 0, csvu.csvr.FieldsPerRecord)
	csvu.FieldsPerRecord(-1)
	assert.Equal(t, -1, csvu.csvr.FieldsPerRecord)
}

func Test_ToCsv(t *testing.T) {
	// Prepare test
	p := &person{"Tom", 45, 111.22, "aaa", true}

	// Start test
	gotCsv := ToCsv(p, "|", "YY", "NN")
	assert.Equal(t, "Tom|45|111.22|YY", gotCsv)
}

func Test_pickingColumns(t *testing.T) {
	// Prepare test
	sr := StringReader(strings.Join(testCsvLines, "\n"))
	c := NewCsvUtil(sr).Comma('|').TrailingComma(true).FieldsPerRecord(-1)

	c.Header(map[string]int{"Name": 0, "Balance": 2})

	// Start test
	p := &person2{}
	err := c.SetData(p)
	assert.NotError(t, err)

	assert.Equal(t, "Tony", p.Name)
	assert.Equal(t, float32(123.456), p.Balance)
}

func Test_customTrueFalse(t *testing.T) {
	// Prepare test
	sr := StringReader("YY|NN")
	c := NewCsvUtil(sr).Comma('|').CustomBool([]string{"YY"}, []string{"NN"})

	type YN struct {
		Yes bool
		No  bool
	}

	// Start test
	p := &YN{}
	err := c.SetData(p)
	assert.NotError(t, err)

	assert.Equal(t, true, p.Yes)
	assert.Equal(t, false, p.No)
}

func Test_trim(t *testing.T) {
	// Prepare test
	sr := StringReader("   Tom |12|123|T")
	c := NewCsvUtil(sr).Comma('|').Trim(" ")

	// Start test
	p := &person{}
	err := c.SetData(p)
	assert.NotError(t, err)

	assert.Equal(t, "Tom", p.Name)
	assert.Equal(t, 12, p.Age)
	assert.Equal(t, float32(123), p.Balance)
	assert.Equal(t, true, p.LowBalance)
}

func Test_embededToCsv(t *testing.T) {
	// Prepare test
	b := new(B)
	b.Field1 = "F1"
	b.Field2 = "F2"
	b.Field3 = "F3"

	// Start test
	assert.Equal(t, "F1,F2,F3", ToCsv(b, ",", "Y", "N"))
}

// func Test_setEmbeded(t *testing.T) {
// 	// Prepare test
// 	type A struct {
// 		Field1, Field2 string
// 	}
// 	type B struct {
// 		A
// 		Field3 string
// 	}
// 	b := new(B)

// 	sr := StringReader("F1,F2,F3")
// 	c := NewCsvUtil(sr)

// 	// Start test
// 	c.SetData(b)

// 	fmt.Println(b)
// 	t.Fail()
// }
