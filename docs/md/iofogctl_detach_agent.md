## iofogctl detach agent

Detaches an Agent

### Synopsis

Detaches an Agent.

The Agent will be deprovisioned from the Controller within the namespace.
The Agent will be removed from Controller.

You cannot detach unprovisioned Agents.

The Agent stack will not be uninstalled from the host.

```
iofogctl detach agent NAME [flags]
```

### Examples

```
iofogctl detach agent NAME
```

### Options

```
      --force   Detach agent, even if it still uses resources
  -h, --help    help for agent
```

### Options inherited from parent commands

```
      --detached           Use/Show detached resources
      --http-verbose       Toggle for displaying verbose output of API client
  -n, --namespace string   Namespace to execute respective command within (default "default")
  -v, --verbose            Toggle for displaying verbose output of iofogctl
```

### SEE ALSO

* [iofogctl detach](iofogctl_detach.md)	 - Detach an existing ioFog resource from its ECN


