# Signer Service - Distributed Digital Signing System

Distributed system for automatic signing of executable files using digital certificates, designed to integrate with CI/CD pipelines.

## 📋 Description

The **Signer Service** enables automatic and distributed file signing through:
- **Server**: Receives files via HTTP and coordinates signing tasks
- **Client**: Connects to the server and executes actual signing using digital certificates

## 🏗️ Architecture

```
CI/CD Pipeline → HTTP/gRPC Server → Signing Client → Signed File
```

## ⚡ Workflow

1. **CI/CD Pipeline** sends file to server via HTTP
2. **Server** saves the file and notifies client via gRPC
3. **Client** downloads, signs and returns the signed file
4. **Pipeline** gets the signed file when ready

## 🔧 Installation

### Prerequisites

- Windows Server/Desktop
- Code signing certificate (.pfx/.p12)
- SignTool.exe (included in Windows SDK)
- Administrator privileges

### 📦 Server Installation

1. **Compile or download** the server executable
2. **Install as service** (run as Administrator):
    ```cmd
    signer-server.exe install
    ```
3. **Save the token** displayed on screen to configure clients

**Example output:**
```
Installing SignerServiceServer service ...
Configuring server ... [OK]
Creating service ... [OK]
SignerServiceServer service installed correctly. 
You can use the token: AbCdEf123456789... for authenticate clients
```

### 🖥️ Client Installation

1. **Compile or download** the client executable
2. **Install interactively** (run as Administrator):
    ```cmd
    signer-client.exe -install
    ```
3. **Complete interactive configuration:**
    ```
    Enter server address (e.g., localhost:50051): localhost:50052
    Enter authentication token: [Server token]
    Enter signing certificate path: C:\Certificates\MyCodeSignCert.pfx
    Enter signing key: [Certificate key]
    Enter signing container: [Certificate container]
    ```

## ⚙️ Configuration

### Used Ports

- **HTTP Server**: Port `8081`
- **gRPC Server**: Port `50052`

### Service Management

**Check status:**
```cmd
sc query SignerServiceServer
sc query SignerServiceClient
```

**Restart services:**
```cmd
net stop SignerServiceServer && net start SignerServiceServer
net stop SignerServiceClient && net start SignerServiceClient
```

**View logs:**
```cmd
type C:\SignerService\server\logs\utmstack_agent.log
type C:\SignerService\client\logs\utmstack_agent.log
```

### Update Client Configuration

```cmd
# Change server
signer-client.exe -set-server "new-server:50052"

# Change token
signer-client.exe -set-token "new-token"

# Change certificate
signer-client.exe -set-cert "C:\path\to\new-certificate.pfx"
```

## 🔐 Security

- **Encrypted tokens** for client-server authentication
- **Encrypted configuration** of certificates and keys
- **Secure communication** via gRPC
- **Temporary URLs** for file download/upload
- **Automatic cleanup** of temporary files

## 📝 API Endpoints

| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/api/v1/upload` | Upload file for signing |
| GET | `/api/v1/status/:file_id` | Check signing status |
| GET | `/api/v1/download/:file_id` | Download signed file |

## 🔄 Uninstallation

```cmd
# Stop and uninstall server
signer-server.exe uninstall

# Stop and uninstall client
signer-client.exe -uninstall
```