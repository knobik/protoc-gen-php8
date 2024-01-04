package protoabs

type ProtoFile struct {
	Name    string
	Package string
	Classes []*Class
}

func NewProtoFile(name string, pack string) *ProtoFile {
	return &ProtoFile{
		Name:    name,
		Package: pack,
	}
}
