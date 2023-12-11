package protoabs

type ProtoFile struct {
	Opt     *Options
	Name    string
	Package string
	Classes []*Class
}

func NewProtoFile(name string, pack string, opt *Options) *ProtoFile {
	return &ProtoFile{
		Opt:     opt,
		Name:    name,
		Package: pack,
	}
}
