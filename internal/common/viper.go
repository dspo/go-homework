package common

import (
	"io"
	"os"
	"reflect"

	"github.com/go-viper/mapstructure/v2"
	"github.com/spf13/viper"
	"github.com/valyala/fasttemplate"
)

func init() {
	viper.SetDefault("listen.host", "0.0.0.0")
	viper.SetDefault("listen.port", 8080)
}

// ViperDecodeHook sets up custom decode hooks for Viper, including time duration, slice, and environment variable rendering.
func ViperDecodeHook(config *mapstructure.DecoderConfig) {
	config.DecodeHook = mapstructure.ComposeDecodeHookFunc(
		mapstructure.StringToTimeDurationHookFunc(),
		mapstructure.StringToSliceHookFunc(","),
		ViperDecodeHookFuncRenderEnvs,
	)
}

// ViperDecodeHookFuncRenderEnvs renders environment variables in the given string template.
func ViperDecodeHookFuncRenderEnvs(f, _ reflect.Kind, data any) (any, error) {
	if f != reflect.String {
		return data, nil
	}
	t, err := fasttemplate.NewTemplate(data.(string), "${", "}")
	if err != nil {
		return nil, err // err occurs only when there is a startTag but not an endTag.
	}
	s := t.ExecuteFuncString(func(w io.Writer, tag string) (int, error) {
		return io.WriteString(w, os.Getenv(tag)) // Inserts the value of the environment variable whether it has a value or not.
	})
	return s, nil
}
