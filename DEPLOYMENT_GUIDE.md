# VJVector Local Cluster Deployment Guide

## Overview
This guide explains how to deploy VJVector as a local cluster using Docker Compose. The setup includes a production-ready cluster configuration and a development-friendly setup.

## Architecture

### Production Cluster (`docker compose.yml`)
- **3-node etcd cluster** for coordination
- **1 master node** + **2 slave nodes** for VJVector
- **HAProxy** for load balancing
- **Prometheus** + **Grafana** for monitoring
- **Redis** for caching

### Development Cluster (`docker compose.dev.yml`)
- **Single etcd node** for simplicity
- **1 master node** + **1 slave node** for VJVector
- **Nginx** for simple load balancing
- **Development tools** (debugger, hot reloading)
- **Jaeger** for distributed tracing

## Prerequisites

- Docker and Docker Compose installed
- At least 8GB RAM available
- At least 20GB disk space
- Ports 80, 8080-8085, 9090, 3000, 6379 available

## Quick Start

### 1. Production Deployment

```bash
# Start the production cluster
docker compose up -d

# Check status
docker compose ps

# View logs
docker compose logs -f vjvector-master
```

### 2. Development Deployment

```bash
# Start the development cluster
docker compose -f docker compose.dev.yml up -d

# Check status
docker compose -f docker compose.dev.yml ps

# View logs
docker compose -f docker compose.dev.yml logs -f vjvector-master-dev
```

## Service Endpoints

### Production Cluster
- **VJVector Master**: http://localhost:8080
- **VJVector Slave 1**: http://localhost:8082
- **VJVector Slave 2**: http://localhost:8084
- **Load Balancer**: http://localhost:80
- **HAProxy Stats**: http://localhost:8404 (admin:admin123)
- **Prometheus**: http://localhost:9090
- **Grafana**: http://localhost:3000 (admin:admin)
- **Redis**: localhost:6379

### Development Cluster
- **VJVector Master**: http://localhost:8080
- **VJVector Slave**: http://localhost:8082
- **Load Balancer**: http://localhost:80
- **Prometheus**: http://localhost:9090
- **Grafana**: http://localhost:3000 (admin:admin)
- **Jaeger**: http://localhost:16686
- **Redis**: localhost:6379

## Configuration

### Environment Variables
Key environment variables for VJVector nodes:

```bash
NODE_ROLE=master|slave          # Node role in cluster
NODE_ID=unique-id               # Unique node identifier
ETCD_ENDPOINTS=host:port        # etcd cluster endpoints
SHARD_COUNT=8                   # Number of shards
REPLICA_COUNT=3                 # Number of replicas
HEARTBEAT_INTERVAL=5s           # Heartbeat interval
ELECTION_TIMEOUT=10s            # Leader election timeout
MAX_CONCURRENT_REQUESTS=1000    # Max concurrent requests
REQUEST_TIMEOUT=30s             # Request timeout
HEALTH_CHECK_INTERVAL=30s       # Health check interval
ENABLE_AUTH=false               # Enable authentication
LOG_LEVEL=info                  # Logging level
```

### Configuration Files
- **`config.cluster.yaml`**: Cluster-specific configuration
- **`config.yaml`**: Default configuration
- **`deploy/local/haproxy.cfg`**: HAProxy configuration
- **`deploy/local/nginx.dev.conf`**: Nginx development configuration

## Monitoring and Observability

### Prometheus Metrics
VJVector exposes metrics at `/metrics` endpoint:
- Cluster health metrics
- Performance metrics
- Resource usage metrics
- Custom business metrics

### Grafana Dashboards
Pre-configured dashboards for:
- Cluster overview
- Node performance
- Resource utilization
- Request patterns

### Health Checks
- **`/health`**: Overall service health
- **`/ready`**: Service readiness
- **`/metrics`**: Prometheus metrics

## Development Features

### Hot Reloading
Development containers mount source code for live updates:
```bash
# Source code changes are reflected immediately
docker compose -f docker compose.dev.yml restart vjvector-master-dev
```

### Debugging
- **Delve debugger** available on ports 2345, 2346
- **Profiling endpoints** at `/debug/pprof`
- **Structured logging** with configurable levels

### Tracing
- **Jaeger** for distributed tracing
- **OpenTelemetry** support
- **Request correlation** across nodes

## Scaling

### Add More Nodes
```bash
# Scale slave nodes
docker compose up -d --scale vjvector-slave=5

# Update load balancer configuration
# Edit haproxy.cfg or nginx.conf
```

### Horizontal Scaling
- Add more VJVector nodes
- Configure sharding strategy
- Adjust replica count
- Update load balancer

## Troubleshooting

### Common Issues

#### 1. etcd Connection Issues
```bash
# Check etcd health
docker compose exec etcd-0 etcdctl endpoint health

# View etcd logs
docker compose logs etcd-0
```

#### 2. Node Communication Issues
```bash
# Check node status
curl http://localhost:8080/cluster/status

# View cluster logs
docker compose logs vjvector-master
```

#### 3. Resource Issues
```bash
# Check resource usage
docker stats

# View container logs
docker compose logs -f
```

### Debug Commands
```bash
# Enter container
docker compose exec vjvector-master sh

# Check configuration
cat /app/config.cluster.yaml

# View logs
tail -f /app/logs/vjvector.log

# Check etcd connectivity
etcdctl --endpoints=etcd-0:2379 endpoint health
```

## Performance Tuning

### Memory Optimization
- Adjust `SHARD_COUNT` based on available memory
- Configure `MAX_CONCURRENT_REQUESTS`
- Tune JVM/Go GC parameters

### Network Optimization
- Use host networking for high-throughput scenarios
- Configure connection pooling
- Tune timeout values

### Storage Optimization
- Use SSD storage for etcd
- Configure appropriate cache sizes
- Enable compression where beneficial

## Security Considerations

### Production Security
- Enable authentication (`ENABLE_AUTH=true`)
- Use TLS for all communications
- Implement proper secrets management
- Configure firewall rules

### Development Security
- Disable authentication for ease of development
- Use local network isolation
- Monitor resource usage

## Backup and Recovery

### Data Backup
```bash
# Backup etcd data
docker compose exec etcd-0 etcdctl snapshot save backup.db

# Backup VJVector data
docker cp vjvector-master:/app/data ./backup/
```

### Disaster Recovery
- Restore etcd from snapshot
- Restore VJVector data
- Rebuild cluster from backup

## Maintenance

### Updates
```bash
# Update images
docker compose pull

# Restart services
docker compose up -d
```

### Log Rotation
- Configure log rotation in `config.cluster.yaml`
- Monitor log file sizes
- Implement log aggregation

### Health Monitoring
- Regular health check reviews
- Performance metric analysis
- Resource usage monitoring

## Support and Resources

### Documentation
- [VJVector Architecture](./docs/architecture/)
- [API Reference](./docs/api/)
- [Configuration Guide](./docs/configuration/)

### Monitoring
- [Grafana Dashboards](./deployments/grafana/)
- [Prometheus Rules](./deployments/prometheus/)
- [Alerting Configuration](./deployments/alerting/)

### Troubleshooting
- [Common Issues](./docs/troubleshooting/)
- [Performance Tuning](./docs/performance/)
- [Security Best Practices](./docs/security/)
