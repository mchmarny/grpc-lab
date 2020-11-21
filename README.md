# grpc-lab

## setup

```shell
export SPACE="grpc-lab"
```

Namespace:

```shell
kubectl create ns $SPACE
```

## deployment 

Apply deployment:

```shell
kubectl apply -f deploy/server.yaml
```

Check pod status: 

```shell
kubectl get pods -n grpc-lab

NAME                    READY   STATUS    RESTARTS   AGE
grpc-775965b896-lf4zw   1/1     Running   0          6s
```

Check server logs:

```shell
kubectl logs -l app=grpc -n $SPACE 
```

## service 

```shell
kubectl apply -f deploy/service.yaml
```

```shell
kubectl get service -n $SPACE

NAME   TYPE        CLUSTER-IP   EXTERNAL-IP   PORT(S)     AGE
grpc   ClusterIP   10.0.3.233   <none>        50505/TCP   13s
```

## ingress 

TLS Certs secret 

```shell
kubectl create secret tls tls-secret \
		--key certs/server-key.pem \
		--cert certs/server-cert.pem \
		-n $SPACE 
```

Create ingress

```shell
kubectl apply -f deploy/ingress.yaml
```

```shell
kubectl get ingress -n $SPACE
```

## test

```shell
grpcurl -v -insecure grpc.thingz.io:443 list

grpcurl -v \
  -d '{"id":"id1", "message":"hello"}' \
  grpc.thingz.io:443 \
  io.thingz.grpc.v1.Service/Ping
```

## cleanup 

```shell
kubectl delete -f deploy
kubectl delete secret tls-secret -n $SPACE 
```
