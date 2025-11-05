# 🚀 Deployment Guide

## Overview

This guide covers deploying Raqeem to production environments, including Kubernetes configurations, cloud platforms, monitoring setup, and operational best practices.

## Table of Contents

- [Deployment Options](#deployment-options)
- [Local Kubernetes Deployment](#local-kubernetes-deployment)
- [Production Kubernetes Deployment](#production-kubernetes-deployment)
- [AWS Deployment](#aws-deployment)
- [Environment Configuration](#environment-configuration)
- [Monitoring and Logging](#monitoring-and-logging)
- [Backup and Disaster Recovery](#backup-and-disaster-recovery)
- [Security Hardening](#security-hardening)
- [Performance Tuning](#performance-tuning)
- [Troubleshooting](#troubleshooting)

## Deployment Options

### 1. Local Development (Docker Desktop + Kubernetes)
- **Use Case**: Development, testing
- **Complexity**: Low
- **Setup Time**: <5 minutes
- **See**: [FIRST_TIME_SETUP.md](FIRST_TIME_SETUP.md)

### 2. Self-Hosted Kubernetes
- **Use Case**: On-premise deployment
- **Complexity**: Medium
- **Platforms**: Any Kubernetes cluster
- **See**: [Production Kubernetes Deployment](#production-kubernetes-deployment)

### 3. Managed Kubernetes (Cloud)
- **Use Case**: Production deployment with high availability
- **Complexity**: Medium-High
- **Platforms**: AWS EKS, Google GKE, Azure AKS
- **See**: [AWS Deployment](#aws-deployment)

## Local Kubernetes Deployment

### Quick Start

```bash
# Clone repository
git clone https://github.com/mj-nehme/raqeem.git
cd raqeem

# Start everything
./start.sh
```

### What Gets Deployed

1. **PostgreSQL** - Database (PVC-backed)
2. **MinIO** - Object storage (PVC-backed)
3. **Devices Backend** - FastAPI service
4. **Mentor Backend** - Go service
5. **Frontends** - React dev servers (local)

### Service Access

- **Mentor Dashboard**: `http://localhost:<auto-detected>`
- **Device Simulator**: `http://localhost:<auto-detected>`
- **Devices API**: `http://localhost:30080/docs`
- **Mentor API**: `http://localhost:30081/health`
- **MinIO Console**: `http://localhost:30001`

### Stopping Services

```bash
./stop.sh

# To delete all data
kubectl delete pvc --all -n default
```

## Production Kubernetes Deployment

### Prerequisites

- Kubernetes 1.24+
- kubectl configured with cluster access
- Helm 3.x installed
- Docker registry (Docker Hub, ECR, GCR)
- Domain name and TLS certificates (recommended)

### Architecture Overview

```
                    ┌──────────────┐
                    │   Internet   │
                    └──────┬───────┘
                           │
                    ┌──────▼───────┐
                    │  LoadBalancer│  (or Ingress)
                    │   / HTTPS    │
                    └──────┬───────┘
                           │
        ┌──────────────────┼──────────────────┐
        │                  │                  │
   ┌────▼────┐      ┌──────▼──────┐   ┌──────▼──────┐
   │ Mentor  │      │   Devices   │   │   Frontend  │
   │ Backend │      │   Backend   │   │   (Static)  │
   │ (Go)    │      │  (FastAPI)  │   │             │
   └────┬────┘      └──────┬──────┘   └─────────────┘
        │                  │
        └────────┬─────────┘
                 │
      ┌──────────▼──────────┐
      │    PostgreSQL       │
      │  (RDS/CloudSQL)     │
      └──────────┬──────────┘
                 │
      ┌──────────▼──────────┐
      │      MinIO/S3       │
      │  (Object Storage)   │
      └─────────────────────┘
```

### Step 1: Prepare Infrastructure

#### Create Namespace

```bash
kubectl create namespace raqeem-prod
```

#### Create Secrets

```bash
# Database credentials
kubectl create secret generic postgres-secret \
  --from-literal=username=monitor \
  --from-literal=password=<STRONG_PASSWORD> \
  --from-literal=database=monitoring_db \
  -n raqeem-prod

# MinIO credentials
kubectl create secret generic minio-secret \
  --from-literal=access-key=<ACCESS_KEY> \
  --from-literal=secret-key=<SECRET_KEY> \
  -n raqeem-prod

# Backend configuration
kubectl create secret generic backend-secret \
  --from-literal=devices-db-url=postgresql://monitor:<PASSWORD>@postgres-service:5432/monitoring_db \
  --from-literal=mentor-db-url=postgresql://monitor:<PASSWORD>@postgres-service:5432/monitoring_db \
  -n raqeem-prod
```

### Step 2: Deploy Database

#### Option A: In-Cluster PostgreSQL

```bash
# Create production values file
cat > postgres-prod-values.yaml <<EOF
replicaCount: 1

image:
  repository: postgres
  tag: "16"

persistence:
  enabled: true
  storageClass: "standard"  # Change to your storage class
  size: 20Gi  # Adjust based on needs

resources:
  limits:
    memory: "2Gi"
    cpu: "1000m"
  requests:
    memory: "1Gi"
    cpu: "500m"

env:
  - name: POSTGRES_USER
    valueFrom:
      secretKeyRef:
        name: postgres-secret
        key: username
  - name: POSTGRES_PASSWORD
    valueFrom:
      secretKeyRef:
        name: postgres-secret
        key: password
  - name: POSTGRES_DB
    valueFrom:
      secretKeyRef:
        name: postgres-secret
        key: database

service:
  type: ClusterIP
  port: 5432
EOF

# Deploy
helm upgrade --install postgres ./charts/postgres \
  -f postgres-prod-values.yaml \
  -n raqeem-prod
```

#### Option B: Managed Database (Recommended for Production)

Use AWS RDS, Google Cloud SQL, or Azure Database for PostgreSQL:

```bash
# Skip in-cluster PostgreSQL deployment
# Update backend configuration to point to managed database
kubectl create secret generic postgres-connection \
  --from-literal=url=postgresql://user:pass@rds-endpoint:5432/dbname \
  -n raqeem-prod
```

### Step 3: Deploy Object Storage

#### Option A: In-Cluster MinIO

```bash
# Create production values file
cat > minio-prod-values.yaml <<EOF
replicaCount: 1

image:
  repository: minio/minio
  tag: "latest"

persistence:
  enabled: true
  storageClass: "standard"
  size: 50Gi  # Adjust for screenshot storage needs

resources:
  limits:
    memory: "1Gi"
    cpu: "500m"
  requests:
    memory: "512Mi"
    cpu: "250m"

env:
  - name: MINIO_ROOT_USER
    valueFrom:
      secretKeyRef:
        name: minio-secret
        key: access-key
  - name: MINIO_ROOT_PASSWORD
    valueFrom:
      secretKeyRef:
        name: minio-secret
        key: secret-key

service:
  type: ClusterIP
  port: 9000
  consolePort: 9001
EOF

# Deploy
helm upgrade --install minio ./charts/minio \
  -f minio-prod-values.yaml \
  -n raqeem-prod
```

#### Option B: Cloud Object Storage (Recommended)

Use AWS S3, Google Cloud Storage, or Azure Blob Storage:

```bash
# Skip MinIO deployment
# Update backend configuration with S3 credentials
kubectl create secret generic s3-credentials \
  --from-literal=access-key=<AWS_ACCESS_KEY> \
  --from-literal=secret-key=<AWS_SECRET_KEY> \
  --from-literal=bucket=raqeem-screenshots \
  --from-literal=region=us-east-1 \
  -n raqeem-prod
```

### Step 4: Deploy Backends

#### Devices Backend

```bash
cat > devices-backend-prod-values.yaml <<EOF
replicaCount: 3  # Multiple replicas for HA

image:
  repository: mjnehme/raqeem-devices-backend
  tag: "v1.0.0"  # Use specific version
  pullPolicy: IfNotPresent

resources:
  limits:
    memory: "1Gi"
    cpu: "1000m"
  requests:
    memory: "512Mi"
    cpu: "500m"

env:
  - name: DATABASE_URL
    valueFrom:
      secretKeyRef:
        name: backend-secret
        key: devices-db-url
  - name: MINIO_ENDPOINT
    value: "minio-service:9000"
  - name: MINIO_ACCESS_KEY
    valueFrom:
      secretKeyRef:
        name: minio-secret
        key: access-key
  - name: MINIO_SECRET_KEY
    valueFrom:
      secretKeyRef:
        name: minio-secret
        key: secret-key
  - name: MENTOR_API_URL
    value: "http://mentor-backend:8080"
  - name: PORT
    value: "8080"

service:
  type: LoadBalancer  # or ClusterIP with Ingress
  port: 80
  targetPort: 8080

livenessProbe:
  httpGet:
    path: /health
    port: 8080
  initialDelaySeconds: 30
  periodSeconds: 10

readinessProbe:
  httpGet:
    path: /health
    port: 8080
  initialDelaySeconds: 10
  periodSeconds: 5
EOF

helm upgrade --install devices-backend ./charts/devices-backend \
  -f devices-backend-prod-values.yaml \
  -n raqeem-prod
```

#### Mentor Backend

```bash
cat > mentor-backend-prod-values.yaml <<EOF
replicaCount: 3  # Multiple replicas for HA

image:
  repository: mjnehme/raqeem-mentor-backend
  tag: "v1.0.0"
  pullPolicy: IfNotPresent

resources:
  limits:
    memory: "512Mi"
    cpu: "500m"
  requests:
    memory: "256Mi"
    cpu: "250m"

env:
  - name: DATABASE_URL
    valueFrom:
      secretKeyRef:
        name: backend-secret
        key: mentor-db-url
  - name: PORT
    value: "8080"
  - name: FRONTEND_ORIGIN
    value: "https://dashboard.raqeem.example.com"

service:
  type: LoadBalancer  # or ClusterIP with Ingress
  port: 80
  targetPort: 8080

livenessProbe:
  httpGet:
    path: /health
    port: 8080
  initialDelaySeconds: 30
  periodSeconds: 10

readinessProbe:
  httpGet:
    path: /health
    port: 8080
  initialDelaySeconds: 10
  periodSeconds: 5
EOF

helm upgrade --install mentor-backend ./charts/mentor-backend \
  -f mentor-backend-prod-values.yaml \
  -n raqeem-prod
```

### Step 5: Deploy Frontends

For production, build and serve static files:

```bash
# Build frontends
cd mentor/frontend
npm install
npm run build  # Creates dist/ folder

cd ../../devices/frontend
npm install
npm run build

# Option 1: Serve with nginx
kubectl create configmap frontend-nginx-config \
  --from-file=nginx.conf=<your-nginx-config> \
  -n raqeem-prod

# Option 2: Use S3 + CloudFront (recommended)
# Upload dist/ folders to S3 bucket
aws s3 sync mentor/frontend/dist s3://raqeem-dashboard/
aws s3 sync devices/frontend/dist s3://raqeem-simulator/
```

### Step 6: Configure Ingress (Optional)

```yaml
# ingress.yaml
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: raqeem-ingress
  namespace: raqeem-prod
  annotations:
    cert-manager.io/cluster-issuer: "letsencrypt-prod"
    nginx.ingress.kubernetes.io/ssl-redirect: "true"
spec:
  ingressClassName: nginx
  tls:
  - hosts:
    - api.raqeem.example.com
    - dashboard.raqeem.example.com
    secretName: raqeem-tls
  rules:
  - host: api.raqeem.example.com
    http:
      paths:
      - path: /api/v1
        pathType: Prefix
        backend:
          service:
            name: devices-backend
            port:
              number: 80
      - path: /
        pathType: Prefix
        backend:
          service:
            name: mentor-backend
            port:
              number: 80
  - host: dashboard.raqeem.example.com
    http:
      paths:
      - path: /
        pathType: Prefix
        backend:
          service:
            name: frontend-nginx
            port:
              number: 80
```

Apply ingress:
```bash
kubectl apply -f ingress.yaml
```

## AWS Deployment

### Prerequisites

- AWS account with appropriate permissions
- AWS CLI configured
- eksctl installed
- kubectl configured

### Step 1: Create EKS Cluster

```bash
# Create cluster with eksctl
eksctl create cluster \
  --name raqeem-prod \
  --region us-east-1 \
  --nodegroup-name standard-workers \
  --node-type t3.medium \
  --nodes 3 \
  --nodes-min 2 \
  --nodes-max 5 \
  --managed

# Configure kubectl
aws eks update-kubeconfig --name raqeem-prod --region us-east-1
```

### Step 2: Set Up RDS PostgreSQL

```bash
# Create RDS instance via AWS Console or CLI
aws rds create-db-instance \
  --db-instance-identifier raqeem-db \
  --db-instance-class db.t3.medium \
  --engine postgres \
  --master-username monitor \
  --master-user-password <STRONG_PASSWORD> \
  --allocated-storage 100 \
  --vpc-security-group-ids <SECURITY_GROUP_ID> \
  --publicly-accessible false

# Get endpoint
aws rds describe-db-instances \
  --db-instance-identifier raqeem-db \
  --query 'DBInstances[0].Endpoint.Address' \
  --output text
```

### Step 3: Set Up S3 Bucket

```bash
# Create S3 bucket for screenshots
aws s3 mb s3://raqeem-screenshots-prod

# Configure CORS
cat > cors.json <<EOF
{
  "CORSRules": [
    {
      "AllowedOrigins": ["https://dashboard.raqeem.example.com"],
      "AllowedMethods": ["GET", "PUT", "POST"],
      "AllowedHeaders": ["*"],
      "MaxAgeSeconds": 3000
    }
  ]
}
EOF

aws s3api put-bucket-cors \
  --bucket raqeem-screenshots-prod \
  --cors-configuration file://cors.json
```

### Step 4: Create IAM Roles

```bash
# Create IAM policy for S3 access
cat > s3-policy.json <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Action": [
        "s3:PutObject",
        "s3:GetObject",
        "s3:ListBucket"
      ],
      "Resource": [
        "arn:aws:s3:::raqeem-screenshots-prod",
        "arn:aws:s3:::raqeem-screenshots-prod/*"
      ]
    }
  ]
}
EOF

aws iam create-policy \
  --policy-name RaqeemS3Access \
  --policy-document file://s3-policy.json
```

### Step 5: Deploy to EKS

Follow [Production Kubernetes Deployment](#production-kubernetes-deployment) steps, but:
- Use RDS endpoint for database connection
- Use S3 instead of MinIO
- Configure AWS Load Balancer Controller for services
- Use Route53 for DNS
- Use ACM for TLS certificates

### Step 6: Configure Auto-Scaling

```bash
# Install metrics server
kubectl apply -f https://github.com/kubernetes-sigs/metrics-server/releases/latest/download/components.yaml

# Create HorizontalPodAutoscaler
cat > hpa.yaml <<EOF
apiVersion: autoscaling/v2
kind: HorizontalPodAutoscaler
metadata:
  name: devices-backend-hpa
  namespace: raqeem-prod
spec:
  scaleTargetRef:
    apiVersion: apps/v1
    kind: Deployment
    name: devices-backend
  minReplicas: 2
  maxReplicas: 10
  metrics:
  - type: Resource
    resource:
      name: cpu
      target:
        type: Utilization
        averageUtilization: 70
  - type: Resource
    resource:
      name: memory
      target:
        type: Utilization
        averageUtilization: 80
EOF

kubectl apply -f hpa.yaml
```

## Environment Configuration

### Environment Variables

#### Devices Backend

| Variable | Description | Example |
|----------|-------------|---------|
| `DATABASE_URL` | PostgreSQL connection string | `postgresql://user:pass@host:5432/db` |
| `MINIO_ENDPOINT` | MinIO/S3 endpoint | `minio-service:9000` or `s3.amazonaws.com` |
| `MINIO_ACCESS_KEY` | S3 access key | `AKIAIOSFODNN7EXAMPLE` |
| `MINIO_SECRET_KEY` | S3 secret key | `wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY` |
| `MENTOR_API_URL` | Mentor backend URL | `http://mentor-backend:8080` |
| `PORT` | Server port | `8080` |

#### Mentor Backend

| Variable | Description | Example |
|----------|-------------|---------|
| `DATABASE_URL` | PostgreSQL connection string | `postgresql://user:pass@host:5432/db` |
| `PORT` | Server port | `8080` |
| `FRONTEND_ORIGIN` | Allowed CORS origins | `https://dashboard.example.com` |

### Configuration Files

Create ConfigMaps for non-sensitive configuration:

```bash
kubectl create configmap app-config \
  --from-literal=LOG_LEVEL=info \
  --from-literal=METRICS_ENABLED=true \
  -n raqeem-prod
```

## Monitoring and Logging

### Prometheus + Grafana Setup

```bash
# Add Prometheus helm repo
helm repo add prometheus-community https://prometheus-community.github.io/helm-charts
helm repo update

# Install Prometheus
helm install prometheus prometheus-community/kube-prometheus-stack \
  -n monitoring --create-namespace

# Access Grafana
kubectl port-forward -n monitoring svc/prometheus-grafana 3000:80
# Login: admin / prom-operator
```

### Application Metrics

Add metrics endpoints to backends:

**Python (FastAPI)**:
```python
from prometheus_client import Counter, Histogram, generate_latest

request_count = Counter('http_requests_total', 'Total HTTP requests')
request_latency = Histogram('http_request_duration_seconds', 'HTTP request latency')

@app.get("/metrics")
async def metrics():
    return Response(generate_latest(), media_type="text/plain")
```

**Go (Gin)**:
```go
import "github.com/prometheus/client_golang/prometheus/promhttp"

router.GET("/metrics", gin.WrapH(promhttp.Handler()))
```

### Log Aggregation

#### Option 1: ELK Stack

```bash
helm install elasticsearch elastic/elasticsearch -n logging --create-namespace
helm install kibana elastic/kibana -n logging
helm install filebeat elastic/filebeat -n logging
```

#### Option 2: Loki + Grafana

```bash
helm repo add grafana https://grafana.github.io/helm-charts
helm install loki grafana/loki-stack -n monitoring
```

### Health Checks

Both backends expose `/health` endpoints:

```bash
# Check health
curl http://devices-backend/health
curl http://mentor-backend/health
```

Configure liveness and readiness probes in Kubernetes (shown in deployment configs above).

## Backup and Disaster Recovery

### Database Backups

#### PostgreSQL Backup

```bash
# Create backup script
cat > backup-postgres.sh <<'EOF'
#!/bin/bash
TIMESTAMP=$(date +%Y%m%d_%H%M%S)
kubectl exec -n raqeem-prod postgres-0 -- \
  pg_dump -U monitor monitoring_db | \
  gzip > backup-${TIMESTAMP}.sql.gz

# Upload to S3
aws s3 cp backup-${TIMESTAMP}.sql.gz s3://raqeem-backups/postgres/
EOF

chmod +x backup-postgres.sh

# Schedule with cron
crontab -e
# Add: 0 2 * * * /path/to/backup-postgres.sh
```

#### RDS Automated Backups

Enable in AWS Console or CLI:
```bash
aws rds modify-db-instance \
  --db-instance-identifier raqeem-db \
  --backup-retention-period 7 \
  --preferred-backup-window "02:00-03:00"
```

### Object Storage Backups

#### MinIO Backup

```bash
# Use mc (MinIO Client)
mc mirror minio-service/screenshots s3://backup-bucket/screenshots
```

#### S3 Cross-Region Replication

Configure in AWS Console or with CloudFormation.

### Disaster Recovery Plan

1. **RTO (Recovery Time Objective)**: <30 minutes
2. **RPO (Recovery Point Objective)**: <24 hours

**Recovery Steps**:
1. Restore database from latest backup
2. Restore S3 objects from backup bucket
3. Redeploy applications from tagged versions
4. Verify data integrity
5. Update DNS to point to new cluster

## Security Hardening

### Network Policies

```yaml
# network-policy.yaml
apiVersion: networking.k8s.io/v1
kind: NetworkPolicy
metadata:
  name: backend-policy
  namespace: raqeem-prod
spec:
  podSelector:
    matchLabels:
      app: devices-backend
  policyTypes:
  - Ingress
  - Egress
  ingress:
  - from:
    - podSelector:
        matchLabels:
          app: mentor-backend
    ports:
    - protocol: TCP
      port: 8080
  egress:
  - to:
    - podSelector:
        matchLabels:
          app: postgres
    ports:
    - protocol: TCP
      port: 5432
```

### Pod Security Standards

```yaml
# pod-security.yaml
apiVersion: v1
kind: Namespace
metadata:
  name: raqeem-prod
  labels:
    pod-security.kubernetes.io/enforce: restricted
    pod-security.kubernetes.io/audit: restricted
    pod-security.kubernetes.io/warn: restricted
```

### TLS/SSL Configuration

Use cert-manager for automatic certificate management:

```bash
# Install cert-manager
kubectl apply -f https://github.com/cert-manager/cert-manager/releases/download/v1.13.0/cert-manager.yaml

# Create ClusterIssuer
cat > letsencrypt-prod.yaml <<EOF
apiVersion: cert-manager.io/v1
kind: ClusterIssuer
metadata:
  name: letsencrypt-prod
spec:
  acme:
    server: https://acme-v02.api.letsencrypt.org/directory
    email: admin@example.com
    privateKeySecretRef:
      name: letsencrypt-prod
    solvers:
    - http01:
        ingress:
          class: nginx
EOF

kubectl apply -f letsencrypt-prod.yaml
```

### Secrets Management

Use external secrets management:

```bash
# AWS Secrets Manager
helm install external-secrets external-secrets/external-secrets \
  -n external-secrets-system --create-namespace

# HashiCorp Vault
helm install vault hashicorp/vault -n vault --create-namespace
```

## Performance Tuning

### Database Optimization

```sql
-- Create indexes for frequently queried columns
CREATE INDEX idx_device_metrics_device_timestamp 
  ON device_metrics(device_id, timestamp DESC);

CREATE INDEX idx_device_alerts_device_level 
  ON device_alerts(device_id, level);

-- Partition large tables
CREATE TABLE device_metrics_2024_01 PARTITION OF device_metrics
  FOR VALUES FROM ('2024-01-01') TO ('2024-02-01');
```

### Backend Optimization

- **Connection Pooling**: Configure appropriate pool sizes
- **Caching**: Implement Redis for frequently accessed data
- **Query Optimization**: Use database query analysis tools
- **Async Processing**: Use message queues for heavy tasks

### Kubernetes Resource Limits

Set appropriate resource requests and limits:

```yaml
resources:
  requests:
    memory: "256Mi"
    cpu: "250m"
  limits:
    memory: "512Mi"
    cpu: "500m"
```

## Troubleshooting

### Pod Crash Loop

```bash
# Check pod status
kubectl get pods -n raqeem-prod

# View logs
kubectl logs <pod-name> -n raqeem-prod --previous

# Describe pod for events
kubectl describe pod <pod-name> -n raqeem-prod
```

### Service Unreachable

```bash
# Check service endpoints
kubectl get endpoints -n raqeem-prod

# Test from debug pod
kubectl run debug -it --rm --image=curlimages/curl -n raqeem-prod -- \
  curl http://devices-backend:8080/health
```

### Database Connection Issues

```bash
# Test database connectivity
kubectl run psql -it --rm --image=postgres:16 -n raqeem-prod -- \
  psql -h postgres-service -U monitor -d monitoring_db

# Check connection pool settings
# Increase pool size if seeing "too many connections" errors
```

### High Memory Usage

```bash
# Check resource usage
kubectl top pods -n raqeem-prod

# Analyze memory leaks
kubectl exec -it <pod-name> -n raqeem-prod -- /bin/sh
# Use debugging tools (pprof for Go, memory_profiler for Python)
```

## Maintenance

### Rolling Updates

```bash
# Update image
helm upgrade devices-backend ./charts/devices-backend \
  --set image.tag=v1.1.0 \
  -n raqeem-prod

# Monitor rollout
kubectl rollout status deployment/devices-backend -n raqeem-prod

# Rollback if needed
kubectl rollout undo deployment/devices-backend -n raqeem-prod
```

### Database Migrations

```bash
# Run migrations as a Kubernetes Job
cat > migration-job.yaml <<EOF
apiVersion: batch/v1
kind: Job
metadata:
  name: db-migration
  namespace: raqeem-prod
spec:
  template:
    spec:
      containers:
      - name: migrate
        image: mjnehme/raqeem-devices-backend:v1.1.0
        command: ["alembic", "upgrade", "head"]
        env:
        - name: DATABASE_URL
          valueFrom:
            secretKeyRef:
              name: backend-secret
              key: devices-db-url
      restartPolicy: OnFailure
EOF

kubectl apply -f migration-job.yaml
```

### Scaling

```bash
# Manual scaling
kubectl scale deployment devices-backend --replicas=5 -n raqeem-prod

# Check HPA status
kubectl get hpa -n raqeem-prod
```

## Checklist

### Pre-Deployment

- [ ] All tests pass
- [ ] Security scan completed
- [ ] Secrets configured
- [ ] Database migrated
- [ ] Monitoring configured
- [ ] Backups tested
- [ ] DNS records updated
- [ ] TLS certificates valid

### Post-Deployment

- [ ] Health checks passing
- [ ] Metrics being collected
- [ ] Logs aggregating correctly
- [ ] Backups running
- [ ] Alerts configured
- [ ] Documentation updated
- [ ] Team notified

## Additional Resources

- [Kubernetes Best Practices](https://kubernetes.io/docs/concepts/configuration/overview/)
- [12-Factor App](https://12factor.net/)
- [AWS EKS Documentation](https://docs.aws.amazon.com/eks/)
- [Helm Best Practices](https://helm.sh/docs/chart_best_practices/)
- [PostgreSQL Performance Tuning](https://wiki.postgresql.org/wiki/Performance_Optimization)

## Support

For deployment issues:
1. Check logs: `kubectl logs -n raqeem-prod <pod-name>`
2. Review events: `kubectl get events -n raqeem-prod`
3. Consult [TROUBLESHOOTING.md](TROUBLESHOOTING.md)
4. Open GitHub issue with deployment details
