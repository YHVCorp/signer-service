# Signer Client Service

A Windows service client that connects to the signer server via gRPC to receive and process file signing requests using Windows signtool.

## Features

- **gRPC Streaming**: Subscribes to the server's sign request stream
- **Encrypted Configuration**: Stores sensitive data (certificates, keys, tokens) encrypted locally
- **File Processing**: Downloads files, signs them with signtool, and uploads them back
- **Windows Service**: Can be installed and run as a Windows service
- **Configurable**: Easy CLI commands to update configuration parameters

## Installation

1. **Build the client**:
   ```bash
   go build -o signer-client.exe .
   ```

2. **Install as Windows service**:
   ```cmd
   signer-client.exe -install
   ```
   
   During installation, you'll be prompted to enter:
   - Server address (e.g., `localhost:50051`)
   - Authentication token (hidden input)
   - Signing certificate path (absolute path to .pfx/.p12 file)
   - Signing key (hidden input)
   - Signing container (hidden input)

## Usage

### Service Management
```cmd
# Install service
signer-client.exe -install

# Start service
signer-client.exe -start

# Stop service
signer-client.exe -stop

# Uninstall service
signer-client.exe -uninstall

# Run in foreground (for testing)
signer-client.exe -run
```

### Configuration Management
```cmd
# View configuration options
signer-client.exe -config

# Update specific parameters
signer-client.exe -set-token "new-token"
signer-client.exe -set-cert "C:\path\to\certificate.pfx"
signer-client.exe -set-key "new-key"
signer-client.exe -set-container "new-container"
signer-client.exe -set-server "server.example.com:50051"
```

## How It Works

1. **Service Startup**: Client connects to the gRPC server using the configured server address and token
2. **Stream Subscription**: Subscribes to the `StreamSignRequests` stream
3. **Request Processing**: For each sign request received:
   - Creates a temporary directory
   - Downloads the file from the provided URL
   - Signs the file using Windows signtool with the configured certificate and key
   - Uploads the signed file back to the server
   - Reports success/failure to the server
   - Cleans up temporary files

## Signing Command

The client uses the following signtool command template:
```cmd
signtool sign /fd SHA256 /tr http://timestamp.digicert.com /td SHA256 /f <CERT_PATH> /csp "eToken Base Cryptographic Provider" /k "[{{<KEY>}}]=<CONTAINER>" "file.exe"
```

Where:
- `<CERT_PATH>`: Path to the signing certificate file
- `<KEY>`: Signing key from configuration
- `<CONTAINER>`: Signing container from configuration

## Security

- All sensitive configuration parameters are encrypted using AES encryption
- Each parameter uses a unique random salt for encryption
- Configuration file is stored with restricted permissions (0600)
- Temporary files are automatically cleaned up after processing

## Configuration File

The client stores its configuration in `client-config.yaml` in the same directory as the executable. This file contains encrypted values for:
- Authentication token
- Certificate path
- Signing key
- Signing container
- Server address

## Requirements

- Windows operating system (for signtool and Windows service support)
- Valid code signing certificate
- Network access to the signer server
- Appropriate permissions to install/run Windows services

## Troubleshooting

1. **Service won't start**: Check that configuration file exists and contains valid encrypted data
2. **Signing fails**: Verify certificate path is correct and certificate is accessible
3. **Connection issues**: Check server address and network connectivity
4. **Permission errors**: Ensure service has appropriate permissions to access certificates and temporary directories