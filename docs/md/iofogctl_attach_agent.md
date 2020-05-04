## iofogctl attach agent

Attach an Agent to an existing Namespace

### Synopsis

Attach an Agent to an existing Namespace.

The Agent will be provisioned with the Controller within the Namespace.

Can be used after detach command to re-provision the Agent. Can also be used with Agents that have not been detached.


```
iofogctl attach agent NAME [flags]
```

### Examples

```
iofogctl attach agent NAME --detached
iofogctl attach agent NAME --host AGENT_HOST --user SSH_USER --port SSH_PORT --key SSH_PRIVATE_KEY_PATH
```

### Options

```
  -h, --help          help for agent
      --host string   Hostname of remote host
      --key string    Path to private SSH key
      --port int      Port number that iofogctl uses to SSH into remote hosts (default 22)
      --user string   Username of remote host
```

### Options inherited from parent commands

```
      --debug              Toggle for displaying verbose output of API clients (HTTP and SSH)
      --detached           Use/Show detached resources
  -n, --namespace string   Namespace to execute respective command within (default "default")
  -v, --verbose            Toggle for displaying verbose output of iofogctl
```

### SEE ALSO

* [iofogctl attach](iofogctl_attach.md)	 - Attach an existing ioFog resource to Control Plane


