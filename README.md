# k8s-scheduling-webhook

A k8s mutation webhook responsible for injecting scheduling related settings to all pods created in namespaces with a specific configurable label.

Especially it can inject:

  - nodeAntiAffinity

## Install

### ssl/tls

the `ssl/` dir contains a script to create a self-signed certificate. Those certificates will configure the deploy/ files.

```
cd ssl/ 
make certs 
```

### deploy

Review the files in `deploy/`

And then:
```
kubectl apply -f deploy/
```


## Development

### build 

```
make
```

### docker

to create a docker image .. 

```
make docker
```

it'll be tagged with the current git commit (short `ref`) and `:latest`

```
make push
```

to push the docker image

