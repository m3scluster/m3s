# Enable TLS for MariaDB

To secure the datastore communitaction with TLS, we have to do several points. 

1. Create the following TLS certificate files:
    - root-ca.pem
    - server-cert.pem
    - server-key.pem

2. Save these files into the root directory of the persistant storage of MariaDB and the K3s server. 

    The storage source will be configured by the following environment variables:
    - VOLUME_K3S_SERVER
    - VOLUME_DS

3. Enable datastore SSL support with `DS_MYSQL_SSL=true`.
