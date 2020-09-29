# Seriously, goride

## Generating the SSL cert and private key

Generate the key:

'''
openssl req -new -newkey rsa:2048 -nodes -keyout goride.key -out goride.csr
'''

Self-sign it:

'''
openssl x509 -req -days 365 -in goride.csr -signkey goride.key -out goride.crt
'''

Install the cert in Chrome: chrome://settings/certificates


## TODO List

1. Formatted go output
1. Figure out OAuth authentication with Strava
1. Unmarshal JSON data using a struct to select the fields we want. See https://medium.com/rungo/working-with-json-in-go-7e3a37c5a07b
1. Figure out HTML templating
1. Figure out how to serve static image files. See https://medium.com/rungo/beginners-guide-to-serving-files-using-http-servers-in-go-4e542e628eac.
1. Parse the Strava client secret from a local file (not controlled by Git).
