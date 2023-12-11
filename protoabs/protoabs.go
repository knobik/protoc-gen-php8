package protoabs

import (
	"github.com/sanity-io/litter"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
	"log"
	"strings"
)

var (
	ClassTypeTemplateMap = map[ClassType]string{
		CTypeMessage: "message.tmpl",
		CTypeEnum:    "enum.tmpl",
		CTypeService: "service.tmpl",
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

	if p.IsMap() {
		t = "array"
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
	case "array":
		return "[]"
	default:
		return "null"
	}
}

func dump(value ...interface{}) {
	log.Println(litter.Sdump(value...))
}
