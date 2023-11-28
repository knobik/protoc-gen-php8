package protoabs

import (
	"fmt"
	"google.golang.org/protobuf/types/descriptorpb"
	"strings"
)

type ClassType int

const (
	CTypeMessage ClassType = iota
	CTypeEnum    ClassType = iota
)

type EnumValue struct {
	Name   string
	Number int
}

func NewEnumValue(ev *descriptorpb.EnumValueDescriptorProto) *EnumValue {
	return &EnumValue{
		Name:   ev.GetName(),
		Number: int(ev.GetNumber()),
	}
}

type Class struct {
	File          *ProtoFile
	Parent        *Class
	Name          string
	BaseNamespace string
	ClassPrefix   string
	Type          ClassType
	Properties    []*Property
	EnumValues    []*EnumValue
	Dependencies  []string
	Metadata      *MetadataFile
	Options       *descriptorpb.MessageOptions
}

func (c *Class) IsMapEntry() bool {
	if c.Options != nil {
		return c.Options.GetMapEntry()
	}

	return false
}

func (c *Class) IsEnum() bool {
	return c.Type == CTypeEnum
}

func (c *Class) Enums() []*Property {
	var list []*Property
	for _, p := range c.Properties {
		if p.IsEnum() {
			list = append(list, p)
		}
	}

	return list
}

func (c *Class) FindProperty(name string) *Property {
	for _, p := range c.Properties {
		if p.Name == name {
			return p
		}
	}

	return nil
}

func (c *Class) AddDependency(dependency string) string {
	alias, base := c.aliasDependency(dependency)
	for _, include := range c.Dependencies {
		if include == alias {
			return base
		}
	}

	c.Dependencies = append(c.Dependencies, alias)

	return base
}

func (c *Class) aliasDependency(dependency string) (string, string) {
	base := FQNBasename(dependency)
	count := 0
	for _, include := range c.Dependencies {
		includeBase := FQNBasename(include)
		if includeBase == base && dependency != include {
			count++
		}
	}
	if count > 0 {
		base = fmt.Sprintf("%s%d", base, count)
		dependency += " as " + base
	}

	return dependency, base
}

func (c *Class) ClassName() string {
	return c.ClassPrefix + c.Name
}

func (c *Class) Namespace() string {
	if c.Parent != nil {
		return c.Parent.Namespace() + "\\" + c.Parent.ClassName()
	}

	return c.BaseNamespace
}

func (c *Class) PHPClassDirectory() string {
	return strings.Join(strings.Split(c.Namespace(), "\\"), "/")
}

func (c *Class) PHPClassFilename() string {
	return fmt.Sprintf("%s/%s.php", c.PHPClassDirectory(), c.ClassName())
}

func (c *Class) FQN() string {
	if c.Parent != nil {
		return c.Parent.FQN() + "\\" + c.ClassName()
	}
	return c.Namespace() + "\\" + c.ClassName()
}

func (c *Class) Package() string {
	if c.Parent != nil {
		return fmt.Sprintf("%s.%s", c.Parent.Package(), c.Name)
	}

	return fmt.Sprintf("%s.%s", c.File.Package, c.Name)
}

func NewClass(st ClassType, file *ProtoFile, options *descriptorpb.FileOptions, name string, mo *descriptorpb.MessageOptions, parent *Class) *Class {
	ns := options.GetPhpNamespace()
	if ns == "" {
		ns = PackageToNamespace(file.Package)
	}

	c := &Class{
		File:          file,
		Parent:        parent,
		Name:          name,
		Type:          st,
		BaseNamespace: ns,
		ClassPrefix:   options.GetPhpClassPrefix(),
		Options:       mo,
	}
	c.AddDependency(PHPIncludeMap[st])

	ObjectRefClassMap["."+c.Package()] = c

	return c
}