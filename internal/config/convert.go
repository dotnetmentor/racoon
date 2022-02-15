package config

import (
	"fmt"

	"github.com/dotnetmentor/racoon/internal/output"
	"gopkg.in/yaml.v2"
)

func UnmarshalConfig(t OutputType, config map[string]interface{}) (output.Output, error) {
	b, err := yaml.Marshal(config)
	if err != nil {
		return nil, err
	}

	switch t {
	case OutputTypeDotenv:
		out := output.NewDotenv()
		if err := yaml.Unmarshal(b, &out); err != nil {
			return nil, err
		}
		return out, nil
	case OutputTypeTfvars:
		out := output.NewTfvars()
		if err := yaml.Unmarshal(b, &out); err != nil {
			return nil, err
		}
		return out, nil
	case OutputTypeJson:
		out := output.NewJson()
		if err := yaml.Unmarshal(b, &out); err != nil {
			return nil, err
		}
		return out, nil
	default:
		panic(fmt.Errorf("unsupported output type %s", t))
	}
}

func AsOutput(o OutputConfig) output.Output {
	switch o.Type {
	case OutputTypeDotenv:
		return o.output.(output.Dotenv)
	case OutputTypeTfvars:
		return o.output.(output.Tfvars)
	case OutputTypeJson:
		return o.output.(output.Json)
	default:
		panic(fmt.Errorf("unsupported output type %s", o.Type))
	}
}
