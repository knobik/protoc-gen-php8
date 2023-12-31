package protoabs

import (
	"github.com/sanity-io/litter"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
	"log"
	"slices"
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

	if p.IsRepeated() {
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

func PrefixReserved(value string, isConstant bool) string {
	if IsReservedKeyword(value, isConstant) {
		value = Opts.ReservedPrefix + value
	}

	return value
}

func IsReservedKeyword(value string, isConstant bool) bool {
	kw := []string{
		"abstract", "and", "array", "as", "break", "callable", "case", "catch", "class", "clone", "const", "continue",
		"declare", "default", "die", "do", "echo", "else", "elseif", "empty", "enddeclare", "endfor", "endforeach",
		"endif", "endswitch", "endwhile", "eval", "exit", "extends", "final", "finally", "fn", "for", "foreach",
		"function", "global", "goto", "if", "implements", "include", "include_once", "instanceof", "insteadof",
		"interface", "isset", "list", "match", "namespace", "new", "or", "parent", "print", "private", "protected",
		"public", "readonly", "require", "require_once", "return", "self", "static", "switch", "throw", "trait",
		"try", "unset", "use", "var", "while", "xor", "yield", "int", "float", "bool", "string", "true", "false",
		"null", "void", "iterable",
	}
	vpn := []string{
		"int", "float", "bool", "string", "true", "false", "null", "void", "iterable", "parent", "self", "readonly",
	}

	if slices.Contains(kw, strings.ToLower(value)) {
		if isConstant && slices.Contains(vpn, strings.ToLower(value)) {
			return false
		}

		return true
	}

	return false
}

func dump(value ...interface{}) {
	log.Println(litter.Sdump(value...))
}
