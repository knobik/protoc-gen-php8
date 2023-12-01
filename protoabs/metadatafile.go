package protoabs

import (
	"encoding/base64"
	"fmt"
	"github.com/iancoleman/strcase"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/descriptorpb"
	"path/filepath"
	"strings"
)

type MetadataFile struct {
	name              string
	metadataNamespace string
	fileDescriptorSet *descriptorpb.FileDescriptorSet
}

func (m *MetadataFile) ClassName() string {
	return strcase.ToCamel(strings.ReplaceAll(filepath.Base(m.name), ".proto", ""))
}

func (m *MetadataFile) MessageAsString() string {
	out, err := proto.Marshal(m.fileDescriptorSet)
	if err != nil {
		panic(err)
	}

	return string(out)
}

func (m *MetadataFile) MessageAsBase64String() string {
	return base64.StdEncoding.EncodeToString([]byte(m.MessageAsString()))
}

func (m *MetadataFile) Namespace() string {
	return m.metadataNamespace
}

func (m *MetadataFile) PHPClassDirectory() string {
	return strings.Join(strings.Split(m.Namespace(), "\\"), "/")
}

func (m *MetadataFile) PHPClassFilename() string {
	return fmt.Sprintf("%s/%s.php", m.PHPClassDirectory(), m.ClassName())
}

func (m *MetadataFile) FQN() string {
	return m.Namespace() + "\\" + m.ClassName()
}

func NewMetadataFile(desc *descriptorpb.FileDescriptorProto, pf *ProtoFile) *MetadataFile {
	clonedDesc := proto.Clone(desc).(*descriptorpb.FileDescriptorProto)
	clonedDesc.SourceCodeInfo = nil

	ns := desc.GetOptions().GetPhpMetadataNamespace()
	if ns == "" {
		ns = PackageToNamespace("GPBMetadata." + desc.GetPackage())
	}

	set := []*descriptorpb.FileDescriptorProto{clonedDesc}
	m := &MetadataFile{
		name:              desc.GetName(),
		metadataNamespace: ns,
		fileDescriptorSet: &descriptorpb.FileDescriptorSet{File: set},
	}

	for _, c := range pf.Classes {
		c.Metadata = m
		if c.Type == CTypeMessage {
			c.AddDependency(m.FQN())
		}
	}

	return m
}
