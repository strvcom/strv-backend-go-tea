package decode

import (
	"github.com/mitchellh/mapstructure"
	"github.com/spf13/viper"
)

func WithTagName(tagName string) viper.DecoderConfigOption {
	return func(dc *mapstructure.DecoderConfig) {
		dc.TagName = tagName
	}
}
