package main

import (
	"github.com/gopherjs/gopherjs/js"

	"github.com/mbrt/gmailctl/internal/engine/apply"
	"github.com/mbrt/gmailctl/internal/engine/config/v1alpha3"
)

func main() {
	res, err := exampleDiff()
	if err != nil {
		js.Global.Get("console").Call("error", err.Error())
		return
	}
	js.Global.Get("console").Call("log", res)
}

func exampleDiff() (string, error) {
	cfg0 := v1alpha3.Config{Version: "v1alpha3"}
	cfg1 := v1alpha3.Config{
		Version: "v1alpha3",
		Rules: []v1alpha3.Rule{
			{
				Filter: v1alpha3.FilterNode{
					From: "foobar",
				},
				Actions: v1alpha3.Actions{
					Archive: true,
				},
			},
		},
	}
	pres0, err := apply.FromConfig(cfg0)
	if err != nil {
		return "", err
	}
	pres1, err := apply.FromConfig(cfg1)
	if err != nil {
		return "", err
	}
	d, err := apply.Diff(pres0.GmailConfig, pres1.GmailConfig)
	if err != nil {
		return "", err
	}
	return d.String(), nil
}
