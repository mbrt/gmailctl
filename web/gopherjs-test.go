package main

import "github.com/gopherjs/gopherjs/js"

func main() {
	js.Global.Set("pet", map[string]interface{}{
		"New": New,
	})
}

type Pet struct {
	name string
}

func New(name string) *js.Object {
	return js.MakeWrapper(&Pet{name})
}

func (p *Pet) Name() string {
	return p.name
}

func (p *Pet) SetName(name string) {
	p.name = name
}
