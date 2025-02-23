Example multi-cluster app exposed to tailnet using Tailscale Kubernetes operator's HA Ingress.

This contains:

- two simple apps, which each return app name and region in which it is running
This is meant to be deployed in GKE (apps determine region by looking at /etc/resolv.conf file which on GKE contains a region hint).
- a ProxyGroup for each cluster with 2 replicas

For each app:
- a 'global' HA Ingress, meant to be deployed in both clusters
- an 'eu' HA Ingress meant to be deployed in EU cluster
- a 'us' HA Ingress mean to be deployed in US cluster

## Setup

1. Create two Kubernetes clusters, one in EU, one in US
1. Ensure that [regional routing][rr] is enabled for your tailnet
1. Update ACLs:

```
"tagOwners": {
		"tag:k8s-operator":   [],
		"tag:ingresses":      ["tag:k8s-operator"],
		"tag:usingress":      ["tag:k8s-operator"],
		"tag:euingress":      ["tag:k8s-operator"],
		"tag:us":             ["tag:k8s-operator"],
		"tag:eu":             ["tag:k8s-operator"],
		"tag:global":         ["tag:k8s-operator"],
		...
	},
...
"autoApprovers": {
		"services": {
			"tag:k8s":    ["tag:ingresses"],
			"tag:us":     ["tag:usingress"],
			"tag:eu":     ["tag:euingress"],
			"tag:global": ["tag:euingress", "tag:usingress"],
		},
		...
```

### In the US cluster

1. Install the operator

```
$ helm upgrade --install operator tailscale-dev/tailscale-operator \
-n tailscale --create-namespace  --set oauth.clientId=<id> \
--set oauth.clientSecret=<key> \
--set operatorConfig.image.repo=europe-west2-docker.pkg.dev/tailscale-sandbox/irbe-images/operator \
--set operatorConfig.image.tag=v0.0.5multicluster
```
1. Create 'us' ProxyGroup

```
$ kubectl apply -f yamls/pg-us.yaml
```
1. Create both apps
```
$ kubectl apply -f yamls/tails/deploy.yaml
$ kubectl apply -f yamls/scales/deploy.yaml
```
1. Create 'global' Ingresses for each app:
```
$ kubectl apply -f yamls/tails/global.yaml
$ kubectl apply -f yamls/scales/global.yaml
```
1. Create 'us' Ingresses for each app:
```
$ kubectl apply -f yamls/tails/us.yaml
$ kubectl apply -f yamls/scales/us.yaml
```

### In the EU cluster

1. Install the operator

```
$ helm upgrade --install operator tailscale-dev/tailscale-operator \
-n tailscale --create-namespace  --set oauth.clientId=<id> \
--set oauth.clientSecret=<key> \
--set operatorConfig.image.repo=europe-west2-docker.pkg.dev/tailscale-sandbox/irbe-images/operator \
--set operatorConfig.image.tag=v0.0.5multicluster
```
1. Create 'eu' ProxyGroup

```
$ kubectl apply -f yamls/pg-eu.yaml
```
1. Create both apps
```
$ kubectl apply -f yamls/tails/deploy.yaml
$ kubectl apply -f yamls/scales/deploy.yaml
```
1. Create 'global' Ingresses for each app:
```
$ kubectl apply -f yamls/tails/global.yaml
$ kubectl apply -f yamls/scales/global.yaml
```
1. Create 'eu' Ingresses for each app:
```
$ kubectl apply -f yamls/tails/eu.yaml
$ kubectl apply -f yamls/scales/eu.yaml
```

## Test

1. See VIPService MagicDNS names for the newly created Ingresses.

For example, in the 'eu' cluster

```
$ kubectl get ingress
NAME          CLASS       HOSTS   ADDRESS                       PORTS     AGE
eu-scales     tailscale   *       eu-scales1.tailxyz.ts.net   80, 443   3h18m
eu-tails      tailscale   *       eu-tails.tailxyz.ts.net     80, 443   3h18m
scales        tailscale   *       scales.tailxyz.ts.net       80, 443   3h18m
tails         tailscale   *       tails.tailxyz.ts.net        80, 443   3h18m
```
1. Test that the 'global' Ingresses traffic is routed to the nearest region

For example, from a client that is in US:
```
$ curl tails.tailxyz.ts.net
Hello from app tails in us-central1-c
...
$ curl scales.tailxyz.ts.net
Hello from app scales in us-central1-c
```

1. Test that 'regional' Ingress traffic is routed as expected

```
$ curl eu-tails.tailxyz.ts.net
Hello from app tails in europe-west2-a
...
$ curl eu-scales.tailxyz.ts.net
Hello from app scales in europe-west2-a
...
$ curl us-scales.tailxyz.ts.net
Hello from app scales in us-central1-c
...
$ curl us-tails.tailxyz.ts.net
Hello from app tails in us-central1-c
```
[rr]: https://tailscale.com/kb/1115/high-availability#regional-routing

