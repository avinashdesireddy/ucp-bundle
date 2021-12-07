# Introduction

To access Mirantis Kubernetes Engine(MKE) using CLI client and kubectl, we need to download and use a UCP client bundle.

This script `ucp-bundle.sh` is a bash script generates and downloads a client bundle.

Manual process is described here - https://docs.mirantis.com/mke/3.5/ops/access-cluster.html

## Syntax
```
$ source ucp-bundle.sh
```

[![asciicast](https://asciinema.org/a/oXwytcI0meYZosvSCF8exgRWE.svg)](https://asciinema.org/a/oXwytcI0meYZosvSCF8exgRWE)


## Configure plugin
```
wget https://github.com/avinashdesireddy/ucp-bundle/blob/master/bin/kubectl-mke?raw=true -O /usr/local/kubectl-mke
```
> Binary is only build for darwin

## Kubectl plugin to download & configure client bundle
```
kubectl mke --ucp-url https://mke.cluster.mirantis.com --ucp-username 'username' --ucp-password 'securepassword'
```

