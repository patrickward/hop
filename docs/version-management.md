# Hop Version Management Guide

This document explains how we manage different versions of the Hop package and how you can use them in your projects. It's mostly a note to self. 

## Version Structure

Hop uses semantic versioning and follows Go's module versioning conventions. We maintain multiple versions through git branches:

- **v1.x**: Maintained in the `v1` branch
- **v2.x**: Maintained in the `main` branch

## Using Different Versions

### Using v1.x (Legacy Version)

To use the v1 version of Hop in your project:

```go
import "github.com/patrickward/hop"
```

In your `go.mod` file:

```
require github.com/patrickward/hop v1.0.0 // Or any other v1.x or v0.x tag
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

### Using Pre-release Versions of v2

To use a pre-release version of v2:

```go
import "github.com/patrickward/hop/v2"
```

In your `go.mod` file:

```
require github.com/patrickward/hop/v2 v2.0.0-alpha.1 // Or any other pre-release tag
```

You can also pin to the latest pre-release using Go's version query syntax:

```
// Get the latest alpha/beta/rc of v2
go get github.com/patrickward/hop/v2@v2.0.0-alpha
```

## Key Differences Between Versions

> ⚠️ **Breaking Changes in v2**:
>
> Detailed information about breaking changes will be listed here.

The main differences between v1 and v2 so far: 

- Replaced the `render` package with the `view` package 
- Removed the `route` package in place of using existing routers or just the standard library's http router
- Removed the `request` package as it had no real use case within the library; should handled on a per-project basis
- Removed the `pulse` packages it was a proof of concept and somewhat standalone 

## Contributing

### Bug Fixes for v1.x

If you need to make bug fixes to the legacy v1 version:

1. Check out the v1 branch:
   ```bash
   git checkout v1
   ```

2. Make your changes and commit them:
   ```bash
   git commit -m "Fix: description of the bug fix"
   ```

3. Push your changes to the v0.x branch:
   ```bash
   git push origin v0.x
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

### Tagging v1.x Releases

```bash
git checkout v1
# Make sure your changes are committed
git tag v1.0.1  # Use appropriate version number
git push origin v1.0.1
```

### Working with Specific Commits

Sometimes you may need to check out a specific commit rather than a branch:

```bash
# Check out a specific commit by its hash
git checkout a1b2c3d4

# Create a branch from a specific commit
git checkout -b bugfix-branch a1b2c3d4

# Tag a specific commit
git tag v1.0.2 a1b2c3d4
```

### Tagging v2.x Releases

#### Pre-release Versions

The Go community typically follows semantic versioning for pre-releases. The most common pattern is:

```bash
git checkout main
# Make sure your changes are committed

# Early development / unstable (alpha)
git tag v2.0.0-alpha.1  # First alpha release
git push origin v2.0.0-alpha.1
git tag v2.0.0-alpha.2  # Second alpha release
git push origin v2.0.0-alpha.2

# Feature complete, but with known bugs (beta)
git tag v2.0.0-beta.1  # First beta release 
git push origin v2.0.0-beta.1

# Feature and API frozen, bug fixes only (release candidate)
git tag v2.0.0-rc.1  # First release candidate
git push origin v2.0.0-rc.1
```

This pre-release versioning approach is widely used in the Go ecosystem and follows the semantic versioning specification. 

- Alpha (-alpha.X): Unstable, APIs may change
- Beta (-beta.X): Feature complete but may have bugs
- Release Candidate (-rc.X): Ready for testing, unlikely to change before final release


#### Stable Releases

Once v2 is considered stable:

```bash
git checkout main
# Make sure your changes are committed
git tag v2.0.0  # First stable release
git push origin v2.0.0

# For subsequent updates
git tag v2.0.1  # Patch updates
git push origin v2.0.1
```

## Migration Guide

If you're upgrading from v1.x to v2.x, please follow these steps:

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

3. Adjust your code to accommodate the breaking changes (detailed above). V1 code will likely have to be rewritten to work with V2, especially if you were using the `render` package. The `view` package also requires a different approach to rendering and data binding within templates. 

## Support Policy

- **v1.x**: Bug fixes only, no new features
- **v2.x**: Active development, new features, and bug fixes
