# Development
## Difficulty
### The relationship between struct and interface
1. embedded
The struct can embed other structs and interfaces, and the interface can embed other interfaces, which makes more complicate to judging whether there is an implementation relationship
between the struct and interface.
2. scan order
The inner struct or inner interface, maybe uncertain when scanning the current file. Because you can't know whether the inner belongs to the current project.
If yes, you should care about its relationship to the current structure or interface. Otherwise, you don't need to care.
3. full name
In different packages, interfaces or structs with the same name are allowed. So the interface or struct name should contain the package name.
This also needs to be noted for the parameter types of methods.
4. implement relationship

## Feature
### No error
If some functions occur an error, it will panic directly, because the `interfacer` is  a small and simple developing tool.