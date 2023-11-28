package protoabs

import (
	"github.com/iancoleman/strcase"
	"strings"
)

type Property struct {
	Number    int
	Name      string
	Type      string
	ProtoType string
	Repeated  bool
	ObjectRef string
}

func (p *Property) IsMap() bool {
	if dep := p.Dependency(); dep != nil {
		return dep.IsMapEntry()
	}

	return false
}

func (p *Property) IsEnum() bool {
	if dep := p.Dependency(); dep != nil {
		return dep.IsEnum()
	}

	return false
}

func (p *Property) Dependency() *Class {
	if p.ObjectRef == "" {
		return nil
	}

	return ObjectRefClassMap[p.ObjectRef]
}

func (p *Property) PropertyName() string {
	return strcase.ToLowerCamel(p.Name)
}

func (p *Property) PropertyType() string {
	if p.IsMap() {
		return "array|MapField"
	}

	if p.Repeated {
		return "array|RepeatedField"
	}

	return p.Type
}

func (p *Property) PropertyDefault() string {
	return phpDefault(p)
}

func (p *Property) CommentPropertyType() string {
	if p.IsMap() {
		return p.PropertyType()
	}

	return strings.ReplaceAll(p.PropertyType(), "array", p.Type+"[]")
}

func (p *Property) AccessorName() string {
	return strcase.ToCamel(p.Name)
}

func (p *Property) IsObject() bool {
	return p.ObjectRef != ""
}
