# Interfacer
English | [中文文档](doc/user/README_CN.md)

Implement new methods in the interface everywhere

快速给现有接口添加新方法，给相关结构**自动生成默认实现**

![effect](pic/golang.gif)

## 🎉 Effect

![effect](pic/effect.gif)

## 🤔️ Why this tool
When you add a method to an interface, the corresponding implementation cannot be found after adding the method.
Because of the addition of an additional method, the previous implementations are all invalid.
In general, adding a method to an interface is cumbersome and time-consuming.

## 📚 Getting started

### 📜 Install 
There are three ways to install it, including:
1. Binary file

    you can get the latest version by the link: [interfacer](https://github.com/SimFG/interfacer/blob/main/interfacer?raw=true)

2. Install by `go install`

   **Tips**: If you use the windows, you can download the source code and build the tool.

    ```shell
    go install github.com/SimFG/interfacer
    ```

3. Source build

   a. clone the repository
    ```shell
    git clone github.com/SimFG/interfacer
    ```
   b. build
    ```shell
    go  build interfacer.go interfacer_handle.go
    ```

### 🔬 Detailed usage
link: [USAGE](doc/user/USAGE.md)

### 🪧 Tips
Some problems may be encountered during use, as the tool is currently under development. You can give me an issue If you encounter any problems using.

## 🧭 Roadmap
- Dealing with generated code out of order of comments
  - for interface ✅
  - for struct method
- Exclude dirs or files
  - accurate ✅
  - fuzzy matching
- Debug mode, print the detail info ✅
- It will be ignored if the new method has existed ✅
- Scan many third modules' files and then get all the implements of the interface, like the protobuf interface ✅
- Check the input params ✅
- Customer the configuration yaml file path ✅
- Update the Chinese README file ✅
- Beautify the README file ✅
- Dynamic display of program running time ✅
- Show scan progress ✅
- Using the function name and receiver type to check whether the method has existed  ✅
- Configure whether to add a new line when generating the new method for the interface ✅
- Handle the no name params in the method, like `foo(bool)`
- Configure whether to add the new method if the inner struct or interface has contained the new method 
- Individually add a method for different structs that implements different interfaces of the third modules
- Write the method content easily by the way similar to writing interface
- Contributing doc
- Handle the more params that are the same type, like `foo(a, b bool)`
- Idea plugin
- Customer the insert position
- Generate more complex default method implement by the template
- Support to generate more methods
- More readable codes
- Handle the un-import interface or struct param when adding the new method 
- Handle some special struct or interface
- More useful tool of reading and writing `go` file
- Ast example doc