package main

import (
	"github.com/croomes/kubectl-plugin/cmd/preflight/cli"
	_ "k8s.io/client-go/plugin/pkg/client/auth"
)

func main() {
	cli.InitAndExecute()
}
