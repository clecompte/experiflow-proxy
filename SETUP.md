# ExperiFlow Proxy - Setup Guide

## Prerequisites

1. **Install Go** (1.21 or later)
   ```bash
   # macOS
   brew install go

   # Linux
   wget https://go.dev/dl/go1.21.5.linux-amd64.tar.gz
   sudo tar -C /usr/local -xzf go1.21.5.linux-amd64.tar.gz
   export PATH=$PATH:/usr/local/go/bin

   # Windows
   # Download installer from https://go.dev/dl/
   ```

2. **Verify Go installation**
   ```bash
   go version
   # Should output: go version go1.21.x ...
   ```

## Quick Start (3 Steps)

### Step 1: Install Dependencies

```bash
cd experiflow-proxy
go mod download
```

### Step 2: Start Required Services

You need three services running:

**Terminal 1 - Origin Server (your test page)**:
```bash
cd /path/to/experiflow
python -m http.server 8080
```

**Terminal 2 - ExperiFlow API**:
```bash
cd experiflow-api
uvicorn main:app --reload --port 8000
```

### Step 3: Start the Proxy

**Terminal 3 - ExperiFlow Proxy**:
```bash
cd experiflow-proxy
./start.sh
```

Or manually:
```bash
export ORIGIN_URL=http://localhost:8080
export EXPERIFLOW_API_URL=http://localhost:8000
export EXPERIMENT_IDS=54ce9030-4da3-4866-8b25-6d956207f325

go build -o proxy ./cmd/proxy
./proxy
```

## Test It!

Open your browser to:
```
http://localhost:8090/test-page.html
```

**Expected behavior**:
- ✅ **NO visual flicker!**
- ✅ Button already shows "Try It Free Today!" (variant)
- ✅ OR button shows "Get Started Now" (control)
- ✅ No flash of original content

Check response headers:
```bash
curl -I http://localhost:8090/test-page.html

# Look for:
X-EF-Experiment: 54ce9030-4da3-4866-8b25-6d956207f325
X-EF-Variant: Green CTA Button Variant
X-EF-Transform: hit
X-EF-Timing: total=35ms
```

## Verify Zero Flicker

1. **Clear cookies** in your browser
2. **Open DevTools** → Network tab
3. **Reload** the page
4. **Watch the button** - it should render with the variant text immediately, no change!

Compare this to the client-side tracker at `localhost:8080` where you'll see flicker.

## Troubleshooting

### "go: command not found"
Install Go (see Prerequisites above)

### "Failed to fetch variants"
- Check ExperiFlow API is running on port 8000
- Check experiment ID is correct
- Check experiment status is "running"

### "Proxy error"
- Check origin server is running on port 8080
- Check the URL in browser matches your test page

### "No transformations applied"
- Check experiment has published visual changes
- Check variant assignment in cookies (DevTools → Application → Cookies)
- Check `X-EF-Transform` header (should be `hit` not `control`)

## Next Steps

1. **Test variant assignment**: Clear cookies and reload multiple times - you should see both control and variant (50/50 split)

2. **Check performance**: Look at `X-EF-Timing` header - should be < 50ms

3. **Production deployment**: See [README.md](README.md) for Docker, Kubernetes, and cloud deployment guides

## Configuration

All configuration is via environment variables. See [README.md](README.md#configuration) for full list.

**Most important**:
- `ORIGIN_URL` - Your site URL
- `EXPERIFLOW_API_URL` - ExperiFlow API URL
- `EXPERIMENT_IDS` - Comma-separated experiment IDs

## Architecture

```
Browser Request
      ↓
localhost:8090 (Proxy)
      ↓
   [Assign Variant] → localhost:8000 (ExperiFlow API)
      ↓
   [Fetch HTML] → localhost:8080 (Origin)
      ↓
   [Transform HTML]
      ↓
Browser receives transformed HTML (NO FLICKER!)
```

## Support

Questions? Check:
- [README.md](README.md) - Full documentation
- [Implementation Plan](../docs/server-side-ab-testing-implementation-plan.md) - Architecture details
