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
kubectl apply -f deploy/app.yaml
kubectl rollout status deployment.apps/ping -n $SPACE
```

Check pod status: 

```shell
kubectl get pods -n grpc-lab

NAME                    READY   STATUS    RESTARTS   AGE
ping-554f558fbd-h96rf   1/1     Running   0          15s
```

Check server logs:

```shell
kubectl logs -l app=ping -n $SPACE 
```

## service 

```shell
kubectl apply -f deploy/service.yaml
```

```shell
kubectl get service -n $SPACE

NAME   TYPE        CLUSTER-IP   EXTERNAL-IP   PORT(S)     AGE
ping   ClusterIP   10.0.31.13   <none>        50505/TCP   6s
```

## test 

```shell
kubectl port-forward svc/ping 50505 -n $SPACE
```

```shell
grpcurl -plaintext \
        -d '{"id":"id1", "message":"hello"}' \
        -authority=ping.thingz.io \
        localhost:50505 \
        io.thingz.grpc.v1.Service/Ping
```

Response

```json
{
  "id": "5beea6b8-e9b7-4564-8635-e56212884eb9",
  "message": "hello",
  "reversed": "olleh",
  "count": "1"
}
```

## ingress 

TLS Certs secret 

```shell
kubectl create secret tls tls-secret \
		--key certs/ingress-key.pem \
		--cert certs/ingress-cert.pem \
		-n $SPACE 
```

Create ingress

```shell
kubectl apply -f deploy/ingress.yaml
```

```shell
kubectl get ingress -n $SPACE
```

## test via ingress

```shell
grpcurl -d '{"id":"id1", "message":"hello"}' \
  ping.thingz.io:443 \
  io.thingz.grpc.v1.Service/Ping
```

Responds:

```json
{
  "id": "3c55a679-2b6b-49bd-b3d8-73c17f4a2c9b",
  "message": "hello",
  "reversed": "olleh",
  "count": "2"
}
```

## cleanup 

```shell
kubectl delete -f deploy
kubectl delete secret tls-secret -n $SPACE 
```
