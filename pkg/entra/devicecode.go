package entra

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/chromedp/cdproto/cdp"
	"github.com/chromedp/chromedp"
	"log/slog"
	"net/http"
	"net/url"
	"strings"
	//"io"
)

type DeviceAuth struct {
	DeviceCode      string `json:"device_code"`
	UserCode        string `json:"user_code"`
	VerificationUri string `json:"verification_uri"`
	ExpiredIn       int    `json:"expires_in"`
	Interval        int    `json:"interval"`
}

type DeviceAuthV1 struct {
	DeviceCode      string `json:"device_code"`
	UserCode        string `json:"user_code"`
	VerificationUri string `json:"verification_url"`
	ExpiredIn       string `json:"expires_in"`
	Interval        string `json:"interval"`
}

type AuthenticationResult struct {
	TokenType    string `json:"token_type"`
	Scope        string `json:"scope"`
	AccessToken  string `json:"access_token"`
	IdToken      string `json:"id_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int    `json:"expires_in"`
}

type AuthenticationResultV1 struct {
	TokenType    string `json:"token_type"`
	Scope        string `json:"scope"`
	AccessToken  string `json:"access_token"`
	IdToken      string `json:"id_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    string `json:"expires_in"`
}

type AuthenticationError struct {
	Type        string `json:"error"`
	Description string `json:"error_description"`
}

const (
	PENDING string = "authorization_pending"
)

// https://learn.microsoft.com/en-us/entra/identity-platform/v2-oauth2-device-code
func RequestDeviceAuth(tenant string, clientId string, scopes []string) (*DeviceAuth, error) {
	resp, err := http.PostForm("https://login.microsoftonline.com/"+tenant+"/oauth2/v2.0/devicecode",
		url.Values{"client_id": {clientId}, "scope": {strings.Join(scopes, " ")}})

	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		errMsg := "Request failed with status code:" + resp.Status
		return nil, errors.New(errMsg)
	}

	var deviceAuth DeviceAuth
	err = json.NewDecoder(resp.Body).Decode(&deviceAuth)

	if err != nil {
		return nil, err
	}
	return &deviceAuth, nil
}

func RequestDeviceAuthV1(tenant string, clientId string, resource string) (*DeviceAuthV1, error){
	resp, err := http.PostForm("https://login.microsoftonline.com/"+tenant+"/oauth2/devicecode?api-version=1.0", 
		url.Values{"client_id": {clientId}, "resource=":{resource}})

	if err != nil {
		return nil, err
	}
	//this breaks next code, use only to debug
	//defer resp.Body.Close()
	//bodyBytes, err := io.ReadAll(resp.Body)
	//bodystring := string(bodyBytes)
	//println(bodystring)
	if resp.StatusCode != http.StatusOK {
		errMsg := "Request failed with status code:" + resp.Status
		return nil, errors.New(errMsg)
	}

	var deviceAuth DeviceAuthV1
	err = json.NewDecoder(resp.Body).Decode(&deviceAuth)

	if err != nil {
		println(err.Error())
		return nil, err
	}
	println(deviceAuth.DeviceCode)
	return &deviceAuth, nil	
}

func RequestToken(tenant string, clientId string, deviceAuth *DeviceAuth) (*AuthenticationResult, error) {
	resp, err := http.PostForm("https://login.microsoftonline.com/"+tenant+"/oauth2/v2.0/token",
		url.Values{"grant_type": {"urn:ietf:params:oauth:grant-type:device_code"}, "client_id": {clientId}, "device_code": {deviceAuth.DeviceCode}})

	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		if resp.StatusCode == http.StatusBadRequest {
			var authErr AuthenticationError
			err = json.NewDecoder(resp.Body).Decode(&authErr)
			if err != nil {
				return nil, err
			}

			if authErr.Type == PENDING {
				return nil, nil
			} else if authErr.Type != "" {
				return nil, errors.New("Polling of device_code concluded with " + authErr.Type)
			}
		}

		errMsg := "Request failed with status code:" + resp.Status
		return nil, errors.New(errMsg)
	}

	var authResult AuthenticationResult
	err = json.NewDecoder(resp.Body).Decode(&authResult)

	if err != nil {
		return nil, err
	}

	return &authResult, nil
}

func RequestTokenV1(tenant string, clientId string, deviceAuth *DeviceAuthV1) (*AuthenticationResultV1, error) {
	resp, err := http.PostForm("https://login.microsoftonline.com/"+tenant+"/oauth2/token?api-version=1.0",
		url.Values{"grant_type": {"urn:ietf:params:oauth:grant-type:device_code"}, "client_id": {clientId}, "code": {deviceAuth.DeviceCode}})

	if err != nil {
		return nil, err
	}
	//this breaks next code, use only to debug
	//defer resp.Body.Close()
	//bodyBytes, err := io.ReadAll(resp.Body)
	//bodystring := string(bodyBytes)
	//println(bodystring)
	if resp.StatusCode != http.StatusOK {
		if resp.StatusCode == http.StatusBadRequest {
			var authErr AuthenticationError
			err = json.NewDecoder(resp.Body).Decode(&authErr)
			if err != nil {
				return nil, err
			}

			if authErr.Type == PENDING {
				return nil, nil
			} else if authErr.Type != "" {
				return nil, errors.New("Polling of device_code concluded with " + authErr.Type + ", error description: " + authErr.Description)
			}
		}

		errMsg := "Request failed with status code:" + resp.Status
		return nil, errors.New(errMsg)
	}

	var authResult AuthenticationResultV1
	err = json.NewDecoder(resp.Body).Decode(&authResult)

	if err != nil {
		return nil, err
	}

	return &authResult, nil
}

func EnterDeviceCodeWithHeadlessBrowser(deviceAuth *DeviceAuthV1, userAgent string) (string, error) {
	allocatorOpts := chromedp.DefaultExecAllocatorOptions[:]
	allocatorOpts = append(allocatorOpts, chromedp.Flag("headless", true))
	allocatorOpts = append(allocatorOpts, chromedp.UserAgent(userAgent))
	ctx, cancel := chromedp.NewExecAllocator(context.Background(), allocatorOpts...)

	var contextOpts []chromedp.ContextOption
	contextOpts = append(contextOpts, chromedp.WithDebugf(slog.Debug))
	ctx, cancel = chromedp.NewContext(ctx, contextOpts...)

	defer cancel()

	var finalUrl string
	var aadTitleHint []*cdp.Node
	err := chromedp.Run(ctx,
		chromedp.Navigate(`https://microsoft.com/devicelogin`),

		chromedp.WaitVisible(`#idSIButton9`),
		chromedp.SendKeys(`#otc`, deviceAuth.UserCode),
		chromedp.Click(`#idSIButton9`),

		chromedp.WaitVisible(`#cantAccessAccount`),
		chromedp.Click(`#cantAccessAccount`),

		chromedp.WaitVisible(`#aadTitleHint, #ContentPlaceholderMainContent_ButtonCancel`),
		chromedp.Nodes(`aadTitleHint`, &aadTitleHint, chromedp.AtLeast(0)),
	)

	if err != nil {
		return "", err
	}

	if len(aadTitleHint) > 0 {
		err := chromedp.Run(ctx,
			chromedp.WaitVisible(`#aadTitleHint`),
			chromedp.Click(`#aadTitleHint`),
		)

		if err != nil {
			return "", err
		}
	}

	err = chromedp.Run(ctx,
		chromedp.WaitVisible(`#ContentPlaceholderMainContent_ButtonCancel`),
		chromedp.Click(`#ContentPlaceholderMainContent_ButtonCancel`),

		chromedp.WaitVisible(`#cantAccessAccount`),
		chromedp.Location(&finalUrl),
	)

	if err != nil {
		return "", err
	}

	return finalUrl, nil
}
