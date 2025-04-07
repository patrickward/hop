# Hop Version Management Guide

This document explains how we manage different versions of the Hop package and how you can use them in your projects. It's mostly a note to myself. 

## Version Structure

Hop uses semantic versioning and follows Go's module versioning conventions. We maintain multiple versions through git branches:

- **v0.x-v1.x (pre-v1 or v1)**: Maintained in the `v1` branch
- **v2.x**: Maintained in the `main` branch

## Using Different Versions

### Using v0.x/v1.x (Legacy Version)

To use the pre-v2 version of Hop in your project:

```go
import "github.com/patrickward/hop"
```

In your `go.mod` file:

```
require github.com/patrickward/hop v0.5.0 // Or any other v0.x/v1.x tag
```

### Using v2.x (Current Version)

To use the current v2 version of Hop:

```go
import "github.com/patrickward/hop/v2"
```

In your `go.mod` file:

```
require github.com/patrickward/hop/v2 v2.0.0 // Or any other v2.x tag
```

## Key Differences Between Versions

> ⚠️ **Breaking Changes in v2**:
>
> Detailed information about breaking changes will be listed here.

The main difference between v1 and v2 is the elimination of the `render`, `request`, `route`, and `pulse` packages. The `view` package 
has replaced the `render` package and the `request` and `route` packages have been removed as they weren't specific to the package and really should be within individual projects. The `pulse` package has been removed as it should either be in a separate package or within the project itself. 

## Contributing

### Bug Fixes for v0.x/v1.x

If you need to make bug fixes to the legacy v0.x version:

1. Check out the v1 branch:
   ```bash
   git checkout v1
   ```

2. Make your changes and commit them:
   ```bash
   git commit -m "Fix: description of the bug fix"
   ```

3. Push your changes to the v1 branch:
   ```bash
   git push origin v1
   ```

### Developing v2.x

For ongoing development of v2:

1. Check out the main branch:
   ```bash
   git checkout main
   ```

2. Make your changes and commit them:
   ```bash
   git commit -m "Feature: description of the new feature"
   ```

3. Push your changes to the main branch:
   ```bash
   git push origin main
   ```

## Release Process

We use Git tags to mark releases for both version streams:

### Tagging v0.x/v1.x Releases

```bash
git checkout v1
# Make sure your changes are committed
git tag v0.5.1  # Use appropriate version number
git push origin v0.5.1
```

### Tagging v2.x Releases

```bash
git checkout main
# Make sure your changes are committed
git tag v2.0.1  # Use appropriate version number
git push origin v2.0.1
```

## Migration Guide

If you're upgrading from v0.x/v1.x to v2.x, please follow these steps:

1. Update your import paths:
   ```go
   // Old
   import "github.com/patrickward/hop"
   
   // New
   import "github.com/patrickward/hop/v2"
   ```

2. Update your go.mod file:
   ```
   require github.com/patrickward/hop/v2 v2.0.0
   ```

3. Adjust your code to accommodate the breaking changes (detailed above).

## Support Policy

- **v0.x/v1.x**: Bug fixes only, no new features
- **v2.x**: Active development, new features, and bug fixes
