package nginxtype

import (
	"errors"
	"strconv"
	"strings"
)

// Uint32Bps :
type Uint32Bps uint32

// UnmarshalYAML :
func (i *Uint32Bps) UnmarshalYAML(unmarshal func(interface{}) error) error {
	size, err := unmarshalBpsYAML(unmarshal)
	if err != nil {
		return err
	}
	*i = Uint32Bps(size)
	return nil
}

// Val :
func (i *Uint32Bps) Val() uint32 {
	return uint32(*i)
}

// Uint64Bps :
type Uint64Bps uint64

// UnmarshalYAML :
func (i *Uint64Bps) UnmarshalYAML(unmarshal func(interface{}) error) error {
	size, err := unmarshalBpsYAML(unmarshal)
	if err != nil {
		return err
	}
	*i = Uint64Bps(size)
	return nil
}

// Val :
func (i *Uint64Bps) Val() uint64 {
	return uint64(*i)
}

// Int64Bps :
type Int64Bps int64

// UnmarshalYAML :
func (i *Int64Bps) UnmarshalYAML(unmarshal func(interface{}) error) error {
	size, err := unmarshalBpsYAML(unmarshal)
	if err != nil {
		return err
	}
	*i = Int64Bps(size)
	return nil
}

// Val :
func (i *Int64Bps) Val() int64 {
	return int64(*i)
}

// IntBps :
type IntBps int

// UnmarshalYAML :
func (i *IntBps) UnmarshalYAML(unmarshal func(interface{}) error) error {
	size, err := unmarshalBpsYAML(unmarshal)
	if err != nil {
		return err
	}
	*i = IntBps(size)
	return nil
}

func unmarshalBpsYAML(unmarshal func(interface{}) error) (int64, error) {
	var s string
	if err := unmarshal(&s); err != nil {
		return 0, err
	}
	return parseBps(s)
}

// Val :
func (i *IntBps) Val() int {
	return int(*i)
}

func parseBps(s string) (int64, error) {
	if s == "" {
		return 0, errors.New("parseBps: string is empty")
	}

	var scale int64 = 1
	s = strings.TrimSpace(s)

	unit := strings.ToUpper(s[len(s)-1:])
	var strNum string

	switch unit {
	case "K":
		strNum = s[:len(s)-1]
		scale = 1000

	case "M":
		strNum = s[:len(s)-1]
		scale = 1000 * 1000

	case "G":
		strNum = s[:len(s)-1]
		scale = 1000 * 1000 * 1000

	case "T":
		strNum = s[:len(s)-1]
		scale = 1000 * 1000 * 1000 * 1000

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
