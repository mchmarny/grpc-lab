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
kubectl apply -f deployments/app.yaml -n $SPACE
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
kubectl apply -f deployments/service.yaml -n $SPACE
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
	  -authority="ping.thingz.io" \
	  :50505 \
	  io.thingz.grpc.v1.Service/Ping
```

Response

```json
{
  "id": "id1",
  "message": "hello",
  "reversed": "olleh",
  "count": "2",
  "created": "1605994941996399038",
  "metadata": {
    "address": "[::]:50505"
  }
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
kubectl apply -f deployments/ingress.yaml -n $SPACE
```

Watch the ingress until the `ADDRESS` column gets populated:

```shell
kubectl get ingress -n $SPACE -w
```

Now go to the DNS server and create an `A` entry for the value in the `HOSTS` column. You can test when this gets propagated using `dig ping.thingz.io` for example, or whatever the `HOST` value is.

## test via ingress

Now, test it again but this time without the `-plaintext` and `-authority` flags: 

```shell
grpcurl -d '{"id":"id1", "message":"hello"}' \
  ping.thingz.io:443 \
  io.thingz.grpc.v1.Service/Ping
```

If everything goes well you should see the same response we go using port forwarding before. 

## cleanup 

```shell
kubectl delete -f deployments/ -n $SPACE
kubectl delete secret tls-secret -n $SPACE 
kubectl delete ns $SPACE
```
