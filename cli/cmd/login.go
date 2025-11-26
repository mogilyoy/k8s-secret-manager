/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"syscall"

	"github.com/mogilyoy/k8s-secret-manager/internal/api"
	"github.com/spf13/cobra"
	"golang.org/x/term"
)

// Переменные для флагов loginCmd
var (
	loginUsername string
	loginPassword string
)

// ErrorBadRequest (используется для обработки ошибок)
type ErrorResponse struct {
	StatusCode   int    `json:"statusCode"`
	ErrorCode    string `json:"errorCode"`
	ErrorMessage string `json:"errorMessage"`
}

var loginCmd = &cobra.Command{
	Use:   "login",
	Short: "Authenticate and retrieve a JWT token",
	Long: `Login attempts to authenticate the user using the provided username and password 
and stores the resulting JWT token for subsequent API calls. If the password is not 
provided via a flag, the CLI will securely prompt for it.`,
	Example: `
  ksec login -u admin 
  > password:

  # Pass password directly (less secure)
  ksec login -u admin -p secure_pass`,
	RunE: runLogin,
}

func init() {
	rootCmd.AddCommand(loginCmd)
	loginCmd.Flags().StringVarP(&loginUsername, "username", "u", "", "Your API username")
	loginCmd.Flags().StringVarP(&loginPassword, "password", "p", "", "Your API password (passed directly)")
	loginCmd.MarkFlagRequired("username")
}

func runLogin(cmd *cobra.Command, args []string) error {
	if loginPassword == "" {
		fmt.Printf("Enter password for %s: ", loginUsername)

		bytePassword, err := term.ReadPassword(int(syscall.Stdin))
		if err != nil {
			return fmt.Errorf("error reading password: %w", err)
		}

		loginPassword = string(bytePassword)
		fmt.Println()

		if loginPassword == "" {
			return fmt.Errorf("password cannot be empty")
		}
	}

	authReq := api.AuthUserRequest{
		Username: loginUsername,
		Password: loginPassword,
	}

	reqBody, err := json.Marshal(authReq)
	if err != nil {
		return fmt.Errorf("failed to marshal request body: %w", err)
	}

	authURL := fmt.Sprintf("%s/user/auth", serverURL)

	fmt.Printf("Attempting to log in to %s...\n", authURL)

	req, err := http.NewRequest("POST", authURL, bytes.NewBuffer(reqBody))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	responseBytes, statusCode, err := doAPIRequest(req)
	if err != nil {
		return err
	}

	if statusCode != http.StatusOK && statusCode != http.StatusCreated {
		var errResp ErrorResponse
		if json.Unmarshal(responseBytes, &errResp) == nil {
			return fmt.Errorf("API call failed (Status: %d, Code: %s): %s", errResp.StatusCode, errResp.ErrorCode, errResp.ErrorMessage)
		}
		return fmt.Errorf("API call failed with unexpected status: %d %s", statusCode, http.StatusText(statusCode))
	}

	var successResponse api.AuthUserResponse
	if err := json.Unmarshal(responseBytes, &successResponse); err != nil {
		return fmt.Errorf("failed to decode successful response (Status %d): %w", statusCode, err)
	}

	token = successResponse.Token
	if err := saveToken(token); err != nil {
		fmt.Printf("⚠️ Warning: Failed to save token to disk: %s\n", err)
	}
	fmt.Println("✅ Login successful!")
	fmt.Printf("Token received and stored (Expires in: %d seconds).\n", successResponse.ExpiresIn)

	return nil
}
