# Load Balancing Multiple Instances on Same Port

This guide explains how to run multiple `file-service` instances and expose them on a single port using nginx.

## Architecture

```
Client (localhost:50050)
         ↓
      Nginx (Load Balancer)
    ↙  ↓  ↓  ↘
 50051 50052 50053 50054
  ↓     ↓    ↓     ↓
[Instance 1, 2, 3, 4]
```

## Setup Steps

### 1. Verify You Have 4 Instances Running

```bash
# Start 4 instances with PM2
pm2 start pm2-multiple.json

# Verify all are running
pm2 list
```

Expected output:
```
┌─────────────────────┬─────┬─────────┬──────────┐
│ App name            │ id  │ version │ status   │
├─────────────────────┼─────┼─────────┼──────────┤
│ file-service-1      │ 0   │ N/A     │ online   │
│ file-service-2      │ 1   │ N/A     │ online   │
│ file-service-3      │ 2   │ N/A     │ online   │
│ file-service-4      │ 3   │ N/A     │ online   │
└─────────────────────┴─────┴─────────┴──────────┘
```

### 2. Install Nginx (if not already installed)

```bash
# Ubuntu/Debian
sudo apt-get update
sudo apt-get install -y nginx

# Verify
nginx -v
```

**Important:** Nginx needs to be built with gRPC support. Check if you have `--with-http_v2_module`:

```bash
nginx -V 2>&1 | grep -o 'with-http_v2'
```

If not found, you may need to compile nginx with gRPC support.

### 3. Configure Nginx

Copy the provided config file:

```bash
# Copy to nginx available sites
sudo cp nginx-grpc.conf /etc/nginx/sites-available/file-service

# Create symlink to enable it
sudo ln -s /etc/nginx/sites-available/file-service /etc/nginx/sites-enabled/file-service

# Remove default site if it conflicts
sudo rm /etc/nginx/sites-enabled/default 2>/dev/null || true
```

### 4. Test Nginx Configuration

```bash
# Test syntax
sudo nginx -t

# Expected output:
# nginx: the configuration file /etc/nginx/nginx.conf syntax is ok
# nginx: configuration file /etc/nginx/nginx.conf test is successful.
```

### 5. Restart Nginx

```bash
sudo systemctl restart nginx

# Verify it's running
sudo systemctl status nginx

# Check logs
sudo tail -f /var/log/nginx/error.log
```

## Testing

### Test Individual Instances

```bash
# Test each instance directly
grpcurl -plaintext localhost:50051 fileservice.FileService.GetFileStats
grpcurl -plaintext localhost:50052 fileservice.FileService.GetFileStats
grpcurl -plaintext localhost:50053 fileservice.FileService.GetFileStats
grpcurl -plaintext localhost:50054 fileservice.FileService.GetFileStats
```

### Test Load Balancer (Nginx)

```bash
# Test through nginx load balancer
grpcurl -plaintext localhost:50050 fileservice.FileService.GetFileStats
```

### Make Requests from Node.js

Update your Node.js client to use the load balancer port:

```javascript
const GRPC_SERVER_URL = "localhost:50050";  // Nginx load balancer
```

Or via environment variable:

```bash
GRPC_FILE_SERVER_URL=localhost:50050 npm start
```

## Configuration Options

### Change Number of Instances

Edit `pm2-multiple.json` and add/remove app entries. Make sure to:
1. Change the app name (file-service-N)
2. Change the port (50051, 50052, etc.)
3. Update `nginx-grpc.conf` to include all new ports

### Change Load Balancing Strategy

Default is round-robin. To use other strategies, edit `nginx-grpc.conf`:

```nginx
# Least connections (recommended for gRPC)
upstream file_service_grpc {
    least_conn;
    server localhost:50051;
    server localhost:50052;
    server localhost:50053;
    server localhost:50054;
}

# Weighted round-robin
upstream file_service_grpc {
    server localhost:50051 weight=3;
    server localhost:50052 weight=2;
    server localhost:50053 weight=1;
    server localhost:50054 weight=1;
}

# IP hash (keeps client mapped to one server)
upstream file_service_grpc {
    ip_hash;
    server localhost:50051;
    server localhost:50052;
    server localhost:50053;
    server localhost:50054;
}
```

### Change Nginx Listen Port

Edit the `listen` line in `nginx-grpc.conf`:

```nginx
# Current (port 50050)
listen 50050 http2;

# Change to different port
listen 9000 http2;  # New port
```

## Monitoring

### View PM2 Status

```bash
# List all processes
pm2 list

# Monitor in real-time
pm2 monit

# View logs for specific instance
pm2 logs file-service-1
pm2 logs file-service-2
```

### View Nginx Logs

```bash
# Access logs (load balanced requests)
sudo tail -f /var/log/nginx/access.log

# Error logs
sudo tail -f /var/log/nginx/error.log
```

## Troubleshooting

### "Permission denied" on nginx config

```bash
sudo chown root:root /etc/nginx/sites-available/file-service
sudo chmod 644 /etc/nginx/sites-available/file-service
```

### Instance crashes but nginx still routing

Nginx will keep trying to send requests to crashed instance. Monitor PM2:

```bash
pm2 list  # Check if instances are running
pm2 restart all  # Restart crashed instances
```

### High latency through nginx

Adjust timeouts in `nginx-grpc.conf`:

```nginx
grpc_connect_timeout 20s;
grpc_send_timeout 20s;
grpc_read_timeout 20s;
```

### Nginx doesn't support gRPC

```bash
# Check if compiled with HTTP/2
nginx -V 2>&1 | grep http_v2

# If missing, recompile with:
# ./configure --with-http_v2_module
```

## Performance Tips

1. **Use least_conn load balancing** - Best for gRPC with long-lived connections
2. **Disable gzip** - HTTP/2 compression is automatic
3. **Increase buffer_size** - For large file uploads (4m default)
4. **Monitor connections** - Use `watch 'netstat -an | grep 500[1-4]'`

## Cleanup

To stop everything:

```bash
# Stop PM2
pm2 delete all
pm2 save

# Stop nginx
sudo systemctl stop nginx

# Remove nginx config
sudo rm /etc/nginx/sites-enabled/file-service
sudo systemctl restart nginx
```

## Summary

| Configuration | Value |
|---|---|
| Nginx Listen Port | 50050 |
| Backend Instances | 4 (50051-50054) |
| Load Balancing | Round-robin (default) |
| Protocol | gRPC over HTTP/2 |
| Auto-restart | Yes (PM2) |
| Memory Limit | 2GB per instance |

Client connects to **localhost:50050** instead of **localhost:50051**.
