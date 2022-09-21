#! /bin/bash

NAME="m3sdatastore"
DOMAIN="mini"
HOSTS="server client"

cat > config.cnf <<EOF
[ req ]
prompt = no
default_bits = 4096
default_md = sha256
distinguished_name = req_distinguished_name
x509_extensions = v3_ca
req_extensions = v3_req

[ v3_ca ]
subjectKeyIdentifier=hash
authorityKeyIdentifier=keyid:always,issuer
basicConstraints = CA:true

[req_distinguished_name]
countryName = DE
stateOrProvinceName = SH
localityName = EN
organizationalUnitName = IT

[ v3_req ]
# Extensions to add to a certificate request
basicConstraints = CA:FALSE
keyUsage = nonRepudiation, digitalSignature, keyEncipherment
subjectAltName = @alt_names

[alt_names]
DNS.1 = ${NAME}
DNS.2 = *.${DOMAIN}
DNS.3 = ${NAME}.${DOMAIN}
EOF

openssl genrsa 2048 > ca-key.pem

openssl req -sha256 -new -x509 -nodes -days 3600 -key ca-key.pem -out ca-cert.pem -subj "/CN=${NAME}.${DOMAIN}" -config ./config.cnf


for HOST in $HOSTS
do
	SUBJ="$SUBJ_BASE/CN=$HOST.$DOMAIN"

	# Create server certificate, remove passphrase, and sign it
	# Create the server key
  openssl req -sha256 -newkey rsa:2048 -days 3600 -nodes -keyout $HOST-key.pem -out $HOST-req.pem -config <( cat config.cnf )

	# Process the server RSA key
	openssl rsa -in $HOST-key.pem -out $HOST-key.pem

	# Sign the server certificate
  openssl x509 -sha256 -req -in $HOST-req.pem -days 3600 -CA ca-cert.pem -CAkey ca-key.pem -set_serial 01 -out $HOST-cert.pem

	# Verify certificates
	openssl verify -CAfile ca-cert.pem $HOST-cert.pem
done
