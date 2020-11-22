# grpc-lab

Simple gRPC service with HTTP extension based on a proto API definition and deployment to Kubernetes. 

## setup

Export a couple of environment variables to make this readme reproducible:

> Choose domain for which you can control. When setting up ingress, you will have to create DNS entry. 

```shell
export DOMAIN="thingz.io"
export SPACE="grpc-lab"
```

Create namespace:

```shell
kubectl create ns $SPACE
```

## deployment 

> All commands assume you are executing from the root of the repo

Apply deployment:

```shell
kubectl apply -f deployments/app.yaml -n $SPACE
kubectl rollout status deployment.apps/ping -n $SPACE
```

Check pod status: 

```shell
kubectl get pods -n $SPACE

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


## ingress 

Create TLS certificates:

```shell
certbot certonly --manual \
                 --preferred-challenges dns \
                 -d "*.${DOMAIN}" \
                 --config-dir ./certs \
                 --logs-dir ./certs \
                 --work-dir ./certs
```

> When promoted to configure DNS TXT record, you can check on the propagation status by digging `dig "_acme-challenge.${DOMAIN}" TXT`

Once the certs were generated, create the TLS secret:

```shell
kubectl create secret tls tls-secret \
		--key "live/${DOMAIN}/privkey.pem" \
		--cert "live/${DOMAIN}/fullchain.pem" \
		-n $SPACE 
```

Create gRPC and HTTP ingresses

```shell
kubectl apply -f deployments/grpc-ingress.yaml -n $SPACE
kubectl apply -f deployments/http-ingress.yaml -n $SPACE
```

Watch the ingresses until the `ADDRESS` column gets populated:

```shell
kubectl get ingress -n $SPACE -w
```

Now go to the DNS server and create an `A` entry for the `*` and the IP value in the `HOSTS` column. You can test when this gets propagated using `dig gping.thingz.io` for example, or whatever the `HOST` value is.

## test

Once the DNS is set up, test the gRPC endpoint:

```shell
grpcurl -d '{"id":"id1", "message":"hello"}' \
  gping.thingz.io:443 \
  io.thingz.grpc.v1.Service/Ping
```

Response should look something like this:

```json
{
  "id": "id1",
  "message": "hello",
  "reversed": "olleh",
  "count": "1",
  "created": "1606055127674975938",
  "metadata": {
    "address": "[::]:50505"
  }
}
```

Now, test the HTTP endpoint: 

```shell
curl -i -d '{"id":"id1", "message":"hello"}'\
      -H "Content-type: application/json" \
      https://ping.thingz.io:443/v1/ping
```

Again, the response should look something like this:

```json
HTTP/2 200
date: Sun, 22 Nov 2020 14:27:20 GMT
content-type: application/json
content-length: 129
grpc-metadata-content-type: application/grpc
strict-transport-security: max-age=15724800; includeSubDomains

{
  "id":"id1",
  "message":"hello",
  "reversed":"olleh",
  "count":"4",
  "created":"1606055240599449988",
  "metadata":{
    "address":"[::]:50505"
  }
}
```

## cleanup 

```shell
kubectl delete -f deployments/ -n $SPACE
kubectl delete secret tls-secret -n $SPACE 
kubectl delete ns $SPACE
```
