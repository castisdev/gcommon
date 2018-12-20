package nginxtype

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
	yaml "gopkg.in/yaml.v2"
)

func Test_parseSize(t *testing.T) {
	type args struct {
		s string
	}
	tests := []struct {
		name    string
		args    args
		want    int64
		wantErr bool
	}{
		{
			name:    "10",
			args:    args{"10"},
			want:    10,
			wantErr: false,
		},
		{
			name:    "10K - upper case",
			args:    args{"10K"},
			want:    10 * 1024,
			wantErr: false,
		},
		{
			name:    "10k - lower case",
			args:    args{"10k"},
			want:    10 * 1024,
			wantErr: false,
		},
		{
			name:    "10M",
			args:    args{"10M"},
			want:    10 * 1024 * 1024,
			wantErr: false,
		},
		{
			name:    "10G",
			args:    args{"10G"},
			want:    10 * 1024 * 1024 * 1024,
			wantErr: false,
		},
		{
			name:    "10T",
			args:    args{"10T"},
			want:    10 * 1024 * 1024 * 1024 * 1024,
			wantErr: false,
		},
		{
			name:    "empty",
			args:    args{""},
			want:    0,
			wantErr: true,
		},
		{
			name:    "invalid suffix",
			args:    args{"10A"},
			want:    0,
			wantErr: true,
		},
		{
			name:    "long suffix",
			args:    args{"10KB"},
			want:    0,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseSize(tt.args.s)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseSize() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("parseSize() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_UnmarshalYAML_Size(t *testing.T) {
	type args struct {
		unmarshal func(interface{}) error
	}
	tests := []struct {
		name    string
		data    string
		value   interface{}
		wantErr bool
	}{
		{
			name:    "Uint32Size",
			data:    "v: 5M",
			value:   map[string]Uint32Size{"v": Uint32Size(5 * 1024 * 1024)},
			wantErr: false,
		},
		{
			name:    "Uint64Size",
			data:    "v: 5M",
			value:   map[string]Uint64Size{"v": Uint64Size(5 * 1024 * 1024)},
			wantErr: false,
		},
		{
			name:    "Int64Size",
			data:    "v: 5M",
			value:   map[string]Int64Size{"v": Int64Size(5 * 1024 * 1024)},
			wantErr: false,
		},
		{
			name:    "IntSize",
			data:    "v: 5M",
			value:   map[string]IntSize{"v": IntSize(5 * 1024 * 1024)},
			wantErr: false,
		},
	}

	for _, item := range tests {
		tp := reflect.ValueOf(item.value).Type()
		var value interface{}
		switch tp.Kind() {
		case reflect.Map:
			value = reflect.MakeMap(tp).Interface()
		case reflect.String:
			value = reflect.New(tp).Interface()
		case reflect.Ptr:
			value = reflect.New(tp.Elem()).Interface()
		default:
			t.Fatalf("missing case for %s", tp)
		}
		err := yaml.Unmarshal([]byte(item.data), value)
		if _, ok := err.(*yaml.TypeError); !ok && (err != nil) != item.wantErr {
			assert.NotNil(t, err, "UnmarshalYAML, error = %v, wantError %v", err, item.wantErr)
		}
		if tp.Kind() == reflect.String {
			assert.Equal(t, *value.(*string), item.value, "UnmarshalYAML, error = %v, wantError %v", err, item.wantErr)
		} else {
			assert.Equal(t, value, item.value, "UnmarshalYAML, error = %v, wantError %v", err, item.wantErr)
		}
	}
}
