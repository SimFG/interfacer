module github.com/SimFG/interfacer/scanner

go 1.18

require (
	github.com/SimFG/interfacer/tool v0.0.1
	go.uber.org/zap v1.23.0
	github.com/samber/lo v1.33.0
)

require (
	go.uber.org/atomic v1.7.0 // indirect
	go.uber.org/multierr v1.6.0 // indirect
	golang.org/x/exp v0.0.0-20220303212507-bbda1eaf7a17 // indirect
)

replace github.com/SimFG/interfacer/tool => ../tool
