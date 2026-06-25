# Go File Service - gRPC Server for File Compression & Repair

High-performance gRPC microservice for compressing and repairing PDF and image files. Built with Go for speed and reliability.

## Features

- ✅ **PDF Compression** - Compress valid PDFs with quality preservation
- ✅ **PDF Repair** - Repair and compress corrupted PDFs  
- ✅ **Image Compression** - JPEG, PNG with granular quality control (60-95%)
- ✅ **Large File Support** - Handle files up to 100MB
- ✅ **Production Ready** - Error handling, validation, PM2 integration
- ✅ **Multi-Instance** - Run multiple instances with PM2 load balancing

## Quick Start

### Option 1: Automated Setup (Recommended)

```bash
cd /home/hiren/packleader/code/go-file-service
chmod +x setup.sh
./setup.sh
```

This will:
- Download dependencies
- Install protoc plugins
- Generate gRPC code
- Build the binary
- Create `.env` file with defaults
- Display next steps

### Option 2: Manual Setup

1. **Install Dependencies**
   ```bash
   go mod download
   go mod tidy
   ```

2. **Install protoc plugins**
   ```bash
   go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
   go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
   ```

3. **Generate gRPC code**
   ```bash
   protoc --go_out=. --go-grpc_out=. \
     --go_opt=module=file-service --go-grpc_opt=module=file-service \
     proto/fileservice.proto
   ```

4. **Build**
   ```bash
   go build -o file-service-server
   ```

5. **Create `.env` file**
   ```bash
   cp .env.example .env
   ```

## Running the Server

### Development (Single Instance)

```bash
# Default port 50051
./file-service-server

# Custom port
FILE_SERVICE_PORT=50052 ./file-service-server

# Using .env file
export $(cat .env | xargs)
./file-service-server
```

### Production (PM2 with Auto-Restart)

#### Setup PM2

```bash
# Install PM2 globally
npm install -g pm2

# Start with PM2 (uses pm2.json config)
pm2 start pm2.json

# Monitor
pm2 logs file-service

# Setup auto-restart on reboot
pm2 startup
pm2 save
```

#### Run Specific Instances

```bash
# Start single instance on port 50051
pm2 start --name "file-service-1" ./file-service-server -- --port 50051

# Start multiple instances for load balancing
pm2 start pm2.json
```

#### PM2 Management

```bash
# View all processes
pm2 list

# View detailed logs
pm2 logs file-service

# Restart service
pm2 restart file-service

# Stop service
pm2 stop file-service

# Delete service
pm2 delete file-service

# Save current PM2 state
pm2 save

# Restore saved state after reboot
pm2 resurrect
```

## Configuration

### Environment Variables

Create `.env` file in the project root:

```env
# Service Port (default: 50051)
FILE_SERVICE_PORT=50051

# Maximum file size in MB (default: 100)
MAX_FILE_SIZE_MB=100

# Temporary directory for processing
TEMP_DIR=/tmp/file-service

# Log level: debug, info, warn, error
LOG_LEVEL=info
```

### PM2 Configuration (pm2.json)

For production multi-instance setup:

```json
{
  "apps": [
    {
      "name": "file-service",
      "script": "./file-service-server",
      "instances": 4,
      "exec_mode": "cluster",
      "env": {
        "FILE_SERVICE_PORT": "50051"
      },
      "error_file": "logs/error.log",
      "out_file": "logs/out.log",
      "log_date_format": "YYYY-MM-DD HH:mm:ss Z",
      "merge_logs": true
    }
  ]
}
```

## API Reference

### CompressImage
Compress image with specified quality (60-95).

**Request:** `{ file_data, mime_type, quality }`  
**Response:** `{ status, compressed_data, original_size, compressed_size, compression_ratio, error_message }`

### CompressPDF
Compress a valid PDF.

**Request:** `{ file_data, operation, preserve_quality }`  
**Response:** `{ status, processed_data, original_size, compressed_size, compression_ratio, was_corrupted, was_repaired, error_message }`

### RepairPDF
Repair and compress a corrupted PDF.

**Request:** `{ file_data, operation, preserve_quality }`  
**Response:** Same as CompressPDF

### GetFileStats
Get file information (size, page count, validity).

**Request:** `{ file_data }`  
**Response:** `{ original_size, page_count, file_type, is_valid }`

## Integration with Node.js Backend

The service is already integrated in `packleader-BE`. To connect:

```javascript
const FileCompressionClient = require('./helpers/grpcFileClient');

const grpcClient = new FileCompressionClient();
await grpcClient.initialize();

// Compress image
const result = await grpcClient.compressImage(buffer, 'image/jpeg', 85);

// Compress PDF
const result = await grpcClient.compressPdf(pdfBuffer);

// Repair corrupted PDF
const result = await grpcClient.repairPdf(pdfBuffer);
```

## Performance

Typical compression ratios on 4-core machine:

| File Type | Size | Reduction | Time |
|-----------|------|-----------|------|
| PDF | < 5MB | 20-40% | 100-300ms |
| PDF | 5-50MB | 20-40% | 500ms-2s |
| Image (JPEG) | < 2MB | 10-30% | 50-100ms |
| Image (PNG) | < 2MB | 5-15% | 50-100ms |

## Troubleshooting

### "Connection refused" on port 50051
```bash
# Check if service is running
ps aux | grep file-service-server

# Check if port is in use
lsof -i :50051

# Start the service
./file-service-server
```

### "Failed to repair PDF"
Some PDFs are too corrupted to repair. Try with Node.js fallback or contact support with the PDF file.

### "protoc: command not found"
```bash
# macOS
brew install protobuf

# Ubuntu/Debian
sudo apt-get install protobuf-compiler

# Manual installation
# Download from: https://github.com/protocolbuffers/protobuf/releases
```

### "Cannot find module providing package"
Rebuild the project:
```bash
go mod tidy
go build -o file-service-server
```

## Project Structure

```
go-file-service/
├── proto/                      # Protocol buffer definitions
│   ├── fileservice.proto
│   ├── fileservice.pb.go
│   └── fileservice_grpc.pb.go
├── service/                    # Business logic
│   ├── pdf_service.go
│   ├── image_service.go
│   └── helper.go
├── server.go                   # gRPC server implementation
├── go.mod / go.sum             # Dependencies
├── .env.example                # Configuration template
├── pm2.json                    # PM2 production config
├── setup.sh                    # Automated setup script
└── README.md                   # This file
```

## Dependencies

- `github.com/pdfcpu/pdfcpu` - PDF processing & repair
- `github.com/disintegration/imaging` - Image compression
- `google.golang.org/grpc` - gRPC framework
- `google.golang.org/protobuf` - Protocol Buffers

## Development

### Regenerate Proto Files

```bash
protoc --go_out=. --go-grpc_out=. \
  --go_opt=module=file-service --go-grpc_opt=module=file-service \
  proto/fileservice.proto
```

### View Go Server Logs

```bash
# During development
./file-service-server

# With PM2
pm2 logs file-service

# View last 100 lines
pm2 logs file-service --lines 100
```

## License

MIT
