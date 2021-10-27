#!/bin/bash

INTF=${1-en0}

echo "Generating Root CA"
cat <<EOF | cfssl gencert -initca - | cfssljson -bare ca -
{
  "CN": "Root CA",
  "key": {
    "algo": "rsa",
    "size": 2048
  },
  "names": [
    {
      "C": "US",
      "ST": "NC",
      "O": "Hackers, Inc."
    }
  ]
}
EOF

echo
echo "Generating Intermediate CA"
cat <<EOF | cfssl genkey -initca - | cfssljson -bare support
{
  "CN": "IT/Support (Intermediate)",
  "key": {
    "algo": "rsa",
    "size": 2048
  },
  "names": [
    {
      "C": "US",
      "ST": "NC",
      "O": "Hackers, Inc.",
      "OU": "Support"
    }
  ]
}
EOF
cfssl sign -ca ca.pem -ca-key ca-key.pem --config config.json -profile authority support.csr | cfssljson -bare support

echo
echo "Generating Server Certificate"
cat <<EOF | cfssl gencert -ca=support.pem -ca-key=support-key.pem -config=config.json -profile=server - | cfssljson -bare server
{
  "CN": "$(hostname)",
  "hosts": [
    "$(hostname)",
    "$(ifconfig $INTF | grep 'inet[^6]' | awk '{print $2}')",
    "grpc-server",
    "grpc-server.test"
  ],
  "key": {
    "algo": "rsa",
    "size": 2048
  },
  "names": [
    {
      "C": "US",
      "ST": "NC",
      "O": "Hackers, Inc."
    }
  ]
}
EOF

echo
echo "Generating Client Certificate"
cat <<EOF | cfssl gencert -ca=support.pem -ca-key=support-key.pem -config=config.json -profile=client - | cfssljson -bare client
{
  "CN": "client",
  "hosts": [""],
  "key": {
    "algo": "rsa",
    "size": 2048
  },
  "names": [
    {
      "C": "US",
      "ST": "NC",
      "O": "Hackers, Inc."
    }
  ]
}
EOF

echo
echo "Done!"