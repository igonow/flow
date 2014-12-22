package flow

import (
	"reflect"
)

var network *Net //manager

type graph struct {
	Graph // graph "superclass" embedded
}

type Net struct {
	graph graph
}

type Context struct {
	Output chan interface{}
}

func (c *Context) Init() {
	c.Output = make(chan interface{})
}

func NewNet() *Net {
	if network == nil {
		network = new(Net)
		network.graph.InitGraphState()
	}
	return network
}

func Network() *Net {
	return NewNet()
}

func (n *Net) Run() {
	RunNet(&n.graph)
}

func (n *Net) Wait() <-chan struct{} {
	return n.graph.Wait()
}

func (n *Net) Add(cs ...interface{}) {
	for _, c := range cs {
		n.mustComponent(c)
		network.graph.Add(c, n.getName(c))
		if tag, ok := n.getInOutput(c, "input"); ok {
			n.graph.MapInPort(n.getName(c), n.getName(c), tag)
		}
		if tag, ok := n.getInOutput(c, "output"); ok {
			n.graph.MapOutPort(n.getName(c), n.getName(c), tag)
		}
	}
}

func (n *Net) Connect(sender, receiver interface{}) {
	n.mustComponent(sender)
	n.mustComponent(receiver)
	n.graph.Connect(n.getName(sender),
		n.getPort(sender, "sender"),
		n.getName(receiver),
		n.getPort(receiver, "receiver"))
}

func (n *Net) mustComponent(c interface{}) {
	component := reflect.ValueOf(c).Elem().FieldByName("Component")
	if !component.IsValid() || component.Type().Name() != "Component" {
		panic("argument is not a valid component instance, forget combine flow.Component ?")
	}
}

func (n *Net) getInOutput(c interface{}, tp string) (tag string, ok bool) {
	s := reflect.TypeOf(c).Elem()
	for i := 0; i < s.NumField(); i++ {
		if tag = s.Field(i).Tag.Get("flow"); tag != "" {
			if tag == tp {
				return s.Field(i).Name, true
			}
		}
	}
	return
}

func (n *Net) getPort(c interface{}, tp string) string {
	var tag string
	s := reflect.TypeOf(c).Elem()
	for i := 0; i < s.NumField(); i++ {
		if tag = s.Field(i).Tag.Get("flow"); tag != "" {
			if tag == tp {
				return s.Field(i).Name
			}
		}
	}
	panic(`component must have a tag as flow:"xxx"`)
}

func (n *Net) SetInPort(c, channel interface{}) {
	n.mustComponent(c)
	if !n.graph.SetInPort(n.getName(c), channel) {
		panic("set in port error")
	}
}

func (n *Net) SetOutPort(c, channel interface{}) {
	n.mustComponent(c)
	n.graph.SetOutPort(n.getName(c), channel)
}

func (n *Net) getName(c interface{}) string {
	tp := reflect.TypeOf(c).Elem()
	return tp.PkgPath() + "/" + tp.Name()
}
