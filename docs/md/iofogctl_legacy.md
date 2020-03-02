## iofogctl legacy

Execute commands using legacy CLI

### Synopsis

Execute commands using legacy CLI

```
iofogctl legacy resource NAME COMMAND ARGS... [flags]
```

### Examples

```
iofogctl get all
iofogctl legacy controller NAME iofog
iofogctl legacy agent NAME status
```

### Options

```
  -h, --help   help for legacy
```

### Options inherited from parent commands

```
      --detached           Use/Show detached resources
      --http-verbose       Toggle for displaying verbose output of API client
  -n, --namespace string   Namespace to execute respective command within (default "testing")
  -v, --verbose            Toggle for displaying verbose output of iofogctl
```

### SEE ALSO

* [iofogctl](iofogctl.md)	 - 


