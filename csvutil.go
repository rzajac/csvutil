// CSV utilities for Go tests
//
// Csvutil (c) Rafal Zajac <rzajac@gmail.com>
// http://github.com/rzajac/csvutil
//
// Licensed under the MIT license

// Package provides tools to set struct values based on CSV line / file
package csvutil

import (
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"reflect"
	"strconv"
	"strings"
)

// Structure fields cache.
var fCache map[string][]*sField

// CSV headers cache.
var hCache map[string]map[string]int

// Provides primitives to read CSV file and set values on structures.
type Reader struct {
	csvr         *csv.Reader         // CSV reader
	header       map[string]int      // The names of the CSV columns
	csvLine      []string            // The CSV column values
	customHeader bool                // True if custom CSV header was set
	customTBool  map[string]struct{} // Custom true values
	customFBool  map[string]struct{} // Custom false values
	trim         string              // Characters to trim
}

// NewCsvParser returns new Reader
func NewCsvUtil(r io.Reader) *Reader {
	reader := &Reader{
		csvr: csv.NewReader(r)}
	reader.customTBool = make(map[string]struct{})
	reader.customFBool = make(map[string]struct{})
	return reader
}

// Comma sets field delimiter (set to ',' by default).
func (r *Reader) Comma(s rune) *Reader {
	r.csvr.Comma = s
	return r
}

// TrailingComma allow trailing comma (set to false by default).
func (r *Reader) TrailingComma(b bool) *Reader {
	r.csvr.TrailingComma = b
	return r
}

// Comment sets number of expected fields per record.
func (r *Reader) Comment(c rune) *Reader {
	r.csvr.Comment = c
	return r
}

// FieldsPerRecord sets number of fields.
func (r *Reader) FieldsPerRecord(i int) *Reader {
	r.csvr.FieldsPerRecord = i
	return r
}

// LazyQuotes allow lazy quotes.
func (r *Reader) LazyQuotes(b bool) *Reader {
	r.csvr.LazyQuotes = b
	return r
}

// CustomBool set custom boolean values
func (r *Reader) CustomBool(t []string, f []string) *Reader {
	for _, tv := range t {
		r.customTBool[tv] = struct{}{}
	}
	for _, fv := range f {
		r.customFBool[fv] = struct{}{}
	}
	return r
}

// Trim list of characters to trim before returning CSV column value.
func (r *Reader) Trim(t string) *Reader {
	r.trim = t
	return r
}

// boolTr translates custom true / false values to string that strconv.ParseBool() understands.
func (r *Reader) boolTr(value string) string {
	if _, ok := r.customTBool[value]; ok {
		return "T" // One of the supported true string values
	}
	if _, ok := r.customFBool[value]; ok {
		return "F" // One of the supported true string values
	}
	return value
}

// read reads one record from CSV file.
func (r *Reader) read() ([]string, error) {
	var err error
	r.csvLine, err = r.csvr.Read()
	return r.csvLine, err
}

// Header sets header for CSV file.
func (r *Reader) Header(h map[string]int) *Reader {
	r.header = h
	r.customHeader = true
	return r
}

// SetData sets values from CSV record on passed struct.
// Returns error or io.EOF when no more records exist.
func (r *Reader) SetData(v interface{}) error {
	var err error
	var ok bool

	_, err = r.read()
	if err != nil {
		return err
	}

	// Initialize cache if its not there yet
	if hCache == nil {
		hCache = make(map[string]map[string]int)
	}

	structFields, structName := getFields(v)

	if !r.customHeader {
		if r.header, ok = hCache[structName]; !ok {
			r.header = getHeaders(structFields)
			hCache[structName] = r.header
		}
	}

	value := reflect.ValueOf(v).Elem()
	for _, sf := range structFields {
		strValue := r.colByName(sf.name)
		err = r.setValue(value, sf, strValue)
		if err != nil {
			return err
		}
	}

	return err
}

// LastCsvLine returns most recent CSV line that has been read from the file.
func (r *Reader) LastCsvLine() string {
	return strings.Join(r.csvLine, string(r.csvr.Comma))
}

// colByName returns CSV column value by name.
func (r *Reader) colByName(colName string) string {
	value := r.csvLine[r.header[colName]]
	if r.trim != "" {
		value = strings.Trim(value, r.trim)
	}
	return value
}

// ToCsv takes a struct and returns CSV line.
func ToCsv(v interface{}, delim, boolTrue, boolFalse string) string {
	var csvLine []string
	var strValue string
	t := reflect.ValueOf(v)

	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	if t.Kind() != reflect.Struct {
		panic("Expected pointer to a struct")
	}

	for i := 0; i < t.NumField(); i++ {
		structField := t.Type().Field(i)
		field := t.Field(i)

		if structField.Anonymous {
			strValue = ToCsv(field.Interface(), delim, boolTrue, boolFalse)
			csvLine = append(csvLine, strValue)
			continue
		}

		if !skip(structField.Tag) && field.CanInterface() {
			strValue = getValue(field, boolTrue, boolFalse)
			csvLine = append(csvLine, strValue)
		}
	}

	return strings.Join(csvLine, delim)
}

// sField described structure field.
type sField struct {
	name string
	typ  reflect.Type
}

// getFields returns array of sField for the passed struct.
func getFields(v interface{}) ([]*sField, string) {
	structFields := []*sField{}

	t := reflect.TypeOf(v)

	if t.Kind() != reflect.Ptr {
		panic("Expected pointer")
	}
	t = t.Elem()
	if t.Kind() != reflect.Struct {
		panic("Expected pointer to a struct")
	}

	// Initialize cache if its not there yet
	if fCache == nil {
		fCache = make(map[string][]*sField)
	}

	var ok bool
	structName := t.String()

	if structFields, ok = fCache[structName]; ok {
		return structFields, structName
	}

	for i := 0; i < t.NumField(); i++ {
		structField := t.Field(i)
		if !skip(structField.Tag) && reflect.ValueOf(v).Elem().Field(i).CanSet() {
			f := &sField{}
			f.name = structField.Name
			f.typ = structField.Type
			structFields = append(structFields, f)
		}
	}

	fCache[structName] = structFields

	return structFields, structName
}

// skip returns true if struct field is tagged with skip.
func skip(tag reflect.StructTag) bool {
	return strings.HasPrefix(tag.Get("csv"), "-")
}

// getHeaders returns array of CSV column names in order they appear in the record.
func getHeaders(fields []*sField) map[string]int {
	header := make(map[string]int)
	for idx, field := range fields {
		header[field.name] = idx
	}
	return header
}

// setValue sets structure value from CSV column.
func (r *Reader) setValue(v reflect.Value, f *sField, value string) (err error) {
	elem := v.FieldByName(f.name)
	if elem.CanSet() {
		switch f.typ.Kind() {
		case reflect.String:
			elem.SetString(value)
			return
		case reflect.Int:
			fallthrough
		case reflect.Int8:
			fallthrough
		case reflect.Int16:
			fallthrough
		case reflect.Int32:
			fallthrough
		case reflect.Int64:
			var i64 int64
			if value == "" {
				elem.SetInt(0)
			} else {
				i64, err = strconv.ParseInt(value, 10, 64)
				elem.SetInt(i64)
			}
			return
		case reflect.Uint:
			fallthrough
		case reflect.Uint8:
			fallthrough
		case reflect.Uint16:
			fallthrough
		case reflect.Uint32:
			fallthrough
		case reflect.Uint64:
			var u64 uint64
			if value == "" {
				elem.SetUint(0)
			} else {
				u64, err = strconv.ParseUint(value, 10, 64)
				elem.SetUint(u64)
			}
			return
		case reflect.Float32:
			fallthrough
		case reflect.Float64:
			var f64 float64
			if value == "" {
				elem.SetFloat(f64)
			} else {
				f64, err = strconv.ParseFloat(value, 64)
				elem.SetFloat(f64)
			}
			return
		case reflect.Bool:
			var b bool
			b, err = strconv.ParseBool(r.boolTr(value))
			elem.SetBool(b)
		default:
			return errors.New(fmt.Sprintf("Unsupported structure field set %s -> %v.", f.name, value))
		}
	} else {
		return errors.New("Wasn't able to set value on filed: " + f.name + " <- " + value)
	}

	return
}

// getValue gets string representation of the struct field.
func getValue(field reflect.Value, boolTrue, boolFalse string) string {
	switch field.Kind() {
	case reflect.Int:
		return strconv.Itoa(field.Interface().(int))
	case reflect.Int8:
		return strconv.FormatInt(int64(field.Interface().(int8)), 10)
	case reflect.Int16:
		return strconv.FormatInt(int64(field.Interface().(int16)), 10)
	case reflect.Int32:
		return strconv.FormatInt(int64(field.Interface().(int32)), 10)
	case reflect.Int64:
		return strconv.FormatInt(field.Interface().(int64), 10)
	case reflect.Uint:
		return strconv.FormatUint(uint64(field.Interface().(uint)), 10)
	case reflect.Uint8:
		return strconv.FormatUint(uint64(field.Interface().(uint8)), 10)
	case reflect.Uint16:
		return strconv.FormatUint(uint64(field.Interface().(uint16)), 10)
	case reflect.Uint32:
		return strconv.FormatUint(uint64(field.Interface().(uint32)), 10)
	case reflect.Uint64:
		return strconv.FormatUint(field.Interface().(uint64), 10)
	case reflect.Float32:
		return strconv.FormatFloat(float64(field.Interface().(float32)), 'f', -1, 32)
	case reflect.Float64:
		return strconv.FormatFloat(field.Interface().(float64), 'f', -1, 64)
	case reflect.String:
		return field.Interface().(string)
	case reflect.Bool:
		if field.Interface().(bool) {
			return boolTrue
		} else {
			return boolFalse
		}
	default:
		panic("Wasn't able to get value for filed: " + field.Type().Name() + " field type:" + field.Type().String())
	}
}
