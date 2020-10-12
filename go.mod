module github.com/croomes/kubectl-plugin

go 1.12

replace github.com/replicatedhq/troubleshoot => /Users/ferran/repos/troubleshoot

require (
	github.com/ahmetalpbalkan/go-cursor v0.0.0-20131010032410-8136607ea412
	github.com/fatih/color v1.7.0
	github.com/golang/groupcache v0.0.0-20190702054246-869f871628b6 // indirect
	github.com/gophercloud/gophercloud v0.13.0 // indirect
	github.com/hashicorp/go-cleanhttp v0.5.1 // indirect
	github.com/manifoldco/promptui v0.3.2
	github.com/mattn/go-isatty v0.0.9
	github.com/mattn/go-runewidth v0.0.9 // indirect
	github.com/mholt/archiver v3.1.1+incompatible
	github.com/mitchellh/go-wordwrap v1.0.1 // indirect
	github.com/nwaples/rardecode v1.1.0 // indirect
	github.com/pkg/errors v0.9.1
	github.com/replicatedhq/termui/v3 v3.1.1-0.20200811145416-f40076d26851
	github.com/replicatedhq/troubleshoot v0.9.44
	github.com/spf13/cobra v0.0.5
	github.com/spf13/viper v1.4.0
	github.com/tj/go-spin v1.1.0
	gopkg.in/yaml.v2 v2.3.0
	k8s.io/apimachinery v0.18.3
	k8s.io/cli-runtime v0.18.0
	k8s.io/client-go v0.18.2
)
