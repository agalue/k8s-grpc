#!/bin/bash

echo "### Root CA ..."
openssl x509 -in ca.pem -text -noout

echo "### Intermediate CA ..."
openssl x509 -in support.pem -text -noout

echo "### Server Certificate ..."
openssl x509 -in server.pem -text -noout

echo "### Client Certificate ..."
openssl x509 -in client.pem -text -noout
