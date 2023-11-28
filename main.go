package main

import (
	"bytes"
	"embed"
	"errors"
	"fmt"
	recurparse "github.com/karelbilek/template-parse-recursive"
	"github.com/sanity-io/litter"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/descriptorpb"
	"google.golang.org/protobuf/types/pluginpb"
	"io"
	"io/fs"
	"log"
	"os"
	"protoc-gen-php8/protoabs"
	"slices"
	"strings"
	"text/template"
)

//go:embed templates/*
var templateFiles embed.FS

func parseProtoFile(desc *descriptorpb.FileDescriptorProto) *protoabs.ProtoFile {
	f := protoabs.NewProtoFile(desc.GetName(), desc.GetPackage())

	for _, message := range desc.GetMessageType() {
		f.Classes = append(f.Classes, parseMessage(f, desc.GetOptions(), message, nil))
	}

	for _, enum := range desc.GetEnumType() {
		f.Classes = append(f.Classes, parseEnum(f, desc.GetOptions(), enum, nil))
	}

	return f
}

func parseEnum(f *protoabs.ProtoFile, options *descriptorpb.FileOptions, enum *descriptorpb.EnumDescriptorProto, parent *protoabs.Class) *protoabs.Class {
	c := protoabs.NewClass(protoabs.CTypeEnum, f, options, enum.GetName(), nil, parent)

	for _, ev := range enum.GetValue() {
		c.EnumValues = append(c.EnumValues, protoabs.NewEnumValue(ev))
	}

	return c
}

func parseMessage(f *protoabs.ProtoFile, options *descriptorpb.FileOptions, message *descriptorpb.DescriptorProto, parent *protoabs.Class) *protoabs.Class {
	c := protoabs.NewClass(protoabs.CTypeMessage, f, options, message.GetName(), message.GetOptions(), parent)

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

func parseField(field *descriptorpb.FieldDescriptorProto) *protoabs.Property {
	return &protoabs.Property{
		Number:    int(field.GetNumber()),
		Name:      field.GetName(),
		Type:      phpProtoType(field.GetType()),
		ProtoType: stringProtoType(field.GetType()),
		ObjectRef: field.GetTypeName(),
		Repeated:  field.GetLabel() == descriptorpb.FieldDescriptorProto_LABEL_REPEATED,
	}
}

func generateClassesFiles(t *template.Template, f *protoabs.ProtoFile) []*pluginpb.CodeGeneratorResponse_File {
	var files []*pluginpb.CodeGeneratorResponse_File
	for _, c := range f.Classes {
		var buffer bytes.Buffer

		if err := t.ExecuteTemplate(&buffer, protoabs.ClassTypeTemplateMap[c.Type], c); err != nil {
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

func generateMetadataFile(t *template.Template, m *protoabs.MetadataFile) []*pluginpb.CodeGeneratorResponse_File {
	var buffer bytes.Buffer

	if err := t.ExecuteTemplate(&buffer, "metadata.tmpl", m); err != nil {
		panic(err)
	}

	file := &pluginpb.CodeGeneratorResponse_File{
		Name:    proto.String(m.PHPClassFilename()),
		Content: proto.String(buffer.String()),
	}
	files := []*pluginpb.CodeGeneratorResponse_File{file}

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

func fillDependencies(files []*protoabs.ProtoFile) {
	for _, f := range files {
		for _, c := range f.Classes {
			for _, p := range c.Properties {
				// map
				if p.IsMap() == true {
					c.AddDependency("Google\\Protobuf\\Internal\\MapField")
					c.AddDependency("Google\\Protobuf\\Internal\\GPBType")
					c.AddDependency("Google\\Protobuf\\Internal\\GPBUtil")

					// if the value is a class, add it as dependency
					if vd := p.Dependency().FindProperty("value").Dependency(); vd != nil {
						p.Type = c.AddDependency(vd.FQN())
					}
				}

				// repeated only
				if p.Repeated && p.IsMap() == false {
					c.AddDependency("Google\\Protobuf\\Internal\\RepeatedField")
					c.AddDependency("Google\\Protobuf\\Internal\\GPBType")
					c.AddDependency("Google\\Protobuf\\Internal\\GPBUtil")
				}

				// normal classes
				if p.Dependency() != nil && p.IsMap() == false {
					p.Type = c.AddDependency(p.Dependency().FQN())
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

	var files []*protoabs.ProtoFile
	var metadataFiles []*protoabs.MetadataFile

	for _, protoFile := range request.GetProtoFile() {
		if protoFile.GetSyntax() != "proto3" {
			panic(errors.New("only proto3 syntax is supported"))
		}

		pf := parseProtoFile(protoFile)
		files = append(files, pf)
		metadataFiles = append(metadataFiles, protoabs.NewMetadataFile(protoFile, pf))
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
	for _, metadataFile := range metadataFiles {
		resultFiles = append(resultFiles, generateMetadataFile(t, metadataFile)...)
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
