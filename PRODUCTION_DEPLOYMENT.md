# ExperiFlow Proxy - Production Deployment Guide

**Status:** Ready for Deployment
**Target Platform:** Railway (or any Docker-compatible platform)
**Purpose:** Server-side A/B testing with zero visual flicker
**Last Updated:** October 28, 2025

---

## Table of Contents

1. [Overview](#overview)
2. [Prerequisites](#prerequisites)
3. [Deployment Options](#deployment-options)
4. [Railway Deployment (Recommended)](#railway-deployment-recommended)
5. [Environment Variables](#environment-variables)
6. [Testing & Verification](#testing--verification)
7. [Custom Domain Setup](#custom-domain-setup)
8. [Monitoring & Troubleshooting](#monitoring--troubleshooting)
9. [Performance & Scaling](#performance--scaling)
10. [Security Considerations](#security-considerations)

---

## Overview

The ExperiFlow Proxy is a high-performance Go reverse proxy that enables server-side A/B testing by transforming HTML before it reaches the browser, eliminating visual flicker entirely.

**Architecture:**
```
User → Proxy (transforms HTML) → Your Origin Site
         ↓
   ExperiFlow API
   (provides transform specs)
```

**Key Features:**
- ✅ Zero visual flicker (HTML transformed server-side)
- ✅ High performance (10,000+ RPS per instance)
- ✅ Fail-safe design (passes through on errors)
- ✅ Observable (timing and variant assignment headers)
- ✅ Stateless (easy horizontal scaling)

---

## Prerequisites

### Required Services
- [x] ExperiFlow API deployed (https://api.experiflow.com) ✅
- [x] Origin website/application (the site you want to test)
- [ ] Experiment created in ExperiFlow with published visual changes
- [ ] Railway account (or Docker-compatible hosting)

### Required Information
- Origin site URL (e.g., `https://www.yoursite.com`)
- Experiment ID from ExperiFlow dashboard
- (Optional) Custom domain for proxy

---

## Deployment Options

### Option 1: Railway (Recommended)
- **Pros**: Easy deployment, automatic HTTPS, custom domains, good free tier
- **Cost**: $5/month (Hobby plan)
- **Setup time**: ~10 minutes

### Option 2: Docker on VPS
- **Pros**: Full control, potentially cheaper at scale
- **Cost**: Variable ($5-50/month depending on provider)
- **Setup time**: ~30 minutes

### Option 3: Kubernetes
- **Pros**: Enterprise-grade, auto-scaling
- **Cost**: Variable (typically $20+/month)
- **Setup time**: ~1-2 hours

### Option 4: AWS ECS/Fargate, Google Cloud Run, etc.
- See [README.md](README.md) for cloud-specific deployment guides

---

## Railway Deployment (Recommended)

### Step 1: Prepare Repository

The proxy is ready to deploy with these files:
- ✅ `Dockerfile` - Multi-stage build configuration
- ✅ `railway.json` - Railway-specific configuration
- ✅ `.env.production.template` - Environment variable documentation
- ✅ `cmd/proxy/main.go` - Main application with health check endpoint

### Step 2: Create Railway Project

1. Go to https://railway.app/new
2. Click **"Deploy from GitHub repo"**
3. Select your GitHub account and repository
4. Choose the **`experiflow-proxy`** directory as root

**OR use Railway CLI:**
```bash
cd experiflow-proxy
railway login
railway init
railway up
```

### Step 3: Configure Environment Variables

In Railway Dashboard → Variables, set:

**Required:**
```bash
ORIGIN_URL=https://www.yoursite.com
EXPERIFLOW_API_URL=https://api.experiflow.com
EXPERIMENT_IDS=your-experiment-id-here
```

**Recommended:**
```bash
FAIL_OPEN=true
TRANSFORM_TIMEOUT=50ms
ENABLE_LOGGING=true
```

**Optional:**
```bash
EXPERIFLOW_EDGE_TOKEN=your-token-here  # If using API auth
PORT=8090  # Railway auto-sets this
```

### Step 4: Deploy

Railway will automatically:
1. Build Docker image using Dockerfile
2. Deploy the proxy
3. Assign a public URL (e.g., `your-proxy.up.railway.app`)
4. Start health checks at `/health`

**Deployment time:** 2-5 minutes

### Step 5: Verify Deployment

Test the health endpoint:
```bash
curl https://your-proxy.up.railway.app/health
```

Expected response:
```json
{"status":"healthy","service":"experiflow-proxy"}
```

---

## Environment Variables

### Proxy Settings

| Variable | Required | Default | Description |
|----------|----------|---------|-------------|
| `PORT` | No | `8090` | Port to listen on (Railway sets automatically) |
| `ORIGIN_URL` | **Yes** | - | Your origin server URL |
| `READ_TIMEOUT` | No | `10s` | HTTP read timeout |
| `WRITE_TIMEOUT` | No | `10s` | HTTP write timeout |

### ExperiFlow API Settings

| Variable | Required | Default | Description |
|----------|----------|---------|-------------|
| `EXPERIFLOW_API_URL` | **Yes** | - | ExperiFlow API base URL |
| `EXPERIFLOW_EDGE_TOKEN` | No | - | Optional API authentication token |
| `TRANSFORM_TIMEOUT` | No | `50ms` | Timeout for transformations (keep low!) |

### Experiment Configuration

| Variable | Required | Default | Description |
|----------|----------|---------|-------------|
| `EXPERIMENT_IDS` | **Yes** | - | Comma-separated experiment IDs |

**Example:**
```bash
EXPERIMENT_IDS=54ce9030-4da3-4866-8b25-6d956207f325,another-exp-id
```

### Feature Flags

| Variable | Required | Default | Description |
|----------|----------|---------|-------------|
| `FAIL_OPEN` | No | `true` | Pass through on errors (CRITICAL for production) |
| `ENABLE_LOGGING` | No | `true` | Enable request logging |
| `ENABLE_METRICS` | No | `true` | Enable metrics collection |

---

## Testing & Verification

### 1. Test Health Endpoint

```bash
curl -I https://your-proxy.up.railway.app/health
```

Expected:
```
HTTP/2 200
Content-Type: application/json
{"status":"healthy","service":"experiflow-proxy"}
```

### 2. Test Proxy with Origin

```bash
curl -i https://your-proxy.up.railway.app/
```

Check response headers for:
```
X-EF-Experiment: [your-experiment-id]
X-EF-Variant: [Control|Variant Name]
X-EF-Transform: hit|control|miss|timeout
X-EF-Timing: total=35ms
```

### 3. Test in Browser

1. Open proxy URL in browser
2. Check Network tab → Response Headers
3. Verify experiment headers are present
4. Check for cookie: `ef_variant_[experiment-id]`

### 4. Verify Transformations

1. Create an experiment with visual changes in ExperiFlow
2. Set experiment status to "Running"
3. Visit proxy URL
4. HTML should be transformed (check source code)
5. No visual flicker should occur

---

## Custom Domain Setup

### Option 1: Point Entire Domain to Proxy

If you want all traffic to go through the proxy:

1. In Railway: Settings → Domains → Add Domain
2. Enter your domain: `www.yoursite.com`
3. Update DNS with CNAME:
   ```
   Type:  CNAME
   Name:  www
   Value: [railway-provided-cname]
   TTL:   Auto
   ```

### Option 2: Use Subdomain for Testing

For gradual rollout, use a test subdomain:

1. In Railway: Settings → Domains → Add Domain
2. Enter: `test.yoursite.com`
3. Update DNS:
   ```
   Type:  CNAME
   Name:  test
   Value: [railway-provided-cname]
   TTL:   Auto
   ```
4. Test traffic goes through proxy
5. Later, switch main domain

### Option 3: CDN in Front (Advanced)

For global performance:
```
User → CDN (Cloudflare/Fastly) → Proxy → Origin
```

1. Point CDN to proxy URL
2. Configure CDN to respect `Vary: Cookie` header
3. Set CDN to pass through `ef_variant_*` cookies

---

## Monitoring & Troubleshooting

### Railway Logs

View logs in Railway Dashboard → Deployments → Logs

**Key log patterns:**
```
[ExperiFlow Proxy] Starting on port 8090
[ExperiFlow Proxy] Origin: https://www.yoursite.com
[ExperiFlow Proxy] Active experiments: [exp-id]
[ExperiFlow Proxy] Assigned user to variant: Variant Name
[ExperiFlow Proxy] Applied 2 transformations (took 28ms)
```

### Response Headers for Debugging

Every proxied response includes:

| Header | Values | Meaning |
|--------|--------|---------|
| `X-EF-Experiment` | UUID | Active experiment ID |
| `X-EF-Variant` | String | Assigned variant name |
| `X-EF-Transform` | `hit` | Transformations applied successfully |
|  | `control` | User in control (no transforms) |
|  | `miss` | No experiment matched |
|  | `timeout` | Transform timed out (failed open) |
| `X-EF-Timing` | `total=35ms` | Time taken for transformations |

### Common Issues

#### 1. Proxy Returns 502 Bad Gateway

**Symptom:** HTTP 502 error

**Causes & Solutions:**
- **Origin unreachable:** Verify `ORIGIN_URL` is correct and publicly accessible
- **Origin timeout:** Increase `READ_TIMEOUT` and `WRITE_TIMEOUT`
- **Origin SSL issues:** Ensure origin has valid SSL certificate

**Check:**
```bash
curl -I $ORIGIN_URL  # Should return 200
```

#### 2. No Transformations Applied

**Symptom:** `X-EF-Transform: miss` or `control`

**Causes & Solutions:**
- **Experiment not active:** Check experiment status in dashboard (should be "Running")
- **Wrong experiment ID:** Verify `EXPERIMENT_IDS` matches dashboard
- **No visual changes:** Ensure experiment has published visual changes
- **User in control:** 50% of users see control by default (this is expected)

**Debug:**
```bash
# Check API endpoint
curl https://api.experiflow.com/experiments/transform-spec/[exp-id]
```

#### 3. Transformations Timing Out

**Symptom:** `X-EF-Transform: timeout`, slow responses

**Causes & Solutions:**
- **API latency:** Check API health at `https://api.experiflow.com/health`
- **Complex HTML:** Large pages take longer to parse
- **Timeout too low:** Increase `TRANSFORM_TIMEOUT` (but keep < 100ms)

**Monitor:**
```bash
# Check timing header
curl -I https://your-proxy.up.railway.app/ | grep X-EF-Timing
```

#### 4. High Memory Usage

**Symptom:** Railway shows high memory usage, possible OOM kills

**Causes & Solutions:**
- **Too many concurrent requests:** Scale horizontally (add more instances)
- **Memory leak:** Check logs for abnormal patterns
- **Large responses:** HTML parsing uses memory proportional to page size

**Scale up:**
- Railway: Increase memory in Settings
- Or deploy additional instances

---

## Performance & Scaling

### Single Instance Performance

| Metric | Value |
|--------|-------|
| **RPS** | 10,000+ |
| **p95 Latency** | < 50ms added |
| **Memory** | ~50-100MB |
| **CPU** | ~100m (0.1 core) idle, ~500m (0.5 core) under load |

### Horizontal Scaling

The proxy is **stateless** - scale by adding instances:

**Railway:**
1. Settings → Replicas → Increase count
2. Railway load balances automatically

**Kubernetes:**
```yaml
replicas: 5  # Scale to 5 instances
```

**Load Balancer:**
```
     ┌─────────────┐
     │ Load Balancer│
     └──────┬───────┘
            │
    ┌───────┼───────┐
    │       │       │
 Proxy-1 Proxy-2 Proxy-3
    │       │       │
    └───────┴───────┘
          │
       Origin
```

### Recommended Scaling Thresholds

| Traffic | Instances | Memory per Instance |
|---------|-----------|---------------------|
| < 1,000 RPS | 1 | 128MB |
| 1,000 - 5,000 RPS | 2-3 | 256MB |
| 5,000 - 10,000 RPS | 5+ | 256MB |
| > 10,000 RPS | 10+ | 512MB |

### Caching Considerations

**What's cached:**
- Transform specs from API (in-memory, per instance)
- HTTP responses respect origin cache headers

**What's NOT cached:**
- User variant assignments (stored in cookies)
- HTML content (always fresh from origin)

**Future:** Redis caching for transform specs across instances

---

## Security Considerations

### Data Privacy ✅

- **Customer HTML never leaves infrastructure** - only transform specs sent to API
- **No user data stored** - only cookies for variant assignment
- **HTTPS enforced** - Railway provides automatic SSL
- **HttpOnly cookies** - variant cookies not accessible via JavaScript
- **SameSite protection** - CSRF protection on cookies

### API Authentication (Optional)

Set `EXPERIFLOW_EDGE_TOKEN` to require authentication:

```bash
EXPERIFLOW_EDGE_TOKEN=your-secure-token-here
```

The proxy will send:
```
Authorization: Bearer your-secure-token-here
```

### Network Security

**Recommended:**
- Deploy proxy in same region as origin (lower latency)
- Use private networking if possible (origin not publicly accessible)
- Enable Railway's DDoS protection
- Monitor unusual traffic patterns

### Compliance

- **SOC2/HIPAA:** Proxy can run in customer infrastructure
- **GDPR:** No PII stored, only anonymous variant cookies
- **PCI:** Does not handle payment data

---

## Production Checklist

Before going live:

### Configuration
- [ ] `FAIL_OPEN=true` is set (CRITICAL)
- [ ] `ORIGIN_URL` points to production site
- [ ] `EXPERIFLOW_API_URL` points to production API (https://api.experiflow.com)
- [ ] `EXPERIMENT_IDS` contains valid, running experiments
- [ ] `TRANSFORM_TIMEOUT` set to reasonable value (50ms recommended)

### Testing
- [ ] Health endpoint returns 200: `/health`
- [ ] Proxy successfully fetches from origin
- [ ] Experiment headers present: `X-EF-*`
- [ ] Transformations applied: `X-EF-Transform: hit`
- [ ] Variant cookie set: `ef_variant_[exp-id]`
- [ ] No visual flicker in browser
- [ ] Latency acceptable (< 100ms added)

### Deployment
- [ ] Deployed to production environment
- [ ] SSL certificate active (HTTPS working)
- [ ] Custom domain configured (if applicable)
- [ ] Multiple instances deployed (for redundancy)
- [ ] Load balancer configured (if needed)

### Monitoring
- [ ] Logs accessible and readable
- [ ] Response headers monitored
- [ ] Error rate alerts configured
- [ ] Latency alerts configured
- [ ] Memory/CPU usage monitored

### Documentation
- [ ] Team knows how to view logs
- [ ] Rollback procedure documented
- [ ] On-call contact list updated
- [ ] Experiment IDs documented

---

## Rollback Procedures

### Quick Rollback

If issues occur, immediately:

**Option 1: Disable Experiments**
```bash
# Set EXPERIMENT_IDS to empty
EXPERIMENT_IDS=
# Restart service
```

**Option 2: Point Domain Back to Origin**
Update DNS CNAME to point directly to origin (bypassing proxy)

**Option 3: Scale Down**
Railway: Settings → Replicas → Set to 0

### Gradual Rollback

1. Reduce traffic: 100% → 50% → 0%
2. Monitor metrics at each step
3. Identify and fix issues
4. Re-enable gradually

---

## Cost Breakdown

### Railway (Recommended)

| Resource | Usage | Cost |
|----------|-------|------|
| **Compute** | 1 instance | $5/month |
| **Bandwidth** | 100GB included | $0.10/GB over |
| **SSL** | Automatic | Free |
| **Custom Domain** | Unlimited | Free |

**Total:** ~$5-15/month for small-medium traffic

### Other Platforms

- **AWS ECS:** ~$15-50/month
- **Google Cloud Run:** ~$10-30/month
- **VPS (DigitalOcean):** $5-20/month

### Scaling Costs

At 10,000 RPS:
- 5 instances × $5 = $25/month (Railway)
- Plus bandwidth costs

---

## Support & Resources

### Documentation
- **Setup Guide:** [SETUP.md](SETUP.md)
- **README:** [README.md](README.md)
- **Architecture:** [SUMMARY.md](SUMMARY.md)

### ExperiFlow Resources
- **API Documentation:** https://api.experiflow.com/docs
- **Dashboard:** https://app.experiflow.com
- **Backend Deployment:** [experiflow-api/PRODUCTION_DEPLOYMENT.md](../experiflow-api/PRODUCTION_DEPLOYMENT.md)

### Railway Resources
- **Documentation:** https://docs.railway.app
- **Status:** https://railway.app/status
- **Community:** https://discord.gg/railway

---

## Next Steps

1. ✅ **Deploy to Railway** - Follow steps above
2. ⏳ **Configure environment variables** - Set origin URL and experiment IDs
3. ⏳ **Test deployment** - Verify health endpoint and transformations
4. ⏳ **Set up custom domain** - Point your domain to the proxy
5. ⏳ **Monitor performance** - Check logs and response headers
6. ⏳ **Create documentation** - Document your specific setup

---

**Deployment Status:** Ready for Production
**Platform:** Railway (or Docker-compatible)
**Service:** ExperiFlow Proxy (Server-Side A/B Testing)
**Maintainer:** ExperiFlow Team
