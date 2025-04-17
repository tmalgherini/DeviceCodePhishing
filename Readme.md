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

More Details can be found on that [Blog Post](https://denniskniep.github.io/posts/09-device-code-phishing)

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



## Run with Docker
By default, it runs with tenant set to `common` and with the AuthenticationBroker ClientId `29d9ed98-a469-4536-ade2-f981bc1d605e`
```shell
docker run -p 8080:8080 ghcr.io/denniskniep/device-code-phishing:v1.0.0
```

Use the args if one want to define a specific tenant, a different clientId or a custom userAgent
```shell
docker run -p 8080:8080 ghcr.io/denniskniep/device-code-phishing:v1.0.0 --tenant <tenantId> --client-id <clientId> --user-agent <userAgent> --verbose
```

## Use
Open Url: 
http://localhost:8080/lure

## Build it yourself 
```shell
go build main.go
```

```shell
./main server
```


## Build & Run it yourself with Docker
```shell
docker build . -t device-code-phishing
```

```shell
docker run -p 8080:8080 device-code-phishing
```

## Entra ClientIds

| ClientId                             | Description                     |
|--------------------------------------|---------------------------------|
| 29d9ed98-a469-4536-ade2-f981bc1d605e | Microsoft Authentication Broker |
| 9ba1a5c7-f17a-4de9-a1f1-6178c8d51223 | Microsoft Intune Company Portal |

Hint: Use Microsoft Intune Company Portal for bypassing Intune compliant device Conditional Access Policy ([More Details](https://i.blackhat.com/EU-24/Presentations/EU-24-Chudo-Unveiling-the-Power-of-Intune-Leveraging-Intune-for-Breaking-Into-Your-Cloud-and-On-Premise.pdf))

## Disclaimer
Provided as educational content only!