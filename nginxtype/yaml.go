package nginxtype

import (
	"encoding/hex"
	"fmt"
	"net/url"
	"regexp"
)

////////////////////////////////////////////////////////////////////////////////

// Regexp :
type Regexp struct {
	*regexp.Regexp
}

// UnmarshalYAML :
func (r *Regexp) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var s string
	if err := unmarshal(&s); err != nil {
		return err
	}
	if s == "" {
		r = nil
		return nil
	}

	var err error
	if r.Regexp, err = regexp.Compile(s); err != nil {
		r = nil
		return fmt.Errorf("%s is invalid regular expression, %v", s, err)
	}

	return nil
}

// MarshalYAML :
func (r *Regexp) MarshalYAML() (interface{}, error) {
	if r == nil || r.Regexp == nil {
		return nil, nil
	}
	return r.String(), nil
}

////////////////////////////////////////////////////////////////////////////////

// URL :
type URL struct {
	*url.URL
}

// UnmarshalYAML :
func (u *URL) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var s string
	if err := unmarshal(&s); err != nil {
		return err
	}
	if s == "" {
		u = nil
		return nil
	}

	var err error
	if u.URL, err = url.Parse(s); err != nil {
		u = nil
		return fmt.Errorf("%s is invalid url, %v", s, err)
	}

	return nil
}

// MarshalYAML :
func (u *URL) MarshalYAML() (interface{}, error) {
	if u == nil || u.URL == nil {
		return nil, nil
	}
	return u.String(), nil
}

////////////////////////////////////////////////////////////////////////////////

// HexString :
type HexString []byte

// UnmarshalYAML :
func (hx *HexString) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var s string
	if err := unmarshal(&s); err != nil {
		return err
	}
	if s == "" {
		hx = nil
		return nil
	}

	var err error
	if *hx, err = hex.DecodeString(s); err != nil {
		hx = nil
		return fmt.Errorf("%s is invalid hex string, %v", s, err)
	}

	return nil
}

// MarshalYAML :
func (hx *HexString) MarshalYAML() (interface{}, error) {
	if hx == nil {
		return nil, nil
	}
	return hex.EncodeToString(*hx), nil
}
