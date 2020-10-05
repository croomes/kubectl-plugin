package main

import (
	"github.com/croomes/kubectl-plugin/cmd/bundle/cli"
	_ "k8s.io/client-go/plugin/pkg/client/auth"
)

func main() {
	cli.InitAndExecute()
}
