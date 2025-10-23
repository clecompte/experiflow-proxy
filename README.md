# ExperiFlow Proxy

High-performance Go reverse proxy for server-side A/B testing with zero visual flicker.

## Features

- ✅ **Zero Flicker**: HTML transformed server-side before reaching browser
- ✅ **High Performance**: Written in Go, handles 10k+ RPS per instance
- ✅ **Deterministic Assignment**: HMAC-based bucketing for consistent user experience
- ✅ **Fail-Safe**: Configurable fail-open with timeout protection
- ✅ **Self-Hosted**: No third-party dependencies, deploy anywhere
- ✅ **Observable**: Request timing, variant assignment, transform status headers
- ✅ **Production-Ready**: Cookie management, caching, error handling

## Quick Start

### Local Development

#### 1. Start your origin server
```bash
# Terminal 1: Origin server (test page)
cd /path/to/experiflow
python -m http.server 8080
```

#### 2. Start ExperiFlow API
```bash
# Terminal 2: API
cd experiflow-api
uvicorn main:app --reload --port 8000
```

#### 3. Build and run the proxy
```bash
# Terminal 3: Proxy
cd experiflow-proxy

# Set environment variables
export ORIGIN_URL=http://localhost:8080
export EXPERIFLOW_API_URL=http://localhost:8000
export EXPERIMENT_IDS=54ce9030-4da3-4866-8b25-6d956207f325

# Build
go build -o proxy ./cmd/proxy

# Run
./proxy
```

#### 4. Test it!
```bash
# Open in browser
open http://localhost:8090/test-page.html

# Or curl
curl http://localhost:8090/test-page.html
```

**Expected Result**: No visual flicker! HTML is already transformed.

### Docker

```bash
# Build
docker build -t experiflow-proxy .

# Run
docker run -p 8090:8090 \
  -e ORIGIN_URL=http://your-site.com \
  -e EXPERIFLOW_API_URL=http://api.experiflow.com \
  -e EXPERIMENT_IDS=your-experiment-id \
  experiflow-proxy
```

### Docker Compose

```bash
# Update docker-compose.yml with your settings
docker-compose up
```

## Configuration

All configuration is done via environment variables:

### Proxy Settings

| Variable | Default | Description |
|----------|---------|-------------|
| `PORT` | `8090` | Port to listen on |
| `ORIGIN_URL` | `http://localhost:8080` | Your origin server URL |
| `READ_TIMEOUT` | `10s` | HTTP read timeout |
| `WRITE_TIMEOUT` | `10s` | HTTP write timeout |

### ExperiFlow API Settings

| Variable | Default | Description |
|----------|---------|-------------|
| `EXPERIFLOW_API_URL` | `http://localhost:8000` | ExperiFlow API base URL |
| `EXPERIFLOW_EDGE_TOKEN` | (empty) | Optional API authentication token |
| `TRANSFORM_TIMEOUT` | `50ms` | Timeout for transformation operations |

### Experiment Configuration

| Variable | Default | Description |
|----------|---------|-------------|
| `EXPERIMENT_IDS` | (empty) | Comma-separated experiment IDs to activate |

Example: `EXPERIMENT_IDS=exp1,exp2,exp3`

### Feature Flags

| Variable | Default | Description |
|----------|---------|-------------|
| `FAIL_OPEN` | `true` | Pass through on errors (recommended) |
| `ENABLE_LOGGING` | `true` | Enable request logging |
| `ENABLE_METRICS` | `true` | Enable metrics collection |

## Architecture

```
┌─────────┐      ┌──────────────────┐      ┌────────────┐
│ Browser │─────▶│ ExperiFlow Proxy │─────▶│ Your Site  │
└─────────┘      └──────────────────┘      └────────────┘
                          │
                          │ (Transform Spec)
                          ▼
                 ┌─────────────────┐
                 │ ExperiFlow API  │
                 └─────────────────┘
```

### Request Flow

1. **Request arrives** at proxy
2. **Check cookie** for existing variant assignment
3. **Assign variant** if new user (HMAC-based bucketing)
4. **Fetch transform spec** from ExperiFlow API (~2KB JSON)
5. **Fetch HTML** from origin server
6. **Apply transformations** (parse → modify → render)
7. **Return transformed HTML** with cookie set
8. **Browser receives** - no flicker!

### Performance

- Transform spec fetch: p95 < 20ms
- HTML transformation: p95 < 30ms
- **Total added latency**: p95 < 50ms
- Handles: 10,000+ RPS per instance

## Response Headers

The proxy adds observability headers to every response:

```
X-EF-Experiment: 54ce9030-4da3-4866-8b25-6d956207f325
X-EF-Variant: Green CTA Button Variant
X-EF-Transform: hit|control|miss|timeout
X-EF-Timing: total=35ms
```

Use these for debugging and monitoring.

## Deployment

### Production Checklist

- [ ] Set `FAIL_OPEN=true` (always)
- [ ] Configure `EXPERIFLOW_EDGE_TOKEN` for API auth
- [ ] Set appropriate timeouts (`TRANSFORM_TIMEOUT=50ms`)
- [ ] Deploy multiple instances (horizontal scaling)
- [ ] Add load balancer in front
- [ ] Monitor `X-EF-*` headers
- [ ] Set up alerts for timeout rate

### Kubernetes

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: experiflow-proxy
spec:
  replicas: 3
  selector:
    matchLabels:
      app: experiflow-proxy
  template:
    metadata:
      labels:
        app: experiflow-proxy
    spec:
      containers:
      - name: proxy
        image: experiflow-proxy:latest
        ports:
        - containerPort: 8090
        env:
        - name: ORIGIN_URL
          value: "http://your-app-service"
        - name: EXPERIFLOW_API_URL
          value: "https://api.experiflow.com"
        - name: EXPERIMENT_IDS
          value: "exp-id-1,exp-id-2"
        - name: FAIL_OPEN
          value: "true"
        resources:
          requests:
            memory: "64Mi"
            cpu: "100m"
          limits:
            memory: "256Mi"
            cpu: "500m"
        livenessProbe:
          httpGet:
            path: /
            port: 8090
          initialDelaySeconds: 5
          periodSeconds: 10
```

### AWS ECS / Fargate

See `ecs-task-definition.json` (TODO)

### Google Cloud Run

```bash
gcloud run deploy experiflow-proxy \
  --image gcr.io/your-project/experiflow-proxy \
  --set-env-vars ORIGIN_URL=https://your-site.com \
  --set-env-vars EXPERIFLOW_API_URL=https://api.experiflow.com \
  --set-env-vars EXPERIMENT_IDS=exp-id \
  --memory 256Mi \
  --cpu 1 \
  --max-instances 10
```

## Troubleshooting

### No transformations being applied

1. Check experiment ID is correct: `echo $EXPERIMENT_IDS`
2. Check API is reachable: `curl $EXPERIFLOW_API_URL/health`
3. Check experiment status is "running"
4. Check variant has published visual changes
5. Look at response headers: `X-EF-Transform` should be `hit`

### Timeouts

1. Increase `TRANSFORM_TIMEOUT` (but keep < 100ms)
2. Check API latency
3. Ensure `FAIL_OPEN=true`
4. Scale API horizontally

### High memory usage

1. Reduce number of concurrent transformations
2. Add resource limits in Docker/K8s
3. Scale horizontally instead of vertically

## Development

### Build

```bash
go build -o proxy ./cmd/proxy
```

### Test

```bash
go test ./...
```

### Run locally

```bash
go run ./cmd/proxy
```

### Format code

```bash
go fmt ./...
```

## License

MIT

## Support

For issues or questions:
- GitHub Issues: https://github.com/experiflow/proxy/issues
- Documentation: https://docs.experiflow.com
- Email: support@experiflow.com
# experiflow-proxy
