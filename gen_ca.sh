#!/bin/sh

openssl genrsa -out myproxy-ca.key 2048
openssl req -new -x509 -days 3650 -key myproxy-ca.key -out myproxy-ca.crt -subj "/CN=myproxy"
openssl genrsa -out cert.key 2048
mkdir certs/

