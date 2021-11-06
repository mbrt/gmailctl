package main

import (
	"github.com/mbrt/gmailctl/cmd/gmailctl/cmd"
	"github.com/mbrt/gmailctl/cmd/gmailctl/localcred"
)

func main() {
	cmd.APIProvider = localcred.Provider{}
	cmd.Execute()
}
