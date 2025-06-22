# 🚀 Deployment Guide

This guide covers various deployment options for ProjectFlow, from local development to production environments.

## Deployment Options Overview

| Option | Best For | Complexity | Scalability | Cost |
|--------|----------|------------|-------------|------|
| **Local Development** | Development, Testing | Low | Single user | Free |
| **Docker Compose** | Small teams, Staging | Medium | Limited | Low |
| **Kubernetes** | Production, Enterprise | High | High | Variable |
| **Cloud Platforms** | Managed deployments | Medium | High | Variable |

## Local Development Deployment

### Quick Start

1. **Clone and run**:
   ```bash
   git clone https://github.com/aykay76/projectflow.git
   cd projectflow
   go run cmd/server/main.go
   ```

2. **Access the application**:
   - Web Interface: http://localhost:16191
   - Health Check: http://localhost:16191/health

### Environment Configuration

```bash
# Server settings
export PORT=16191
export LOG_LEVEL=INFO

# Storage (file-based for development)
export STORAGE_TYPE=file
export DATA_DIR=./data

# Chat Interface (optional)
export LLM_PROVIDER=groq
export LLM_API_KEY=your_groq_api_key
```

## Docker Deployment

### Single Container

1. **Build the image**:
   ```bash
   podman build -t projectflow .
   ```

2. **Run with file storage**:
   ```bash
   podman run -d \
     --name projectflow \
     -p 16191:16191 \
     -v $(pwd)/data:/app/data \
     -e LLM_PROVIDER=groq \
     -e LLM_API_KEY=your_api_key \
     projectflow
   ```

3. **Run with PostgreSQL**:
   ```bash
   podman run -d \
     --name projectflow \
     -p 16191:16191 \
     -e STORAGE_TYPE=postgres \
     -e DB_HOST=your_db_host \
     -e DB_USER=projectflow \
     -e DB_PASSWORD=your_password \
     -e DB_NAME=projectflow \
     -e LLM_PROVIDER=groq \
     -e LLM_API_KEY=your_api_key \
     projectflow
   ```

### Docker Compose

Create `docker-compose.yml`:

```yaml
version: '3.8'

services:
  projectflow:
    build: .
    ports:
      - "16191:16191"
    environment:
      - STORAGE_TYPE=postgres
      - DB_HOST=postgres
      - DB_USER=projectflow
      - DB_PASSWORD=projectflow_password
      - DB_NAME=projectflow
      - LLM_PROVIDER=groq
      - LLM_API_KEY=${LLM_API_KEY}
    depends_on:
      postgres:
        condition: service_healthy
    volumes:
      - ./data:/app/data
    restart: unless-stopped

  postgres:
    image: postgres:16
    environment:
      - POSTGRES_DB=projectflow
      - POSTGRES_USER=projectflow
      - POSTGRES_PASSWORD=projectflow_password
    volumes:
      - postgres_data:/var/lib/postgresql/data
      - ./scripts/init.sql:/docker-entrypoint-initdb.d/init.sql
    ports:
      - "5432:5432"
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U projectflow"]
      interval: 10s
      timeout: 5s
      retries: 5
    restart: unless-stopped

  # Optional: Ollama for local LLM
  ollama:
    image: ollama/ollama:latest
    ports:
      - "11434:11434"
    volumes:
      - ollama_data:/root/.ollama
    restart: unless-stopped

volumes:
  postgres_data:
  ollama_data:
```

**Deploy**:
```bash
# Set your API key
export LLM_API_KEY=your_groq_api_key

# Start services
docker-compose up -d

# Check status
docker-compose ps

# View logs
docker-compose logs -f projectflow
```

## Kubernetes Deployment

### Basic Deployment

1. **Create namespace**:
   ```yaml
   # namespace.yaml
   apiVersion: v1
   kind: Namespace
   metadata:
     name: projectflow
   ```

2. **ConfigMap for environment variables**:
   ```yaml
   # configmap.yaml
   apiVersion: v1
   kind: ConfigMap
   metadata:
     name: projectflow-config
     namespace: projectflow
   data:
     PORT: "16191"
     LOG_LEVEL: "INFO"
     STORAGE_TYPE: "postgres"
     DB_HOST: "postgres-service"
     DB_USER: "projectflow"
     DB_NAME: "projectflow"
     LLM_PROVIDER: "groq"
   ```

3. **Secret for sensitive data**:
   ```yaml
   # secret.yaml
   apiVersion: v1
   kind: Secret
   metadata:
     name: projectflow-secrets
     namespace: projectflow
   type: Opaque
   data:
     DB_PASSWORD: cHJvamVjdGZsb3dfcGFzc3dvcmQ=  # base64 encoded
     LLM_API_KEY: eW91cl9ncm9xX2FwaV9rZXk=      # base64 encoded
   ```

4. **Deployment**:
   ```yaml
   # deployment.yaml
   apiVersion: apps/v1
   kind: Deployment
   metadata:
     name: projectflow
     namespace: projectflow
   spec:
     replicas: 3
     selector:
       matchLabels:
         app: projectflow
     template:
       metadata:
         labels:
           app: projectflow
       spec:
         containers:
         - name: projectflow
           image: projectflow:latest
           ports:
           - containerPort: 16191
           envFrom:
           - configMapRef:
               name: projectflow-config
           - secretRef:
               name: projectflow-secrets
           livenessProbe:
             httpGet:
               path: /health
               port: 16191
             initialDelaySeconds: 30
             periodSeconds: 10
           readinessProbe:
             httpGet:
               path: /health
               port: 16191
             initialDelaySeconds: 5
             periodSeconds: 5
           resources:
             requests:
               memory: "128Mi"
               cpu: "100m"
             limits:
               memory: "512Mi"
               cpu: "500m"
   ```

5. **Service**:
   ```yaml
   # service.yaml
   apiVersion: v1
   kind: Service
   metadata:
     name: projectflow-service
     namespace: projectflow
   spec:
     selector:
       app: projectflow
     ports:
     - protocol: TCP
       port: 80
       targetPort: 16191
     type: ClusterIP
   ```

6. **Ingress**:
   ```yaml
   # ingress.yaml
   apiVersion: networking.k8s.io/v1
   kind: Ingress
   metadata:
     name: projectflow-ingress
     namespace: projectflow
     annotations:
       nginx.ingress.kubernetes.io/rewrite-target: /
   spec:
     rules:
     - host: projectflow.yourdomain.com
       http:
         paths:
         - path: /
           pathType: Prefix
           backend:
             service:
               name: projectflow-service
               port:
                 number: 80
   ```

**Deploy to Kubernetes**:
```bash
# Apply all resources
kubectl apply -f namespace.yaml
kubectl apply -f configmap.yaml
kubectl apply -f secret.yaml
kubectl apply -f deployment.yaml
kubectl apply -f service.yaml
kubectl apply -f ingress.yaml

# Check deployment status
kubectl get pods -n projectflow
kubectl get services -n projectflow
```

## Cloud Platform Deployments

### AWS (Elastic Container Service)

1. **Build and push image to ECR**:
   ```bash
   # Create ECR repository
   aws ecr create-repository --repository-name projectflow

   # Get login token
   aws ecr get-login-password --region us-west-2 | \
     docker login --username AWS --password-stdin \
     123456789012.dkr.ecr.us-west-2.amazonaws.com

   # Build and tag
   docker build -t projectflow .
   docker tag projectflow:latest \
     123456789012.dkr.ecr.us-west-2.amazonaws.com/projectflow:latest

   # Push
   docker push 123456789012.dkr.ecr.us-west-2.amazonaws.com/projectflow:latest
   ```

2. **Create ECS Task Definition**:
   ```json
   {
     "family": "projectflow-task",
     "networkMode": "awsvpc",
     "requiresCompatibilities": ["FARGATE"],
     "cpu": "256",
     "memory": "512",
     "executionRoleArn": "arn:aws:iam::123456789012:role/ecsTaskExecutionRole",
     "containerDefinitions": [
       {
         "name": "projectflow",
         "image": "123456789012.dkr.ecr.us-west-2.amazonaws.com/projectflow:latest",
         "portMappings": [
           {
             "containerPort": 16191,
             "protocol": "tcp"
           }
         ],
         "environment": [
           {
             "name": "STORAGE_TYPE",
             "value": "postgres"
           },
           {
             "name": "DB_HOST",
             "value": "your-rds-endpoint.amazonaws.com"
           }
         ],
         "secrets": [
           {
             "name": "DB_PASSWORD",
             "valueFrom": "arn:aws:secretsmanager:us-west-2:123456789012:secret:projectflow-db-password"
           },
           {
             "name": "LLM_API_KEY",
             "valueFrom": "arn:aws:secretsmanager:us-west-2:123456789012:secret:projectflow-llm-key"
           }
         ],
         "logConfiguration": {
           "logDriver": "awslogs",
           "options": {
             "awslogs-group": "/ecs/projectflow",
             "awslogs-region": "us-west-2",
             "awslogs-stream-prefix": "ecs"
           }
         }
       }
     ]
   }
   ```

### Google Cloud Platform (Cloud Run)

1. **Build and deploy**:
   ```bash
   # Enable required APIs
   gcloud services enable run.googleapis.com
   gcloud services enable cloudbuild.googleapis.com

   # Build and deploy
   gcloud builds submit --tag gcr.io/your-project/projectflow
   
   gcloud run deploy projectflow \
     --image gcr.io/your-project/projectflow \
     --platform managed \
     --region us-central1 \
     --allow-unauthenticated \
     --set-env-vars="STORAGE_TYPE=postgres,DB_HOST=your-cloud-sql-ip" \
     --set-secrets="DB_PASSWORD=projectflow-db-password:latest,LLM_API_KEY=projectflow-llm-key:latest"
   ```

2. **Set up Cloud SQL**:
   ```bash
   # Create PostgreSQL instance
   gcloud sql instances create projectflow-db \
     --database-version=POSTGRES_14 \
     --region=us-central1 \
     --tier=db-f1-micro

   # Create database and user
   gcloud sql databases create projectflow --instance=projectflow-db
   gcloud sql users create projectflow --instance=projectflow-db --password=your-password
   ```

## Production Considerations

### Security

1. **HTTPS/TLS**:
   ```bash
   # Use reverse proxy (nginx) for TLS termination
   # Or configure ingress controller with TLS certificates
   ```

2. **Authentication**:
   ```bash
   # Implement authentication middleware
   # Consider OAuth2, JWT, or API keys
   ```

3. **Secrets Management**:
   ```bash
   # Use Kubernetes secrets, AWS Secrets Manager, etc.
   # Never hardcode sensitive values
   ```

### Monitoring and Logging

1. **Health Checks**:
   ```yaml
   # Kubernetes health checks
   livenessProbe:
     httpGet:
       path: /health
       port: 16191
   readinessProbe:
     httpGet:
       path: /health
       port: 16191
   ```

2. **Metrics**:
   ```bash
   # ProjectFlow exposes Prometheus metrics on /metrics
   # Configure Prometheus scraping
   ```

3. **Logging**:
   ```bash
   # Structured logging is enabled by default
   # Configure log aggregation (ELK, Fluentd, etc.)
   ```

### Scaling

1. **Horizontal Pod Autoscaler (HPA)**:
   ```yaml
   apiVersion: autoscaling/v2
   kind: HorizontalPodAutoscaler
   metadata:
     name: projectflow-hpa
   spec:
     scaleTargetRef:
       apiVersion: apps/v1
       kind: Deployment
       name: projectflow
     minReplicas: 2
     maxReplicas: 10
     metrics:
     - type: Resource
       resource:
         name: cpu
         target:
           type: Utilization
           averageUtilization: 70
   ```

2. **Database Scaling**:
   ```bash
   # PostgreSQL read replicas for read-heavy workloads
   # Connection pooling (PgBouncer)
   # Database partitioning for large datasets
   ```

### Backup and Recovery

1. **Database Backups**:
   ```bash
   # Automated PostgreSQL backups
   pg_dump -h localhost -U projectflow projectflow > backup.sql
   
   # Cloud provider managed backups
   # AWS RDS automated backups
   # GCP Cloud SQL automated backups
   ```

2. **File Storage Backups**:
   ```bash
   # If using file storage, backup data directory
   tar -czf projectflow-data-$(date +%Y%m%d).tar.gz data/
   ```

### Performance Optimization

1. **Database Optimization**:
   ```sql
   -- Add indexes for common queries
   CREATE INDEX idx_tasks_status ON tasks(status);
   CREATE INDEX idx_tasks_project_id ON tasks(project_id);
   CREATE INDEX idx_tasks_created_at ON tasks(created_at);
   ```

2. **Caching**:
   ```bash
   # Consider Redis for session/cache storage
   # Implement HTTP caching headers
   ```

3. **CDN**:
   ```bash
   # Use CDN for static assets
   # CloudFront, CloudFlare, etc.
   ```

## Troubleshooting Deployments

### Common Issues

1. **Container Won't Start**:
   ```bash
   # Check logs
   docker logs projectflow
   kubectl logs -n projectflow deployment/projectflow
   
   # Check environment variables
   docker exec -it projectflow env
   ```

2. **Database Connection Issues**:
   ```bash
   # Test database connectivity
   docker exec -it projectflow \
     psql -h $DB_HOST -U $DB_USER -d $DB_NAME -c "SELECT 1;"
   ```

3. **Health Check Failures**:
   ```bash
   # Test health endpoint
   curl http://localhost:16191/health
   
   # Check application logs
   docker logs projectflow
   ```

### Debugging Commands

```bash
# Docker
docker exec -it projectflow /bin/sh
docker logs -f projectflow

# Kubernetes
kubectl exec -it deployment/projectflow -n projectflow -- /bin/sh
kubectl logs -f deployment/projectflow -n projectflow
kubectl describe pod -n projectflow

# Check service endpoints
kubectl get endpoints -n projectflow
```

## Migration and Upgrades

### Database Migrations

```bash
# Backup before migration
pg_dump -h localhost -U projectflow projectflow > pre-migration-backup.sql

# Test migration on copy
createdb projectflow_test
psql -h localhost -U projectflow projectflow_test < pre-migration-backup.sql

# Apply migration scripts
psql -h localhost -U projectflow projectflow < migration.sql
```

### Rolling Updates

```bash
# Kubernetes rolling update
kubectl set image deployment/projectflow -n projectflow \
  projectflow=projectflow:new-version

# Monitor rollout
kubectl rollout status deployment/projectflow -n projectflow

# Rollback if needed
kubectl rollout undo deployment/projectflow -n projectflow
```

### Blue-Green Deployment

```bash
# Deploy new version to green environment
kubectl create namespace projectflow-green
kubectl apply -f deployment.yaml -n projectflow-green

# Test green environment
curl http://projectflow-green.yourdomain.com/health

# Switch traffic (update ingress)
kubectl patch ingress projectflow-ingress -n projectflow \
  -p '{"spec":{"rules":[{"host":"projectflow.yourdomain.com","http":{"paths":[{"path":"/","pathType":"Prefix","backend":{"service":{"name":"projectflow-service-green","port":{"number":80}}}}]}}]}}'

# Clean up old version
kubectl delete namespace projectflow-blue
```

## Support and Maintenance

### Regular Maintenance Tasks

1. **Database Maintenance**:
   ```bash
   # Weekly vacuum and analyze
   psql -h localhost -U projectflow projectflow -c "VACUUM ANALYZE;"
   
   # Monitor database size and performance
   ```

2. **Log Rotation**:
   ```bash
   # Configure log rotation
   # Monitor disk usage
   ```

3. **Security Updates**:
   ```bash
   # Regular base image updates
   # Dependency updates
   # Security scanning
   ```

### Monitoring Checklist

- [ ] Application health endpoint responding
- [ ] Database connectivity working
- [ ] LLM provider accessible
- [ ] Disk space sufficient
- [ ] Memory usage within limits
- [ ] CPU usage normal
- [ ] Response times acceptable
- [ ] Error rates low

## Getting Help

For deployment support:

1. **Check logs first**: Application and infrastructure logs
2. **Review documentation**: This guide and troubleshooting docs
3. **Community support**: GitHub issues and discussions
4. **Professional support**: Available for enterprise deployments

---

**Next Steps**: After deployment, see the [User Guide](user-guide.md) for getting started with ProjectFlow.
