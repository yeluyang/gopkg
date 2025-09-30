package streamgraph

func New[Code MsgCode, Visitor MsgVisitor](v Visitor) {}

type config struct {
	routines uint
}

type srcConfig[Code MsgCode, Visitor MsgVisitor] struct {
	config
	Source[Code, Visitor]
}

type builder[Code MsgCode, Visitor MsgVisitor] struct {
	codes   map[Code]bool
	sources []*srcConfig[Code, Visitor]
	sinks   map[Code]*config
	actors  map[Code]*config
}

func (b *builder[Code, Visitor]) Build() *streamGraph[Code, Visitor] {
	return nil
}
