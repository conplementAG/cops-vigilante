# cops-vigilante

There are some problems in our day-to-day operations at [Conplement AG](https://www.conplement.de/) which are simply hair pulling :)

CoreOps Vigilante project is an attempt to take matters in our own hands, and try to implement some automation or self-healing capabilities
to such problems. "Tasks" which are performed via cops-vigilante are explained below in the Tasks section.

# Installation

Installation is supported via Helm.

TODO: GitHub pages and project setup https://medium.com/@mattiaperi/create-a-public-helm-chart-repository-with-github-pages-49b180dbb417

# Tasks

## AKS Windows nodes SNAT issue with complex networking setups (VPN, hub-spoke networks, etc.)

There is a SNAT issue regarding Windows nodes in AKS clusters, which is not fixed for a very long time now. To summarize, the issue is 
that in AzureCNI AKS networking configuration, Windows nodes always perform SNAT, meaning the pods sending packages will have their 
originating IPs replaced by the node IP. This should normally not be an issue (to have NAT-ing in a network), but the problem is that 
this does not work correctly (packages are dropped "randomly", performance issues due to port exhaustion etc.). 

The only fix currently is to exclude certain CIDR ranges from this functionality. To do this, a certain 10-azure.conflist 
has to be modified and loaded, which can be done either via VM Extensions or host-process windows containers. 

The problem is however, to apply this config, a new pod has to be scheduled on a node and there is no process which runs on the 
node to read this config in the background. Config is applied once kubelet is activated on pod scheduling, which in turn 
calls azurecni, which then loads the config. As the config creation / update does not occur atomically with node creation, 
we need a process which will schedule pod creations for some time, so that it is made sure the config is loaded at some point. 

This SNAT task in cops-vigilante keeps a track of ready windows nodes, and schedules windows containers on them for ca. 30 minutes, 
after which the nodes are marked as "healed". Windows container used is as small as possible, currently we use the 
mcr.microsoft.com/oss/kubernetes/pause image, which is under 1 MB size!

If the node is considered healed, an annotation 
```
cops-vigilante-snat-node-healed: "true"
```
will be added. Setting this annotation to "false", or removing it completely, will restart the healing process for this node.

# General features

## TLS via CertManager

To enable TLS for internal HTTP endpoint, you can use the Helm flag "create_certificates". Also make sure to set the "tls" option in the config
section to true to load these certificates.

# Development

Check [CONTRIBUTING.md](CONTRIBUTING.md) 

# License

Check [LICENSE](LICENSE)