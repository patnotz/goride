# Seriously, goride

## Generating the SSL cert and private key

Generate the key:

```shell
openssl req -new -newkey rsa:2048 -nodes -keyout goride.key -out goride.csr
```

Self-sign it:

```shell
openssl x509 -req -days 365 -in goride.csr -signkey goride.key -out goride.crt
```

Install the cert in Chrome: chrome://settings/certificates


## TODO List

1. Add CSV export.