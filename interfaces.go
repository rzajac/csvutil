package csvutil

type Marshaler interface {
	MarshalCSV() ([]byte, error)
}

type Unmarshaler interface {
	UnmarshalCSV([]byte) error
}
