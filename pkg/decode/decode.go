package decode

import (
	"encoding/json"
	"fmt"
	"reflect"

	"github.com/mitchellh/mapstructure"
	"github.com/spf13/viper"
)

func ToMap(src interface{}) (dst map[string]any, err error) {
	b, err := json.Marshal(src)
	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal(b, &dst); err != nil {
		return nil, err
	}
	return dst, nil
}

func WithTagName(tagName string) viper.DecoderConfigOption {
	return func(dc *mapstructure.DecoderConfig) {
		dc.TagName = tagName
	}
}

func WithDecodeHook(hook mapstructure.DecodeHookFunc) viper.DecoderConfigOption {
	return func(dc *mapstructure.DecoderConfig) {
		dc.DecodeHook = hook
	}
}

func UnmarshalJSONHookFunc(
	f reflect.Type,
	t reflect.Type,
	d interface{},
) (interface{}, error) {
	fmt.Println(f, t, d)
	r := reflect.New(t).Interface()
	u, ok := r.(interface{ UnmarshalJSON([]byte) error })
	if !ok {
		return d, nil
	}
	b, err := json.Marshal(d)
	if err != nil {
		return nil, err
	}
	if err := u.UnmarshalJSON(b); err != nil {
		return nil, err
	}
	return r, nil
}
