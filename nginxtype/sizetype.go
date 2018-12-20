package nginxtype

import (
	"errors"
	"strconv"
	"strings"
)

// Uint32Size :
type Uint32Size uint32

// UnmarshalYAML :
func (i *Uint32Size) UnmarshalYAML(unmarshal func(interface{}) error) error {
	size, err := unmarshalSizeYAML(unmarshal)
	if err != nil {
		return err
	}
	*i = Uint32Size(size)
	return nil
}

// Val :
func (i *Uint32Size) Val() uint32 {
	return uint32(*i)
}

// Uint64Size :
type Uint64Size uint64

// UnmarshalYAML :
func (i *Uint64Size) UnmarshalYAML(unmarshal func(interface{}) error) error {
	size, err := unmarshalSizeYAML(unmarshal)
	if err != nil {
		return err
	}
	*i = Uint64Size(size)
	return nil
}

// Val :
func (i *Uint64Size) Val() uint64 {
	return uint64(*i)
}

// Int64Size :
type Int64Size int64

// UnmarshalYAML :
func (i *Int64Size) UnmarshalYAML(unmarshal func(interface{}) error) error {
	size, err := unmarshalSizeYAML(unmarshal)
	if err != nil {
		return err
	}
	*i = Int64Size(size)
	return nil
}

// Val :
func (i *Int64Size) Val() int64 {
	return int64(*i)
}

// IntSize :
type IntSize int

// UnmarshalYAML :
func (i *IntSize) UnmarshalYAML(unmarshal func(interface{}) error) error {
	size, err := unmarshalSizeYAML(unmarshal)
	if err != nil {
		return err
	}
	*i = IntSize(size)
	return nil
}

// Val :
func (i *IntSize) Val() int {
	return int(*i)
}

func unmarshalSizeYAML(unmarshal func(interface{}) error) (int64, error) {
	var s string
	if err := unmarshal(&s); err != nil {
		return 0, err
	}
	return parseSize(s)
}

func parseSize(s string) (int64, error) {
	if s == "" {
		return 0, errors.New("parseSize: string is empty")
	}

	var scale int64 = 1
	s = strings.TrimSpace(s)

	unit := strings.ToUpper(s[len(s)-1:])
	var strNum string

	switch unit {
	case "K":
		strNum = s[:len(s)-1]
		scale = 1024

	case "M":
		strNum = s[:len(s)-1]
		scale = 1024 * 1024

	case "G":
		strNum = s[:len(s)-1]
		scale = 1024 * 1024 * 1024

	case "T":
		strNum = s[:len(s)-1]
		scale = 1024 * 1024 * 1024 * 1024

	default:
		strNum = s
		scale = 1
	}

	size, err := strconv.ParseInt(strNum, 10, 64)
	if err != nil {
		return 0, err
	}

	return size * scale, nil
}
