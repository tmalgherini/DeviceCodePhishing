# DeviceCodePhishing
DeviceCodePhishing is an advanced phishing tool.
It can be used for phishing access-tokens, which in turn allows to bypass two-factor authentication protection, including accounts that exclusively use FIDO for authentication.

This attack technique is even more dangerous than attacker-in-the-middle (AitM) proxies, because the 
user **enters their credentials on the original webpage**, making it nearly impossible to detect the phishing attempt based on a suspicious URL.
Additionally, this technique can **bypass phishing-resistant FIDO** credentials!
In some cases, the user may not even need to enter credentials if a session is already active.

While other tools exist to automate device code phishing attacks, they often come with certain limitations, 
such as requiring the attacker to convince the victim to open the URL and enter the code within a strict 10-minute time frame.
The goal of this tool is to overcomes those limitations by automating the process with a headless browser, which initiates the attack 
as soon as the victim clicks on the phishing link.

Currently, this tool is limited to targeting Microsoft Azure Entra users, but the underlying technique is not restricted to any specific vendor.

For more details, check out the blog post: [Phishing despite FIDO, leveraging the device code flow](https://denniskniep.github.io/posts/09-device-code-phishing)

## How it works
1. The attacker sends a URL to the victim
2. The victim opens that URL
3. When the URL is opened, a headless browser is started, performing the following automated steps:
   - Starts the Device Code Flow with `<tenant>` and `<clientId>`
   - Opens the device-code webpage and enters the corresponding user-code
   - The device-code webpage forwards the user to the URL for interactive authentication
   - Returns the URL for interactive authentication as a redirect to the victim
4. The victim is redirected to the authentication URL
5. The victim completes the authentication
6. The attacker is authenticated



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

## Next steps with obtained tokens
Once you have obtained tokens successfully, you can use them with other attack tools like:
* https://github.com/dafthack/GraphRunner
* https://github.com/f-bader/TokenTacticsV2?tab=readme-ov-file#azure-json-web-token-jwt-manipulation-toolset
* https://github.com/secureworks/family-of-client-ids-research

## Disclaimer
Provided as educational content only!