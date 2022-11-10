# Interfacer
[中文文档](README_CN.md)

Implement new methods in the interface everywhere

**Tips**: some problems may be encountered during use, as the tool is currently under development.

## Have you ever encountered such a problem
When you add a method to an interface, the corresponding implementation cannot be found after adding the method. 
Because of the addition of an additional method, the previous implementations are all invalid. 
In general, adding a method to an interface is cumbersome and time-consuming.

## Effect
![effect](effect.gif)

## Getting started
### install
```shell
go install github.com/SimFG/interfacer
```
**Tips**: If you use the windows, you can download the source code and build the tool.

### Parameter description
There are two ways to set parameters, which are:
1. yaml file
```yaml
project_dir: "/Users/derek/fubang/interfacer/example/all"
project_module: "github.com/SimFG/interfacer/example/all"
interface_full_name: "github.com/SimFG/interfacer/example/all/i.Component"
new_method: "Hello(f int64) (int, error)"
return_default_values: "0,nil"
write_paths:
  - "github.com/SimFG/interfacer/example/proxy/all/s.Node,/Users/derek/xxx/interfacer/example/all/s/st.go"
exclude_dirs:
  - "foo"
enable_debug: false
enable_record: false
```
2. command params
```bash
go run interfacer.go
    --project-dir=/Users/derek/fubang/interfacer/example/all
    --project-module=github.com/SimFG/interfacer/example/all 
    --interface=github.com/SimFG/interfacer/example/all/i.Component 
    --method="Hello(f int64) (int, error)" 
    --returns="0,nil"
```
- project dir: full project dir
- project module: it can be found in the `go.mod` file
- interface: the interface you want to add a new method to it. And its full name is required
- method: declaration of the newly added method
- returns: the default return values of new method
- exclude dirs: these dirs will be ignored
- enable_debug: set true if you find a problem while using this tool, and the processing speed will slow because it needs to write a lot of logs to the files.
- enable_record: set true if you want to get the relations between all structs and interfaces.

### Source build
1. clone the repository
```shell
git clone github.com/SimFG/interfacer
```
2. run
```shell
go run interfacer.go interfacer_handle.go
```

## Roadmap
- Dealing with generated code out of order of comments
  - for interface ✅
  - for struct method
- Exclude dirs or files
  - accurate ✅
  - fuzzy matching
- Debug mode, print the detail info ✅
- It will be ignored if the new method has existed ✅
- Scan one third part file and then get all the implements of the interface, like the protobuf interface ✅
- Check the input params
- Customer the configuration yaml file path ✅
- Write the method content easily by the way similar to writing interface
- Idea plugin
- Customer the insert position
- Support to generate more methods
- More readable codes
- Handle the un-import interface or struct param when adding the new method 
- Handle some special struct or interface
- More useful tool of reading and writing `go` file
- Ast example project