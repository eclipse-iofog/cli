## iofogctl create catalogitem

Create a catalog item

### Synopsis

Create a catalog item on the ioFog controller

```
iofogctl create catalogitem NAME [flags]
```

### Examples

```
iofogctl create catalogitem NAME --x86 x86_IMAGE --arm arm_IMAGE --registry <remote|local> --description DESCRIPTION
```

### Options

```
      --arm string           Container image to use on arm agents
  -d, --description string   Description of catalog item purpose
  -h, --help                 help for catalogitem
  -r, --registry string      Container registry to use. Either 'remote' or 'local'
      --x86 string           Container image to use on x86 agents
```

### Options inherited from parent commands

```
      --config string      CLI configuration file (default is ~/.iofog/config.yaml)
  -n, --namespace string   Namespace to execute respective command within (default "default")
  -v, --verbose            Toggle for displaying verbose output of API client
```

### SEE ALSO

* [iofogctl create](iofogctl_create.md)	 - Create a resource

###### Auto generated by spf13/cobra on 20-Aug-2019