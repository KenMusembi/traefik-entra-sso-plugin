# Traefik Entra SSO Plugin

A Traefik plugin written in Go that uses Microsoft Entra Single Sign-On (SSO) to authenticate users before accessing a request.

## Setup

### Prerequisites

- Go (https://golang.org/dl/)
- Traefik (https://doc.traefik.io/traefik/)

### Installation

1. Clone the repository:
   ```bash
   git clone https://github.com/yourusername/traefik-entra-sso-plugin.git
   cd traefik-entra-sso-plugin
   ```

2. Create a .env file with your Azure credentials:
    ```bash
    env
    AZURE_TENANT_ID=your-tenant-id
    AZURE_CLIENT_ID=your-client-id
    AZURE_CLIENT_SECRET=your-client-secret
    ```

3. Install dependencies:
    ```bash
    go mod tidy
    ```

4. Run the application:
    ```bash
    cd cmd
    go run main.go
    ```
Note that this will use port 8000, you can change this in main.go file.

### Usage
    Update your Traefik configuration to use the plugin.

    Build and start Traefik.

### Contributing
    Contributions are welcome. Please open an issue or submit a pull request.

### License
    This project is licensed under the MIT License.