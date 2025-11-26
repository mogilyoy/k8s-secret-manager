/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/google/uuid"
	"github.com/mogilyoy/k8s-secret-manager/internal/api"
	"github.com/spf13/cobra"
)

var (
	serverURL string
	token     string
)

type Config struct {
	ServerURL string `json:"server_url"`
	Token     string `json:"token"`
}

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "ksec",
	Short: "A util to manage k8s custom secret resourses",
	Long:  `You can create, read, update and delete custom k8s secret resourses that will be delivered to k8s by operator`,

	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			cmd.Help()
			os.Exit(0)
		}
	},
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {

	loadToken()

	defaultServerURL := "https://localhost:8080"
	if serverURL != "" {
		defaultServerURL = serverURL
	}

	rootCmd.PersistentFlags().StringVarP(
		&serverURL,
		"server",
		"s",
		defaultServerURL,
		"Base URL of the K8s Secret Manager API server",
	)

	rootCmd.PersistentFlags().StringVarP(
		&token,
		"token",
		"t",
		token,
		"JWT Bearer token for authentication",
	)

	if err := rootCmd.PersistentFlags().SetAnnotation("server", "env", []string{"KSEC_SERVER_URL"}); err != nil {
		fmt.Fprintln(os.Stderr, "Error setting annotation for server flag:", err)
	}

	if err := rootCmd.PersistentFlags().SetAnnotation("token", "env", []string{"KSEC_TOKEN"}); err != nil {
		fmt.Fprintln(os.Stderr, "Error setting annotation for token flag:", err)
	}

	if serverURL != defaultServerURL {
		rootCmd.PersistentFlags().Set("server", serverURL)
	}
	if token != "" {
		rootCmd.PersistentFlags().Set("token", token)
	}

}

func doAPIRequest(req *http.Request) ([]byte, int, error) {

	requestID := uuid.New().String()
	req.Header.Set("X-Request-ID", requestID)

	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	client := &http.Client{Timeout: 10 * time.Second}

	fmt.Printf(" [Trace] Sending request %s %s with X-Request-ID: %s\n", req.Method, req.URL.Path, requestID)

	resp, err := client.Do(req)
	if err != nil {
		return nil, 0, fmt.Errorf("API request failed (network/timeout): %w", err)
	}
	defer resp.Body.Close()

	if returnedID := resp.Header.Get("X-Request-ID"); returnedID != "" {
		fmt.Printf(" [Trace] Received response with X-Request-ID: %s\n", returnedID)
	}

	responseBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, resp.StatusCode, fmt.Errorf("failed to read response body: %w", err)
	}

	return responseBytes, resp.StatusCode, nil
}

func getAppConfigPath() (string, error) {
	ex, err := os.Executable()
	if err != nil {
		return "", fmt.Errorf("failed to get executable path: %w", err)
	}
	execDir := filepath.Dir(ex)

	configFile := filepath.Join(execDir, "config.json")

	return configFile, nil
}

func saveToken(t string) error {
	configFile, err := getAppConfigPath()
	if err != nil {
		return err
	}

	cfg := Config{Token: t}
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.WriteFile(configFile, data, 0600); err != nil {
		return fmt.Errorf("failed to write config file %s: %w", configFile, err)
	}

	fmt.Printf(" [Config] Token saved successfully to: %s\n", configFile)
	return nil
}

func loadToken() {
	configFile, err := getAppConfigPath()
	if err != nil {
		return
	}

	data, err := os.ReadFile(configFile)
	if err != nil {
		return
	}

	var cfg Config
	if err := json.Unmarshal(data, &cfg); err == nil {

		if cfg.ServerURL != "" {
			serverURL = cfg.ServerURL
			fmt.Printf(" [Config] Server URL loaded: %s\n", serverURL)
		}

		if cfg.Token != "" {
			token = cfg.Token
			fmt.Println(" [Config] Token loaded from file.")
		}
	}
}

func readDataFromFile(filename string) (map[string]string, error) {
	dataBytes, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to read data file %s: %w", filename, err)
	}

	var data map[string]string
	if err := json.Unmarshal(dataBytes, &data); err != nil {
		return nil, fmt.Errorf("failed to unmarshal JSON data from %s: %w", filename, err)
	}

	return data, nil
}

func readGenerationConfigFromFile(filename string) (api.GenerationConfig, error) {
	dataBytes, err := os.ReadFile(filename)
	if err != nil {
		return api.GenerationConfig{}, fmt.Errorf("failed to read config file %s: %w", filename, err)
	}

	var config api.GenerationConfig
	if err := json.Unmarshal(dataBytes, &config); err != nil {
		return api.GenerationConfig{}, fmt.Errorf("failed to unmarshal JSON config from %s: %w", filename, err)
	}

	return config, nil
}
