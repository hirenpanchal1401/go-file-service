#!/bin/bash

set -e

echo "╔════════════════════════════════════════════════╗"
echo "║     Go File Service - Setup Script             ║"
echo "╚════════════════════════════════════════════════╝"
echo ""

# Check if Go is installed
if ! command -v go &> /dev/null; then
    echo "❌ Go is not installed. Please install Go 1.21 or later."
    exit 1
fi

echo "✓ Go version: $(go version)"
echo ""

# Check if protoc is installed
if ! command -v protoc &> /dev/null; then
    echo "⚠️  protoc is not installed."
    echo ""
    echo "Install it using:"
    echo "  macOS:  brew install protobuf"
    echo "  Ubuntu: sudo apt-get install protobuf-compiler"
    echo "  Other:  https://github.com/protocolbuffers/protobuf/releases"
    exit 1
fi

echo "✓ protoc version: $(protoc --version)"
echo ""

# Step 1: Download dependencies
echo "📦 Downloading dependencies..."
go mod download
go mod tidy
echo "✓ Dependencies downloaded"
echo ""

# Step 2: Install protoc plugins
echo "🔧 Installing protoc plugins..."
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
echo "✓ Protoc plugins installed"
echo ""

# Step 3: Generate gRPC code
echo "⚙️  Generating gRPC code..."
if [ ! -d "proto" ]; then
    echo "❌ proto directory not found!"
    exit 1
fi

protoc --go_out=. --go-grpc_out=. \
  --go_opt=module=file-service --go-grpc_opt=module=file-service \
  proto/fileservice.proto

echo "✓ gRPC code generated"
echo "  - proto/fileservice.pb.go"
echo "  - proto/fileservice_grpc.pb.go"
echo ""

# Step 4: Build the binary
echo "🏗️  Building file-service-server..."
go build -o file-service-server .
echo "✓ Binary built: ./file-service-server"
echo ""

# Step 5: Create .env file if it doesn't exist
echo "📝 Setting up configuration..."
if [ ! -f ".env" ]; then
    cp .env.example .env
    echo "✓ Created .env from .env.example"
else
    echo "✓ .env already exists (using existing configuration)"
fi
echo ""

# Step 6: Create logs directory
if [ ! -d "logs" ]; then
    mkdir -p logs
    echo "✓ Created logs directory"
fi
echo ""

# Step 7: Check Node.js integration
echo "📡 Checking Node.js backend integration..."
if [ -f "../packleader-BE/src/helpers/grpcFileClient.js" ]; then
    echo "✓ gRPC client found in packleader-BE"
else
    echo "⚠️  gRPC client not found in packleader-BE"
fi
echo ""

# Final status
echo "╔════════════════════════════════════════════════╗"
echo "║     ✅ Setup Complete!                         ║"
echo "╚════════════════════════════════════════════════╝"
echo ""
echo "🚀 Next Steps:"
echo ""
echo "1. Review configuration:"
echo "   cat .env"
echo ""
echo "2. Start the server (development):"
echo "   ./file-service-server"
echo ""
echo "3. Or with custom port:"
echo "   FILE_SERVICE_PORT=50052 ./file-service-server"
echo ""
echo "4. For production with PM2:"
echo "   npm install -g pm2"
echo "   pm2 start pm2.json"
echo "   pm2 save"
echo ""
echo "5. Start Node.js backend:"
echo "   cd ../packleader-BE"
echo "   npm install"
echo "   npm start"
echo ""
echo "📖 See README.md for more details."
echo ""
