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

## Install the Strava client secret:

1. Copy the client secret from your Strava API settings at https://www.strava.com/settings/api

1. Save this code in a file named `strava_client_secret.txt` in this directory.

## TODO List

1. Add CSV export.