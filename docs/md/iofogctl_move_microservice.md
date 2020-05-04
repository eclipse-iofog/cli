## iofogctl move microservice

Move a Microservice to another agent in the same Namespace

### Synopsis

Move a Microservice to another agent in the same Namespace

```
iofogctl move microservice NAME AGENT_NAME [flags]
```

### Examples

```
iofogctl move microservice NAME AGENT_NAME
```

### Options

```
  -h, --help   help for microservice
```

### Options inherited from parent commands

```
      --debug              Toggle for displaying verbose output of API clients (HTTP and SSH)
      --detached           Use/Show detached resources
  -n, --namespace string   Namespace to execute respective command within (default "default")
  -v, --verbose            Toggle for displaying verbose output of iofogctl
```

### SEE ALSO

* [iofogctl move](iofogctl_move.md)	 - Move an existing resources inside the current Namespace


