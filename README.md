# node-age-controller

A Kubernetes [controller](https://kubernetes.io/docs/concepts/architecture/controller/) to cordon nodes older than a specified age.

It is designed to be used in conjunction with tools like the [cluster-autoscaler](https://github.com/kubernetes/autoscaler/tree/master/cluster-autoscaler) to have nodes removed from service and replaced with new, fresh nodes. This can dramatically increase your patching abilities, because you stop adding new pods to nodes that are old, and may have unpatched vulnerabilities on them.

# Configuration

The controller has numerous configuration options, *please read them carefully*, especially the thresholds!1

## Thresholds

:heavy_exclamation_mark: This controller will cordon nodes. Cordoned nodes are removed from serving services, even if they still have healthy pods on them, see https://github.com/kubernetes/kubernetes/issues/65013 for more information.

*Make sure you set reasonable thresholds to ensure you don't accidentally cordon all nodes*

The controller allows you to set some configuration options either via Environment variables or via the command line to ensure you don't cordon too many nodes.

```
--max-nodes=3             The max number of nodes that can be cordoned at one time
--min-available-nodes=3   How many nodes must be uncordoned before we attempt to cordon a node
```

You can set these options via the following environment variables:

```
MAX_NODES=3
MIN_AVAILABLE_NODES=5
```

## Max Node Age Configuration

The flag `max-node-age` takes a Golang [Duration](https://golang.org/pkg/time/#Duration) type. An example might be:

```
--max-node-age=720h # 30 days
```

If you don't specify a valid duration, the controller will fail to start.

You can also set this via the environment variable `MAX_NODE_AGE=720h`

## Node opt-in/out-out

By default, the controller will operate on all nodes in the cluster. You can explicitly set if a node will be cordoned using the annotation `age.briggs.io/ignore`. For example:

```
age.briggs.io/ignore: "true" # Don't cordon this node
age.briggs.io/ignore: "false" # Do cordon this node
```

## Dry run

You can place the controller in "dry-run" mode where it will specify what it would do, but never actually cordon any nodes by specifying the `--dry-run` flag on startup 



