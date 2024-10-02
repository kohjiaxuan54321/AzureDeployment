package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/to"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/resources/armresources"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/storage/armstorage"
	"github.com/joho/godotenv"
)

// Config holds all the configuration variables loaded from the .env file
type Config struct {
	AzureSubscriptionID     string
	AzureLocation           string
	AzureResourceGroupName  string
	AzureStorageAccountName string
	AzureFunctionAppName    string
	FunctionName            string
	FunctionTemplate        string
	AuthLevel               string
	KeepResource            string
}

// Global variables for Azure SDK clients
var (
	resourcesClientFactory *armresources.ClientFactory
	storageClientFactory   *armstorage.ClientFactory
	resourceGroupClient    *armresources.ResourceGroupsClient
	accountsClient         *armstorage.AccountsClient
)

// functionProjectDir defines the directory for your Function App project
const functionProjectDir = `C:\Project\jx\functionapp` // Ensure this path exists

func main() {
	// Step 1: Load environment variables from .env file
	err := godotenv.Load()
	if err != nil {
		log.Println("Error loading .env file:", err)
	} else {
		log.Println(".env file loaded successfully.")
	}

	// Step 2: Load configuration into Config struct
	config := loadConfig()

	// Step 3: Validate required environment variables
	validateConfig(config)

	// Step 4: Validate that required commands are available
	if !isCommandAvailable("az") {
		log.Fatal("'az' command is not available. Please install Azure CLI.")
	}

	if !isCommandAvailable("func") {
		log.Fatal("'func' command is not available. Please install Azure Functions Core Tools.")
	}

	// Step 5: Initialize Azure SDK credentials
	cred, err := azidentity.NewDefaultAzureCredential(nil)
	if err != nil {
		log.Fatalf("Failed to obtain a credential: %v", err)
	}
	ctx := context.Background()

	// Step 6: Initialize Azure SDK clients
	resourcesClientFactory, err = armresources.NewClientFactory(config.AzureSubscriptionID, cred, nil)
	if err != nil {
		log.Fatalf("Failed to create resources client factory: %v", err)
	}
	resourceGroupClient = resourcesClientFactory.NewResourceGroupsClient()

	storageClientFactory, err = armstorage.NewClientFactory(config.AzureSubscriptionID, cred, nil)
	if err != nil {
		log.Fatalf("Failed to create storage client factory: %v", err)
	}
	accountsClient = storageClientFactory.NewAccountsClient()

	// Step 7: Create Resource Group
	resourceGroup, err := createResourceGroup(ctx, config)
	if err != nil {
		log.Fatalf("Failed to create resource group: %v", err)
	}
	log.Println("Resource Group Created:", *resourceGroup.ID)

	// Step 8: Check Storage Account Name Availability
	availability, err := checkNameAvailability(ctx, config)
	if err != nil {
		log.Fatalf("Failed to check storage account name availability: %v", err)
	}
	if !*availability.NameAvailable {
		log.Fatalf("Storage account name is not available: %s", *availability.Message)
	}

	// Step 9: Create Storage Account
	storageAccount, err := createStorageAccount(ctx, config)
	if err != nil {
		log.Fatalf("Failed to create storage account: %v", err)
	}
	log.Println("Storage Account Created:", *storageAccount.ID)

	// Step 10: Get Storage Account Properties
	properties, err := storageAccountProperties(ctx, config)
	if err != nil {
		log.Fatalf("Failed to get storage account properties: %v", err)
	}
	log.Println("Storage Account Properties ID:", *properties.ID)

	// Step 11: Initialize Function App Project (if not already)
	err = initializeFunctionProject()
	if err != nil {
		log.Fatalf("Failed to initialize Function App project: %v", err)
	}
	log.Println("Function App Project Initialized Successfully.")

	// Step 12: Create New Function using `func new`
	err = createNewFunction(config)
	if err != nil {
		log.Fatalf("Failed to create new Function: %v", err)
	}
	log.Println("New Function Created Successfully.")

	// Step 13: Execute Azure CLI Command to Create Function App
	err = createFunctionApp(config)
	if err != nil {
		log.Fatalf("Failed to create Function App: %v", err)
	}
	log.Println("Function App Created Successfully.")

	// Step 14: Publish Function App
	err = publishFunctionApp(config)
	if err != nil {
		log.Fatalf("Failed to publish Function App: %v", err)
	}
	log.Println("Function App Published Successfully.")

	// Step 15: Cleanup Resources if KEEP_RESOURCE is not set
	if !shouldKeepResource(config.KeepResource) {
		err = cleanup(ctx, config)
		if err != nil {
			log.Fatalf("Failed to clean up resources: %v", err)
		}
		log.Println("Resources cleaned up successfully.")
	}
}

// loadConfig retrieves environment variables and populates the Config struct
func loadConfig() Config {
	return Config{
		AzureSubscriptionID:     os.Getenv("AZURE_SUBSCRIPTION_ID"),
		AzureLocation:           os.Getenv("AZURE_LOCATION"),
		AzureResourceGroupName:  os.Getenv("AZURE_RESOURCE_GROUP_NAME"),
		AzureStorageAccountName: os.Getenv("AZURE_STORAGE_ACCOUNT_NAME"),
		AzureFunctionAppName:    os.Getenv("AZURE_FUNCTION_APP_NAME"),
		FunctionName:            os.Getenv("FUNCTION_NAME"),
		FunctionTemplate:        os.Getenv("FUNCTION_TEMPLATE"),
		AuthLevel:               os.Getenv("AUTH_LEVEL"),
		KeepResource:            os.Getenv("KEEP_RESOURCE"),
	}
}

// validateConfig checks that all required environment variables are set
func validateConfig(cfg Config) {
	missingVars := []string{}

	if cfg.AzureSubscriptionID == "" {
		missingVars = append(missingVars, "AZURE_SUBSCRIPTION_ID")
	}
	if cfg.AzureLocation == "" {
		missingVars = append(missingVars, "AZURE_LOCATION")
	}
	if cfg.AzureResourceGroupName == "" {
		missingVars = append(missingVars, "AZURE_RESOURCE_GROUP_NAME")
	}
	if cfg.AzureStorageAccountName == "" {
		missingVars = append(missingVars, "AZURE_STORAGE_ACCOUNT_NAME")
	}
	if cfg.AzureFunctionAppName == "" {
		missingVars = append(missingVars, "AZURE_FUNCTION_APP_NAME")
	}
	if cfg.FunctionName == "" {
		missingVars = append(missingVars, "FUNCTION_NAME")
	}
	if cfg.FunctionTemplate == "" {
		missingVars = append(missingVars, "FUNCTION_TEMPLATE")
	}
	if cfg.AuthLevel == "" {
		missingVars = append(missingVars, "AUTH_LEVEL")
	}

	if len(missingVars) > 0 {
		log.Fatalf("Missing required environment variables: %v", missingVars)
	}

	log.Println("All required environment variables are set.")
}

// isCommandAvailable checks if a command is available in the system's PATH.
func isCommandAvailable(name string) bool {
	_, err := exec.LookPath(name)
	return err == nil
}

// shouldKeepResource determines whether to keep Azure resources based on KEEP_RESOURCE value
func shouldKeepResource(keep string) bool {
	switch keep {
	case "1", "true", "True", "TRUE":
		return true
	default:
		return false
	}
}

// createResourceGroup creates an Azure Resource Group
func createResourceGroup(ctx context.Context, cfg Config) (*armresources.ResourceGroup, error) {
	resourceGroupResp, err := resourceGroupClient.CreateOrUpdate(
		ctx,
		cfg.AzureResourceGroupName,
		armresources.ResourceGroup{
			Location: to.Ptr(cfg.AzureLocation),
		},
		nil,
	)
	if err != nil {
		return nil, err
	}
	return &resourceGroupResp.ResourceGroup, nil
}

// checkNameAvailability checks if the storage account name is available
func checkNameAvailability(ctx context.Context, cfg Config) (*armstorage.CheckNameAvailabilityResult, error) {
	result, err := accountsClient.CheckNameAvailability(
		ctx,
		armstorage.AccountCheckNameAvailabilityParameters{
			Name: to.Ptr(cfg.AzureStorageAccountName),
			Type: to.Ptr("Microsoft.Storage/storageAccounts"),
		},
		nil,
	)
	if err != nil {
		return nil, err
	}
	return &result.CheckNameAvailabilityResult, nil
}

// createStorageAccount creates an Azure Storage Account
func createStorageAccount(ctx context.Context, cfg Config) (*armstorage.Account, error) {
	pollerResp, err := accountsClient.BeginCreate(
		ctx,
		cfg.AzureResourceGroupName,
		cfg.AzureStorageAccountName,
		armstorage.AccountCreateParameters{
			Kind:     to.Ptr(armstorage.KindStorageV2),
			SKU:      &armstorage.SKU{Name: to.Ptr(armstorage.SKUNameStandardLRS)},
			Location: to.Ptr(cfg.AzureLocation),
			Properties: &armstorage.AccountPropertiesCreateParameters{
				AccessTier: to.Ptr(armstorage.AccessTierCool),
				Encryption: &armstorage.Encryption{
					Services: &armstorage.EncryptionServices{
						File:  &armstorage.EncryptionService{KeyType: to.Ptr(armstorage.KeyTypeAccount), Enabled: to.Ptr(true)},
						Blob:  &armstorage.EncryptionService{KeyType: to.Ptr(armstorage.KeyTypeAccount), Enabled: to.Ptr(true)},
						Queue: &armstorage.EncryptionService{KeyType: to.Ptr(armstorage.KeyTypeAccount), Enabled: to.Ptr(true)},
						Table: &armstorage.EncryptionService{KeyType: to.Ptr(armstorage.KeyTypeAccount), Enabled: to.Ptr(true)},
					},
					KeySource: to.Ptr(armstorage.KeySourceMicrosoftStorage),
				},
			},
		}, nil)
	if err != nil {
		return nil, err
	}
	resp, err := pollerResp.PollUntilDone(ctx, nil)
	if err != nil {
		return nil, err
	}
	return &resp.Account, nil
}

// storageAccountProperties retrieves properties of the Storage Account
func storageAccountProperties(ctx context.Context, cfg Config) (*armstorage.Account, error) {
	storageAccountResponse, err := accountsClient.GetProperties(
		ctx,
		cfg.AzureResourceGroupName,
		cfg.AzureStorageAccountName,
		nil,
	)
	if err != nil {
		return nil, err
	}
	return &storageAccountResponse.Account, nil
}

// initializeFunctionProject initializes a new Azure Functions project if not already initialized
func initializeFunctionProject() error {
	// Check if the project directory exists
	if _, err := os.Stat(functionProjectDir); os.IsNotExist(err) {
		// Create the project directory
		err := os.MkdirAll(functionProjectDir, os.ModePerm)
		if err != nil {
			return fmt.Errorf("failed to create project directory: %v", err)
		}
	}

	// Change to the project directory
	err := os.Chdir(functionProjectDir)
	if err != nil {
		return fmt.Errorf("failed to change directory to project directory: %v", err)
	}

	// Initialize a new Functions project with Node.js runtime
	// This step is optional if your project is already initialized
	cmd := exec.Command("func", "init", "--worker-runtime", "node")
	cmd.Env = os.Environ()
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("func init failed: %v\nOutput: %s", err, string(output))
	}

	log.Printf("func init output:\n%s\n", string(output))
	return nil
}

// createNewFunction creates a new Azure Function using `func new`
func createNewFunction(cfg Config) error {
	// Ensure we are in the project directory
	currentDir, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get current directory: %v", err)
	}
	log.Println("Current Directory:", currentDir)

	// Define the arguments for `func new`
	cmdArgs := []string{
		"new",
		"--name", cfg.FunctionName,
		"--template", cfg.FunctionTemplate,
		"--authlevel", cfg.AuthLevel,
	}

	cmd := exec.Command("func", cmdArgs...)

	// Set environment variables if needed
	cmd.Env = os.Environ()

	// Capture standard output and error
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("func new failed: %v\nOutput: %s", err, string(output))
	}

	log.Printf("func new output:\n%s\n", string(output))
	return nil
}

// createFunctionApp creates an Azure Function App using `az functionapp create`
func createFunctionApp(cfg Config) error {
	cmdArgs := []string{
		"functionapp", "create",
		"--resource-group", cfg.AzureResourceGroupName,
		"--consumption-plan-location", cfg.AzureLocation,
		"--runtime", "node",
		"--runtime-version", "18",
		"--functions-version", "4",
		"--name", cfg.AzureFunctionAppName,
		"--storage-account", cfg.AzureStorageAccountName,
	}

	cmd := exec.Command("az", cmdArgs...)

	// Set environment variables if needed (e.g., AZURE_SUBSCRIPTION_ID)
	cmd.Env = os.Environ()

	// Capture standard output and error
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("az functionapp create failed: %v\nOutput: %s", err, string(output))
	}

	log.Printf("az functionapp create output:\n%s\n", string(output))
	return nil
}

// publishFunctionApp publishes the Function App using `func azure functionapp publish`
func publishFunctionApp(cfg Config) error {
	// Ensure you are in the Function App project directory
	err := os.Chdir(functionProjectDir)
	if err != nil {
		return fmt.Errorf("failed to change directory to project directory: %v", err)
	}

	cmdArgs := []string{
		"azure", "functionapp", "publish", cfg.AzureFunctionAppName,
	}

	cmd := exec.Command("func", cmdArgs...)

	// Set environment variables if needed
	cmd.Env = os.Environ()

	// Capture standard output and error
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("func azure functionapp publish failed: %v\nOutput: %s", err, string(output))
	}

	log.Printf("func azure functionapp publish output:\n%s\n", string(output))
	return nil
}

// cleanup deletes the Resource Group to clean up resources
func cleanup(ctx context.Context, cfg Config) error {
	pollerResp, err := resourceGroupClient.BeginDelete(ctx, cfg.AzureResourceGroupName, nil)
	if err != nil {
		return err
	}

	_, err = pollerResp.PollUntilDone(ctx, nil)
	if err != nil {
		return err
	}
	return nil
}
