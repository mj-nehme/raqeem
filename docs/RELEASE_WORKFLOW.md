# 🧪 Pre-Release Testing & Release Workflow

## The Complete Workflow

### Step 1: Test Your Environment (BEFORE tagging)

```bash
# Start environment with current code
./start.sh

# Wait for everything to come up (~2 minutes)
# Then manually test:

# 1. Open Mentor Dashboard
open http://localhost:15000

# 2. Open Device Simulator  
open http://localhost:14000

# 3. Test creating a device
# 4. Test sending data
# 5. Test viewing data in dashboard
```

**Manual checks:**
- ✅ Can you see the device list?
- ✅ Can you register a new device?
- ✅ Can you send metrics/activities?
- ✅ Does data appear in the dashboard?
- ✅ Do screenshots upload/display?

**Automated checks:**
```bash
# Run automated tests
./scripts/health-check.sh
```

This will verify:
- ✅ All pods running
- ✅ Backends responding
- ✅ Frontends accessible
- ✅ Environment variables set
- ✅ Service discovery working

---

### Step 2: Stop Environment

```bash
./stop.sh
```

---

### Step 3: Create Release (ONLY if tests passed)

```bash
# This will:
# - Validate code
# - Build images with v1.0.0 tags
# - Push to GitHub Container Registry (GHCR)
# - Update Helm charts
# - Create git tag
./scripts/tag-release.sh v1.0.0
```

**What happens:**
```
📦 Docker Images Tagged:
  • ghcr.io/mj-nehme/raqeem/devices-backend:v1.0.0
  • ghcr.io/mj-nehme/raqeem/devices-backend:v1.0.0-86e6e44
  • ghcr.io/mj-nehme/raqeem/devices-backend:latest

  • ghcr.io/mj-nehme/raqeem/mentor-backend:v1.0.0
  • ghcr.io/mj-nehme/raqeem/mentor-backend:v1.0.0-86e6e44
  • ghcr.io/mj-nehme/raqeem/mentor-backend:latest

🏷️  Git Tag: v1.0.0
```

---

### Step 4: Push to GitHub

```bash
# Push the commit and tag
git push origin v1.0.0
git push
```

---

### Step 5: Test the Release

```bash
# Deploy using the versioned release
echo "IMAGE_TAG=v1.0.0" > .deploy/tag.env
./start.sh

# Verify it works with the v1.0.0 images
```

---

## Understanding Image Tags

When you run `./scripts/tag-release.sh v1.0.0`, it creates **ONE image** but gives it **THREE tags**:

```bash
# These all point to THE SAME IMAGE:
ghcr.io/mj-nehme/raqeem/devices-backend:v1.0.0           # Semantic version
ghcr.io/mj-nehme/raqeem/devices-backend:v1.0.0-86e6e44   # Version + git SHA
ghcr.io/mj-nehme/raqeem/devices-backend:latest           # Latest tag
```

**Analogy:** Like having three bookmarks pointing to the same webpage.

**Which to use?**
- **Production**: Use `v1.0.0` (specific, predictable)
- **Development**: Use `latest` (always newest)
- **Debugging**: Use `v1.0.0-86e6e44` (know exact commit)

---

## Current State

### What's Built Right Now
```bash
# Check local images
docker images | grep raqeem

# You have:
ghcr.io/mj-nehme/raqeem/devices-backend:86e6e44    # Git commit SHA
ghcr.io/mj-nehme/raqeem/devices-backend:latest     # Latest tag
ghcr.io/mj-nehme/raqeem/mentor-backend:86e6e44  # Git commit SHA  
ghcr.io/mj-nehme/raqeem/mentor-backend:latest   # Latest tag
```

### What Will Be Built When You Run tag-release.sh
```bash
# Will ADD these tags (same images, new tags):
ghcr.io/mj-nehme/raqeem/devices-backend:v1.0.0
ghcr.io/mj-nehme/raqeem/devices-backend:v1.0.0-86e6e44
# (latest gets updated)

ghcr.io/mj-nehme/raqeem/mentor-backend:v1.0.0
ghcr.io/mj-nehme/raqeem/mentor-backend:v1.0.0-86e6e44
# (latest gets updated)
```

---

## FAQ

### Q: Do I need to rebuild images before tagging?
**A:** No! The tag-release.sh script builds them for you.

### Q: What if the images don't work?
**A:** That's why we test FIRST with `./start.sh` before running tag-release.sh!

### Q: Can I undo a release?
**A:** Yes! Git tags can be deleted:
```bash
git tag -d v1.0.0
git push origin :refs/tags/v1.0.0
```

But Docker images can't be untagged easily, so test first!

### Q: What's the difference between `86e6e44` and `latest` tags?
**A:** 
- `86e6e44` - Git commit SHA (specific code version)
- `latest` - Always points to newest build (changes)

### Q: Should I use `latest` in production?
**A:** NO! Use `v1.0.0` in production. `latest` changes, versions don't.

---

## Recommended Workflow (Summary)

```bash
# 1. Test current code
./start.sh
# ... manually test everything ...
# ... run automated tests ...
./stop.sh

# 2. If all tests pass, create release
./scripts/tag-release.sh v1.0.0

# 3. Push to GitHub
git push origin v1.0.0
git push

# 4. Verify release works
echo "IMAGE_TAG=v1.0.0" > .deploy/tag.env
./start.sh
# ... test again ...

# 5. Done! You have a stable v1.0.0 release
```

---

## Container Registry (Updated!)

**Before (v0.1.x - DockerHub):**
```
jaafarn/raqeem-devices-backend   ⚠️ Deprecated
jaafarn/raqeem-mentor-backend    ⚠️ Deprecated
```

**After (v0.2.0+ - GitHub Container Registry):**
```
ghcr.io/mj-nehme/raqeem/devices-backend   ✅ Current
ghcr.io/mj-nehme/raqeem/mentor-backend    ✅ Current
```

All images are now hosted on GitHub Container Registry (GHCR) for better integration, security, and reliability.

See [GHCR Migration Guide](GHCR_MIGRATION.md) for details.

---

**Next Step:** Run `./start.sh` and test everything before creating v1.0.0!
