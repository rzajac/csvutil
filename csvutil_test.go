package csvutil

import (
	"github.com/rzajac/assert/assert"
	"io"
	"reflect"
	"strings"
	"testing"
)

var testCsvLines = []string{"Tony|23|123.456", "John|34|234.567|"}

type person struct {
	Name    string
	Age     int
	Balance float32
	Skipped string `csv:"-"`
}

type person2 struct {
	Name    string
	Balance float32
}

func StringReader(s string) *strings.Reader {
	return strings.NewReader(s)
}

func Test_NewReader(t *testing.T) {
	// Prepare test
	sr := StringReader(strings.Join(testCsvLines, "\n"))
	c := NewCsvUtil(sr).Comma('|').TrailingComma(true).FieldsPerRecord(-1)

	// Start test
	assert.NotNil(t, c.csvr)
	l, err := c.read()
	assert.NotError(t, err)
	assert.Equal(t, []string{"Tony", "23", "123.456"}, l)
	assert.Equal(t, "Tony|23|123.456", c.LastCsvLine())

	l, err = c.read()
	assert.NotError(t, err)
	assert.Equal(t, []string{"John", "34", "234.567", ""}, l)
	assert.Equal(t, "John|34|234.567|", c.LastCsvLine())
}

func Test_Header(t *testing.T) {
	// Prepare test
	c := NewCsvUtil(nil)

	// Start test
	exp := map[string]int{"Name": 0, "Age": 1, "Balance": 2}
	c.Header(exp)
	assert.Equal(t, exp, c.header)
}

func Test_getFields(t *testing.T) {
	// Prepare test
	p := &person{}

	// Start test
	fields, structName := getFields(p)
	assert.Equal(t, 3, len(fields))
	assert.Equal(t, "csvutil.person", structName)

	assert.Equal(t, "Name", fields[0].name)
	assert.Equal(t, reflect.String, fields[0].typ.Kind())

	assert.Equal(t, "Age", fields[1].name)
	assert.Equal(t, reflect.Int, fields[1].typ.Kind())

	assert.Equal(t, "Balance", fields[2].name)
	assert.Equal(t, reflect.Float32, fields[2].typ.Kind())
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
	assert.Equal(t, map[string]int{"Name": 0, "Age": 1, "Balance": 2}, headers)
}

func Test_SetData(t *testing.T) {
	// Prepare test
	sr := StringReader(strings.Join(testCsvLines, "\n"))
	c := NewCsvUtil(sr).Comma('|').TrailingComma(true).FieldsPerRecord(-1)

	// Start test
	p := &person{Skipped: "aaa"}
	err := c.SetData(p)
	assert.NotError(t, err)
	assert.Equal(t, "Tony", p.Name)
	assert.Equal(t, 23, p.Age)
	assert.Equal(t, float32(123.456), p.Balance)
	assert.Equal(t, "aaa", p.Skipped)

	err = c.SetData(p)
	assert.NotError(t, err)
	assert.Equal(t, "John", p.Name)
	assert.Equal(t, 34, p.Age)
	assert.Equal(t, float32(234.567), p.Balance)
	assert.Equal(t, "aaa", p.Skipped)

	err = c.SetData(p)
	assert.Equal(t, io.EOF, err)
	// The previous data stays intact
	assert.Equal(t, "John", p.Name)
	assert.Equal(t, 34, p.Age)
	assert.Equal(t, float32(234.567), p.Balance)
	assert.Equal(t, "aaa", p.Skipped)
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
	p := &person{"Tom", 45, 111.22, "aaa"}

	// Start test
	gotCsv := ToCsv(p, "|")
	assert.Equal(t, "Tom|45|111.22", gotCsv)
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
	sr := StringReader("Y|N")
	c := NewCsvUtil(sr).Comma('|').CustomBool([]string{"Y"}, []string{"N"})

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
	sr := StringReader("   Tom |12|123")
	c := NewCsvUtil(sr).Comma('|').Trim(" ")

	// Start test
	p := &person{}
	err := c.SetData(p)
	assert.NotError(t, err)

	assert.Equal(t, "Tom", p.Name)
	assert.Equal(t, 12, p.Age)
	assert.Equal(t, float32(123), p.Balance)
}
