// Deprecated, Remove Later
package utils

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/aanand/compose-file/loader"
	ctypes "github.com/aanand/compose-file/types"
)

func YamlServices(yaml []byte, env map[string]string) (*ctypes.Config, error) {
	dict, err := loader.ParseYAML(yaml)
	if err != nil {
		return nil, err
	}

	cds := ctypes.ConfigDetails{
		ConfigFiles: []ctypes.ConfigFile{
			{Config: dict},
		},
		Environment: env,
	}

	return loader.Load(cds)
}

// YamlVariables provide ability to parse all of shell variables like:
// $VARIABLE, ${VARIABLE}, ${VARIABLE:-default}, ${VARIABLE-default}
func YamlVariables(yaml []byte) []string {
	var (
		delimiter     = "\\$"
		substitution  = "[_a-z][_a-z0-9]*(?::?-[^}]+)?"
		patternString = fmt.Sprintf(
			"%s(?i:(?P<escaped>%s)|(?P<named>%s)|{(?P<braced>%s)}|(?P<invalid>))",
			delimiter, delimiter, substitution, substitution,
		)
		pattern = regexp.MustCompile(patternString)

		ret = make([]string, 0, 0)
	)

	pattern.ReplaceAllStringFunc(string(yaml), func(sub string) string {
		matches := pattern.FindStringSubmatch(sub)

		groups := make(map[string]string) // all matched naming parts
		for i, name := range pattern.SubexpNames() {
			if i != 0 {
				groups[name] = matches[i]
			}
		}

		text := groups["named"]
		if text == "" {
			text = groups["braced"]
		}
		if text == "" {
			text = groups["escaped"]
		}

		var (
			sep    string
			fields []string
		)
		switch {
		case text == "":
			goto END
		case strings.Contains(text, ":-"):
			sep = ":-"
		case strings.Contains(text, "-"):
			sep = "-"
		default:
			ret = append(ret, text+":")
			goto END
		}

		fields = strings.SplitN(text, sep, 2)
		ret = append(ret, fields[0]+":"+fields[1])

	END:
		return ""
	})

	return ret
}
