# Azure Functions Go Project

**Azure Functions Go Project** is a Go-based application designed to automate the creation and management of Azure resources, including Azure Function Apps. This project leverages the Azure SDK for Go to streamline the deployment process, making it easier to set up and maintain Azure services programmatically.

## Table of Contents

- [Prerequisites](#prerequisites)
- [Installation](#installation)
- [Configuration](#configuration)
- [Usage](#usage)
- [Important Notes](#important-notes)
- [License](#license)

## Prerequisites

Before you begin, ensure you have met the following requirements:

- **Operating System:** Windows, macOS, or Linux
- **Go:** [Download and install Go](https://golang.org/dl/) (version 1.16 or later)
- **Node.js and npm:** [Download and install Node.js](https://nodejs.org/) (includes npm)
- **Azure CLI:** [Download and install Azure CLI](https://docs.microsoft.com/en-us/cli/azure/install-azure-cli)

### Installing Azure CLI

Azure CLI is essential for interacting with Azure services from the command line.

#### Windows

Download and run the [Azure CLI MSI installer](https://aka.ms/installazurecliwindows).

#### macOS

1. Use Homebrew:
   ```bash
   brew update && brew install azure-cli

## Installation
Follow these steps to set up and run the project:

1. Clone the repository to your local machine using Git:
   ```bash
   git clone https://github.com/yourusername/azure-functions-go-project.git
   cd <your project directory>
2. Install Go Modules:
   ```bash
   go get ./...
3. Install Azure Functions Core Tools
   ```bash
   npm install -g azure-functions-core-tools@4 --unsafe-perm true
   
## Configuration

1. Create .env file:
   ```bash
   AZURE_SUBSCRIPTION_ID=your-azure-subscription-id
   AZURE_LOCATION=westus
   AZURE_RESOURCE_GROUP_NAME=your-resource-group-name
   AZURE_STORAGE_ACCOUNT_NAME=your-storage-account-name
   AZURE_FUNCTION_APP_NAME=your-function-app-name

   FUNCTION_NAME=YourFunctionName
   FUNCTION_TEMPLATE=HTTP trigger
   AUTH_LEVEL=anonymous
   
   KEEP_RESOURCE=1

2. Create a New Function App Directory for Each Run
   ```bash
   mkdir C:\Project\jx\functionapp_<unique_identifier>

## Usage

1. Navigate to the Project Directory:
   ```bash
   cd C:\Project\jx
2. Run the Go Application
   ```bash
   go run main.go

## Important Notes
1. Unique Function App Directory: Ensure you create a new Function App directory for each run as the application does not support overwriting existing directories. This prevents conflicts and potential data loss.
2. Secure Your .env File
3. Azure CLI and Functions Core Tools: Confirm that both the Azure CLI (az) and Azure Functions Core Tools (func) are installed and accessible in your system's PATH.

## License
This project is licensed under the MIT License.
