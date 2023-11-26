package main

import (
	"bytes"
	"embed"
	"fmt"
	"github.com/iancoleman/strcase"
	recurparse "github.com/karelbilek/template-parse-recursive"
	"github.com/sanity-io/litter"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/descriptorpb"
	"google.golang.org/protobuf/types/pluginpb"
	"io"
	"io/fs"
	"log"
	"os"
	"slices"
	"strings"
	"text/template"
)

//go:embed templates/*
var templateFiles embed.FS

var (
	PHPIncludeMap = map[ClassType]string{
		CTypeMessage: "Google\\Protobuf\\Internal\\Message",
	}

	ClassTypeTemplateMap = map[ClassType]string{
		CTypeMessage: "message.tmpl",
		CTypeEnum:    "enum.tmpl",
	}

	objectRefClassMap = map[string]*Class{}
)

type ProtoFile struct {
	Name    string
	Package string
	Classes []*Class
}

type ClassType int

const (
	CTypeMessage ClassType = iota
	CTypeEnum    ClassType = iota
)

type EnumValue struct {
	Name   string
	Number int
}

type Class struct {
	File              *ProtoFile
	Parent            *Class
	Name              string
	BaseNamespace     string
	MetadataNamespace string
	ClassPrefix       string
	Type              ClassType
	Properties        []*Property
	EnumValues        []*EnumValue
	Dependencies      []string
	options           *descriptorpb.MessageOptions
}

func (c *Class) IsMapEntry() bool {
	if c.options != nil {
		return c.options.GetMapEntry()
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

func (c *Class) addDependency(dependency string) string {
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

	return objectRefClassMap[p.ObjectRef]
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

func parseProtoFile(desc *descriptorpb.FileDescriptorProto) *ProtoFile {
	f := &ProtoFile{
		Name:    desc.GetName(),
		Package: desc.GetPackage(),
	}

	for _, message := range desc.GetMessageType() {
		f.Classes = append(f.Classes, parseMessage(f, desc.GetOptions(), message, nil))
	}

	for _, enum := range desc.GetEnumType() {
		f.Classes = append(f.Classes, parseEnum(f, desc.GetOptions(), enum, nil))
	}

	return f
}

func parseEnum(f *ProtoFile, options *descriptorpb.FileOptions, enum *descriptorpb.EnumDescriptorProto, parent *Class) *Class {
	c := newClass(CTypeEnum, f, options, enum.GetName(), nil, parent)

	for _, ev := range enum.GetValue() {
		c.EnumValues = append(c.EnumValues, newEnumValue(ev))
	}

	return c
}

func newEnumValue(ev *descriptorpb.EnumValueDescriptorProto) *EnumValue {
	return &EnumValue{
		Name:   ev.GetName(),
		Number: int(ev.GetNumber()),
	}
}

func newClass(st ClassType, file *ProtoFile, options *descriptorpb.FileOptions, name string, mo *descriptorpb.MessageOptions, parent *Class) *Class {
	ns := options.GetPhpNamespace()
	if ns == "" {
		ns = packageToNamespace(file.Package)
	}

	c := &Class{
		File:              file,
		Parent:            parent,
		Name:              name,
		Type:              st,
		BaseNamespace:     ns,
		MetadataNamespace: options.GetPhpMetadataNamespace(),
		ClassPrefix:       options.GetPhpClassPrefix(),
		options:           mo,
	}
	c.addDependency(PHPIncludeMap[st])

	objectRefClassMap["."+c.Package()] = c

	return c
}

func packageToNamespace(pack string) string {
	parts := strings.Split(pack, ".")
	for i, p := range parts {
		parts[i] = cases.Title(language.Und, cases.NoLower).String(p)
	}
	return strings.Join(parts, "\\")
}

func parseMessage(f *ProtoFile, options *descriptorpb.FileOptions, message *descriptorpb.DescriptorProto, parent *Class) *Class {
	c := newClass(CTypeMessage, f, options, message.GetName(), message.GetOptions(), parent)

	for _, field := range message.GetField() {
		c.Properties = append(c.Properties, parseField(field))
	}

	for _, nested := range message.GetNestedType() {
		f.Classes = append(f.Classes, parseMessage(f, options, nested, c))
	}

	for _, enum := range message.GetEnumType() {
		f.Classes = append(f.Classes, parseEnum(f, options, enum, c))
	}

	return c
}

func parseField(field *descriptorpb.FieldDescriptorProto) *Property {
	return &Property{
		Number:    int(field.GetNumber()),
		Name:      field.GetName(),
		Type:      phpProtoType(field.GetType()),
		ProtoType: stringProtoType(field.GetType()),
		ObjectRef: field.GetTypeName(),
		Repeated:  field.GetLabel() == descriptorpb.FieldDescriptorProto_LABEL_REPEATED,
	}
}

func generateClassesFiles(t *template.Template, f *ProtoFile) []*pluginpb.CodeGeneratorResponse_File {
	var files []*pluginpb.CodeGeneratorResponse_File
	for _, c := range f.Classes {
		var buffer bytes.Buffer

		if err := t.ExecuteTemplate(&buffer, ClassTypeTemplateMap[c.Type], c); err != nil {
			panic(err)
		}

		file := &pluginpb.CodeGeneratorResponse_File{
			Name:              proto.String(c.PHPClassFilename()),
			InsertionPoint:    nil,
			Content:           proto.String(buffer.String()),
			GeneratedCodeInfo: nil,
		}
		files = append(files, file)
	}

	return files
}

func stringProtoType(t descriptorpb.FieldDescriptorProto_Type) string {
	return strings.ReplaceAll(descriptorpb.FieldDescriptorProto_Type_name[int32(t)], "TYPE_", "")
}

func phpProtoType(t descriptorpb.FieldDescriptorProto_Type) string {
	m := map[string]string{
		"DOUBLE":   "float",
		"FLOAT":    "float",
		"INT64":    "int",
		"UINT64":   "int",
		"INT32":    "int",
		"FIXED64":  "int",
		"FIXED32":  "int",
		"BOOL":     "bool",
		"STRING":   "string",
		"MESSAGE":  "object",
		"BYTES":    "string",
		"UINT32":   "int",
		"ENUM":     "object",
		"SFIXED32": "int",
		"SFIXED64": "int",
		"SINT32":   "int",
		"SINT64":   "int",
	}

	return m[stringProtoType(t)]
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

func FQNBasename(fqn string) string {
	parts := strings.Split(fqn, "\\")
	return parts[len(parts)-1]
}

func fillDependencies(files []*ProtoFile) {
	for _, f := range files {
		for _, c := range f.Classes {
			for _, p := range c.Properties {
				// map
				if p.IsMap() == true {
					c.addDependency("Google\\Protobuf\\Internal\\MapField")
					c.addDependency("Google\\Protobuf\\Internal\\GPBType")
					c.addDependency("Google\\Protobuf\\Internal\\GPBUtil")

					// if the value is a class, add it as dependency
					if vd := p.Dependency().FindProperty("value").Dependency(); vd != nil {
						p.Type = c.addDependency(vd.FQN())
					}
				}

				// repeated only
				if p.Repeated && p.IsMap() == false {
					c.addDependency("Google\\Protobuf\\Internal\\RepeatedField")
					c.addDependency("Google\\Protobuf\\Internal\\GPBType")
					c.addDependency("Google\\Protobuf\\Internal\\GPBUtil")
				}

				// normal classes
				if p.Dependency() != nil && p.IsMap() == false {
					p.Type = c.addDependency(p.Dependency().FQN())
				}
			}
		}
	}
}

func main() {
	input, err := io.ReadAll(os.Stdin)
	if err != nil {
		panic(err)
	}

	var request pluginpb.CodeGeneratorRequest
	err = proto.Unmarshal(input, &request)
	if err != nil {
		panic(err)
	}

	var files []*ProtoFile
	for _, protoFile := range request.GetProtoFile() {
		files = append(files, parseProtoFile(protoFile))
	}

	fillDependencies(files)

	t, err := getTemplates()
	if err != nil {
		panic(err)
	}

	var resultFiles []*pluginpb.CodeGeneratorResponse_File
	for _, file := range files {
		if slices.Contains(request.GetFileToGenerate(), file.Name) {
			resultFiles = append(resultFiles, generateClassesFiles(t, file)...)
		}
	}

	response := pluginpb.CodeGeneratorResponse{
		Error:             nil,
		SupportedFeatures: proto.Uint64(uint64(pluginpb.CodeGeneratorResponse_FEATURE_NONE)),
		File:              resultFiles,
	}
	out, err := proto.Marshal(&response)
	if err != nil {
		panic(err)
	}

	_, err = fmt.Fprintf(os.Stdout, string(out))
	if err != nil {
		panic(err)
	}
}

func getTemplates() (*template.Template, error) {
	templateDir, _ := fs.Sub(templateFiles, "templates")
	t := template.New("templates")
	//t.Funcs(template.FuncMap{
	//    "templateOrDefault": func(path string, data any) string {
	//        var buffer bytes.Buffer
	//        if t.Lookup(path) == nil {
	//            path = "message/property/default.tmpl"
	//        }
	//        if err := t.ExecuteTemplate(&buffer, path, data); err != nil {
	//            panic(err)
	//        }
	//        return buffer.String()
	//    },
	//    "templateIfExists": func(path string, data any) string {
	//        var buffer bytes.Buffer
	//        if err := t.ExecuteTemplate(&buffer, path, data); err != nil {
	//            panic(err)
	//        }
	//        return buffer.String()
	//    },
	//    "templateExists": func(path string) bool {
	//        return t.Lookup(path) != nil
	//    },
	//})
	return recurparse.TextParseFS(
		t,
		templateDir,
		"*.tmpl",
	)
}

func dump(value ...interface{}) {
	log.Println(litter.Sdump(value...))
}
