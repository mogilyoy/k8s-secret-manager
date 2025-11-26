/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/mogilyoy/k8s-secret-manager/internal/api"
	"github.com/spf13/cobra"
)

// --- Переменные для флагов ---
var deleteNamespace string
var deleteForce bool
var deleteSecretName string

// deleteCmd implements the "ksec delete NAME" command
var deleteCmd = &cobra.Command{
	Use:     "delete NAME",
	Aliases: []string{"del", "rm"},
	Short:   "Delete a SecretClaim resource",
	Long: `Deletes a specific SecretClaim resource by its name and namespace. 
This corresponds to the DELETE /secrets/{name} API endpoint.`,
	Example: `  ksec delete my-old-secret -n staging
  ksec delete another-secret -n dev --force`,
	Args: cobra.ExactArgs(1),
	RunE: runDeleteSecret,
}

func init() {
	rootCmd.AddCommand(deleteCmd)

	deleteCmd.Flags().StringVarP(&deleteNamespace, "namespace", "n", "", "Target Kubernetes namespace (required)")

	deleteCmd.Flags().BoolVarP(&deleteForce, "force", "f", false, "Skip confirmation prompt")

	deleteCmd.MarkFlagRequired("namespace")
}

func runDeleteSecret(cmd *cobra.Command, args []string) error {
	deleteSecretName = args[0]

	if token == "" {
		return fmt.Errorf("authentication token is missing. Please run 'ksec login' first or provide --token flag")
	}

	if !deleteForce {
		fmt.Printf("⚠️ Are you sure you want to delete SecretClaim '%s' in namespace '%s'? (y/N): ", deleteSecretName, deleteNamespace)
		var confirmation string
		fmt.Scanln(&confirmation)

		if confirmation != "y" && confirmation != "Y" {
			fmt.Println("Deletion cancelled.")
			return nil
		}
	}

	deleteURL := fmt.Sprintf("%s/secrets/%s?namespace=%s", serverURL, deleteSecretName, deleteNamespace)

	httpReq, err := http.NewRequest("DELETE", deleteURL, nil)
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

	var successResponse api.OkResponse
	if err := json.Unmarshal(responseBytes, &successResponse); err != nil {
		return fmt.Errorf("failed to decode successful response (Status %d): %w", statusCode, err)
	}

	fmt.Printf("✅ SecretClaim '%s/%s' deleted successfully.\n", deleteNamespace, deleteSecretName)

	return nil
}
