# Local Deployment Setup

This directory contains the configuration files for local VJVector cluster deployment using Docker Compose.

## Files

- **`haproxy.cfg`** - HAProxy configuration for production load balancing
- **`nginx.dev.conf`** - Nginx configuration for development load balancing
- **`README.md`** - This file

## Quick Start

### Production Cluster
```bash
# From project root
docker-compose up -d

# Or use the convenience script
./scripts/start-cluster.sh start-prod
```

### Development Cluster
```bash
# From project root
docker-compose -f docker-compose.dev.yml up -d

# Or use the convenience script
./scripts/start-cluster.sh start-dev
```

## Load Balancer Configuration

### HAProxy (Production)
- **Port**: 80 (HTTP)
- **Stats**: 8404 (admin:admin123)
- **Health Checks**: Automatic
- **Load Balancing**: Round-robin with health checks

### Nginx (Development)
- **Port**: 80 (HTTP)
- **Status**: /nginx_status
- **Simple round-robin load balancing
- **Health check endpoint support

## Health Checks

Both load balancers support health checks at:
- `/health` - Overall service health
- `/ready` - Service readiness
- `/metrics` - Prometheus metrics

## Monitoring

- **Prometheus**: http://localhost:9090
- **Grafana**: http://localhost:3000 (admin:admin)
- **HAProxy Stats**: http://localhost:8404 (admin:admin123)

## Troubleshooting

### Check Load Balancer Status
```bash
# HAProxy stats
curl -u admin:admin123 http://localhost:8404/stats

# Nginx status
curl http://localhost/nginx_status
```

### View Load Balancer Logs
```bash
# HAProxy
docker-compose logs haproxy

# Nginx (dev)
docker-compose -f docker-compose.dev.yml logs nginx-dev
```

### Test Load Balancing
```bash
# Test multiple requests to see load distribution
for i in {1..10}; do
  curl -s http://localhost/health | grep -o '"status":"[^"]*"'
  sleep 1
done
```
