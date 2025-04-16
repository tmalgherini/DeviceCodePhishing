# DeviceCodePhishing
DeviceCodePhishing is an advanced phishing tool.
It can be used for phishing access-tokens, which in turn allows to bypass 2-factor authentication protection.

This attack technique is even more dangerous than attacker-in-the-middle (AitM) proxies, because the 
user will **enter credentials on the original webpage**, and this technique is able to also **bypass phishing-resistant FIDO** credentials!
Maybe the user even does not need to enter credentials, because a session already exists. 

There are already other tools which automate device code phishing attacks with certain limitations 
(i.e. once the code is generated, the attacker has only 10 minutes time frame to convince the user to open the url and enter the code).
The goal of this tool is to also automate this once the victim clicked on a phishing link with a headless browser.

Currently, this tool can only be used to target Microsoft Azure Entra Users. But the technique is not limited to a certain vendor.

More Details can be found on that [Blog Post]()

## How it works
1. Attacker sends an url to the victim
2. victim opens url
3. At the moment where the url is opened, this tool starts a headless browser which does the following automated steps:
   - starts Device Code Flow with `<tenant>` and `<clientId>` 
   - opens device-code webpage and enters corresponding user-code
   - device-code webpage forwards user to the url for interactive authentication
   - returning url for interactive authentication as redirect to the victim
4. victim is redirected to the authentication url
5. victim completes the authentication
6. attacker is authenticated


## Build
```shell
docker build . -t device-code-phishing
```

## Run
By default, it runs with tenant set to `common` and with the AuthenticationBroker ClientId `29d9ed98-a469-4536-ade2-f981bc1d605e`
```shell
docker run -p 8080:8080 device-code-phishing
```

Use the args if one want to define a specific tenant or a different clientId
```shell
docker run -p 8080:8080 device-code-phishing --tenant <tenantId> --client-id <clientId> --verbose
```

## Use
Open Url: 
http://localhost:8080/lure

## Disclaimer
Provided as educational content only!