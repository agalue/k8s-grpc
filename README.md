# K8s-GRPC

A simple GRPC Example implemented in Go using a Private Certificate Chain for Mutual TLS.

The idea is verify how mTLS works without Kubernetes, and then with Kubernetes terminating mTLS via Nginx Ingress Controller.

## Local Usage

Retrieve the third-party libraries:

```bash
go get -u
```

Generate the Certificates:

```bash
cd pki
./generate.sh en0
cd ..
```

The above assumes the primary IP address of your machine is associated with the interface `en0`. If you're running on Windows, you must modify `generate.sh` to correctly set the `hostname` and IP Address for the Server Certificate. Also, the command `hostname` expects to return the machine's hostname that should resolve to the same IP Address mentioned earlier.

In one terminal, start the server:

```bash
go run server/main.go -port 9000
```

In one another terminal, run the client:

```bash
go run client/main.go -name Rick -server $(hostname):9000
```

If everything went well, you should see on the server-side:

```
2021/10/26 09:35:24 server listening at [::]:9000
2021/10/26 09:36:46 Received: Rick
```

On the client-side:

```
2021/10/26 09:36:46 Greeting: Hello Rick from agalue-mbp.local
```

To stop the server, type `Crtl+C`.

## Docker Usage

Build Docker Image:

```bash=
docker build -t agalue/k8s-grpc-server .
```

Test the dockerized server:

```bash=
docker run -it --rm -p 9000:9000 -v $(pwd)/pki:/pki -h grpc-server agalue/k8s-grpc-server \
  -ca-cert-file=/pki/ca.pem \
  -int-cert-file=/pki/support.pem \
  -srv-cert-file=/pki/server.pem \
  -srv-key-file=/pki/server-key.pem
```

On another terminal, run the client as mentioned before and verify the work.

# Kubernetes Usage

Start Minikube with the Ingress Add-on enabled:

```bash
minikube start --cpus=2 --memory=4g --addons=ingress --addons=ingress-dns
```

Please take a look at the documentation of [ingress-dns](https://github.com/kubernetes/minikube/tree/master/deploy/addons/ingress-dns) for more information about how to use it, to avoid messing with `/etc/hosts`.

For instance, for macOS:

```bash
cat <<EOF | sudo tee /etc/resolver/minikube-default-test
domain test
nameserver $(minikube ip)
search_order 1
timeout 5
EOF
```

Then create the secrets with the certificates for mTLS, the deployment and the ingress:

```bash
cat pki/support.pem pki/ca.pem > pki/ca-chain.pem
cat pki/server.pem pki/ca-chain.pem > pki/server-chain.pem

kubectl create secret tls ingress-cert --key=pki/server-key.pem --cert=pki/server-chain.pem
kubectl create secret generic chain-cert --from-file=ca.crt=pki/ca-chain.pem

kubectl create deployment grpc-server --replicas 2 --port 9000 --image agalue/k8s-grpc-server -- server -tls=false
kubectl expose deployment grpc-server --port 9000

kubectl create ingress grpc-ingress --class=nginx \
  --rule="grpc-server.test/*=grpc-server:9000,tls=ingress-cert" \
  --annotation nginx.ingress.kubernetes.io/ssl-redirect=true \
  --annotation nginx.ingress.kubernetes.io/backend-protocol=GRPC \
  --annotation nginx.ingress.kubernetes.io/auth-tls-verify-client=on \
  --annotation nginx.ingress.kubernetes.io/auth-tls-secret=default/chain-cert \
  --annotation nginx.ingress.kubernetes.io/auth-tls-verify-depth=1 \
  --annotation nginx.ingress.kubernetes.io/auth-tls-pass-certificate-to-upstream=false
```

Finally, verify the commmunication with the client:

```bash
go run client/main.go -name Rick -server grpc-server.test:443
```

And you should see something like this:

```
2021/10/27 07:57:43 Greeting: Hello Rick from grpc-server-6df8fc5db6-fz4sv
```

To destroy the lab:

```bash=
minikube destroy
```
