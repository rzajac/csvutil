## CSVUtil

Provides tools to help set Go structures from CSV lines / files and vice versa.

## Installation

To install assertions run:

    $ go get github.com/rzajac/csvutil

## Documentation

* http://godoc.org/github.com/rzajac/csvutil

### Set fields from CSV

Here we are setting _person_ fields from CSV. The columns in CSV line must be in the same order as fields in the structure (from top to bottom). You may skip fields by tagging them with `csv:"-"`.

```go
var testCsvLines = []string{"Tony|23|123.456", "John|34|234.567|"}

type person struct {
	Name    string
	Age     int
	Skipped string `csv:"-"` // Skip this field when setting the structure
	Balance float32
}

main () {
	// This can be any Reader() interface
	sr := StringReader(strings.Join(testCsvLines, "\n"))

	// Set delimiter to '|', allow for trailing comma and do not check fields per CSV record
	c := NewCsvUtil(sr).Comma('|').TrailingComma(true).FieldsPerRecord(-1)

	// Struct we wil lpopulate with data
	p := &person{Skipped: "aaa"}

	// Set values from CSV line to person structure
	_, err := c.SetData(p)

	// Do work with p

	// Set values from the second CSV line
	_, err := c.SetData(p)

}
```

### Picking only CSV columns we are interested in

```go
type person2 struct {
	Name    string
	Balance float32
}

main () {
	sr := StringReader(strings.Join(testCsvLines, "\n"))

	// Set delimiter to '|', allow for trailing comma and do not check fields per CSV record
	c := NewCsvUtil(sr).Comma('|').TrailingComma(true).FieldsPerRecord(-1)

	// Set header with column names matching structure fields and column indexes on the CSV line.
	// The indexes in the CSV line start with 0.
	c.Header(map[string]int{"Name": 0, "Balance": 2})

	// Struct we wil lpopulate with data
	p := &person2{}

	// Set values from CSV line to person2 structure
	_, err := c.SetData(p)

	// Do work with p

	// Set values from the second CSV line
	_, err := c.SetData(p)
}

```

### Custom true / false values

**CustomBool()** method allows you to set custom true / false values in CSV column.

```go
main () {
	sr := StringReader("Y|N")
	c := NewCsvUtil(sr).CustomBool([]string{"Y"}, []string{"N"})
```

### Create CSV line from struct

```go
p := &person{"Tom", 45, "aaa", 111.22}

csvLine, err := ToCsv(p, "|")
fmt.Println(csvLine) // Prints: Tom|45|111.22
```

## TODO

* Add writing CSV to file

## License

Released under the MIT License.
Assert (c) Rafal Zajac <rzajac@gmail.com>