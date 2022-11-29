# Usage
## Parameter description
There are two ways to set parameters, including the `yaml` file and the command params.
### Example
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
    sub_modules:
      -
        project_dir: "/Users/derek/fubang/interfacer/example/implemente"
        project_module: "github.com/SimFG/interfacer/example/implemente"
        interface_full_name: "github.com/SimFG/interfacer/example/implemente/typee.Node"
        method: "GetStatic()"
        exclude_dirs:
          - "coord"
    ```

2. command params

    ```bash
    ./interfacer
        --project-dir=/Users/derek/fubang/interfacer/example/all
        --project-module=github.com/SimFG/interfacer/example/all 
        --interface=github.com/SimFG/interfacer/example/all/i.Component 
        --method="Hello(f int64) (int, error)" 
        --returns="0,nil"
    ```
### Param meaning
- project dir: full project dir
- project module: it can be found in the `go.mod` file
- interface: the interface you want to add a new method to it. And its full name is required
- method: declaration of the newly added method
- returns: the default return values of new method
- exclude dirs: these dirs will be ignored
- ignore_structs: ignore structs when generating the method
- enable_debug: set true if you find a problem while using this tool, and the processing speed will slow because it needs to write a lot of logs to the files.
- enable_record: set true if you want to get the relations between all structs and interfaces.
- sub_modules: the third modules' configuration. It's suitable to add a new method when the interface in the third module add a new method, like the rpc service in the protobuf.