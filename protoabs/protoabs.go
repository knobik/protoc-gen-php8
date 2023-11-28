package protoabs

import (
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
	"strings"
)

var (
	PHPIncludeMap = map[ClassType]string{
		CTypeMessage: "Google\\Protobuf\\Internal\\Message",
	}

	ClassTypeTemplateMap = map[ClassType]string{
		CTypeMessage: "message.tmpl",
		CTypeEnum:    "enum.tmpl",
	}

	ObjectRefClassMap = map[string]*Class{}
)

func PackageToNamespace(pack string) string {
	parts := strings.Split(pack, ".")
	for i, p := range parts {
		parts[i] = cases.Title(language.Und, cases.NoLower).String(p)
	}
	return strings.Join(parts, "\\")
}

func FQNBasename(fqn string) string {
	parts := strings.Split(fqn, "\\")
	return parts[len(parts)-1]
}

func phpDefault(p *Property) string {
	t := p.Type
	if p.IsEnum() {
		t = "int"
	}

	switch t {
	case "string":
		return "''"
	case "int":
		return "0"
	case "float":
		return "0.0"
	case "bool":
		return "false"
	default:
		return "null"
	}
}
