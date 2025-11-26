/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/mogilyoy/k8s-secret-manager/internal/api"
	"github.com/spf13/cobra"
)

// listCmd represents the list command
var listNamespace string

// listCmd implements the "ksec list" command
var listCmd = &cobra.Command{
	Use:     "list",
	Aliases: []string{"ls"},
	Short:   "List available SecretClaim resources in a namespace",
	Long: `Lists all SecretClaim resources available in the specified namespace. 
This corresponds to the GET /secrets API endpoint.`,
	Example: `  ksec list -n staging`,
	RunE:    runListSecrets,
}

func init() {
	rootCmd.AddCommand(listCmd)

	listCmd.Flags().StringVarP(&listNamespace, "namespace", "n", "", "Target Kubernetes namespace (required)")

	listCmd.MarkFlagRequired("namespace")
}

func runListSecrets(cmd *cobra.Command, args []string) error {
	if token == "" {
		return fmt.Errorf("authentication token is missing. Please run 'ksec login' first")
	}
	listURL := fmt.Sprintf("%s/secrets?namespace=%s", serverURL, listNamespace)

	httpReq, err := http.NewRequest("GET", listURL, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	responseBytes, statusCode, err := doAPIRequest(httpReq)
	if err != nil {
		return err
	}

	if statusCode != http.StatusOK {
		var errResp ErrorResponse
		if json.Unmarshal(responseBytes, &errResp) == nil {
			return fmt.Errorf("API call failed (Status: %d, Code: %s): %s", errResp.StatusCode, errResp.ErrorCode, errResp.ErrorMessage)
		}
		return fmt.Errorf("API call failed with unexpected status: %d %s", statusCode, http.StatusText(statusCode))
	}

	var successResponse api.ListSecretsResponse
	if err := json.Unmarshal(responseBytes, &successResponse); err != nil {
		return fmt.Errorf("failed to decode successful response (Status %d): %w", statusCode, err)
	}

	if len(successResponse.Items) == 0 {
		fmt.Printf("No SecretClaims found in namespace '%s'.\n", listNamespace)
		return nil
	}

	fmt.Printf("Found %d secrets in namespace '%s':\n", len(successResponse.Items), listNamespace)
	printSecretSummary(successResponse.Items)

	return nil
}

func printSecretSummary(secrets []api.SecretSummary) {
	fmt.Printf("\n%-30s %-15s %-15s %s\n", "NAME", "TYPE", "STATUS", "CREATED")
	fmt.Println("---------------------------------------------------------------------------------")

	for _, s := range secrets {
		displayStatus := string(s.Status.CurrentStatus)
		synced := s.Status.Synced
		if *synced {
			displayStatus = fmt.Sprintf("%s (Synced)", displayStatus)
		} else {
			displayStatus = fmt.Sprintf("%s (Pending)", displayStatus)
		}

		created := "N/A"
		if t, err := time.Parse(time.RFC3339, s.CreationTimestamp.String()); err == nil {
			created = t.Format("2006-01-02 15:04:05")
		}

		fmt.Printf("%-30s %-15s %-15s %s\n", s.Name, s.Type, displayStatus, created)
	}
}
