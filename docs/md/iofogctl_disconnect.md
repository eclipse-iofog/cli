## iofogctl disconnect

Disconnect from an ioFog cluster

### Synopsis

Disconnect from an ioFog cluster.

This will NOT teardown any components of the cluster. If you would like to tear down deployments, use the delete command.

This will leave the corresponding namespace empty.

```
iofogctl disconnect [flags]
```

### Examples

```
iofogctl disconnect -n NAMESPACE
```

### Options

```
  -h, --help   help for disconnect
```

### Options inherited from parent commands

```
      --config string      CLI configuration file (default is ~/.iofog/config.yaml)
  -n, --namespace string   Namespace to execute respective command within (default "default")
  -q, --quiet              Toggle for displaying verbose output
  -v, --verbose            Toggle for displaying verbose output of API client
```

### SEE ALSO

* [iofogctl](iofogctl.md)	 - 

###### Auto generated by spf13/cobra on 7-Aug-2019