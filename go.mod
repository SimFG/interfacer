module github.com/SimFG/interfacer

go 1.18

require (
	github.com/SimFG/interfacer/scanner v0.0.1
	github.com/SimFG/interfacer/tool v0.0.1
	github.com/SimFG/interfacer/writer v0.0.1
	github.com/samber/lo v1.33.0
	github.com/spf13/cobra v1.6.1
	go.uber.org/zap v1.23.0
	gopkg.in/yaml.v3 v3.0.1

)

require (
	github.com/SimFG/interfacer/progress v0.0.1 // indirect
	github.com/inconshreveable/mousetrap v1.0.1 // indirect
	github.com/spf13/pflag v1.0.5 // indirect
	go.uber.org/atomic v1.7.0 // indirect
	go.uber.org/multierr v1.6.0 // indirect
	golang.org/x/exp v0.0.0-20220303212507-bbda1eaf7a17 // indirect
)

replace (
	github.com/SimFG/interfacer/progress => ./progress
	github.com/SimFG/interfacer/scanner => ./scanner
	github.com/SimFG/interfacer/tool => ./tool
	github.com/SimFG/interfacer/writer => ./writer
)
