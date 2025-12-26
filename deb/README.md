# Athens Debian Package with Multi-Org GitHub PAT Support

This directory contains the configuration and scripts to build Athens as a Debian package with support for multiple GitHub Personal Access Tokens (PATs) across different organizations.

## Features

- **Multi-Org GitHub PAT Support**: Configure different GitHub tokens for different organizations
- **Systemd Integration**: Service management with systemd
- **Secure Defaults**: Runs as dedicated `athens` user with restricted permissions
- **Easy Configuration**: Simple TOML configuration file

## Building the Package

### Using the Build Script

```bash
# Build with default version (0.0.1) and architecture (amd64)
./scripts/build-deb.sh

# Build with specific version
./scripts/build-deb.sh 1.0.0

# Build for specific architecture
./scripts/build-deb.sh 1.0.0 arm64
```

### Using GitHub Actions

The package is automatically built on:
- Push to `main` branch
- Tagged releases (e.g., `v1.0.0`)
- Manual workflow dispatch

To trigger a manual build:
1. Go to Actions tab
2. Select "Build Debian Package" workflow
3. Click "Run workflow"
4. Optionally specify a version

## Installation

### From .deb File

```bash
sudo dpkg -i athens-multipat_VERSION_ARCH.deb
sudo apt-get install -f  # Install dependencies if needed
```

### Configuration

1. Edit the configuration file:
```bash
sudo nano /etc/athens/config.toml
```

2. Add your Git authentication for private repos (GitHub, GitLab, etc.):
```toml
# GitHub
[[GitAuths]]
  host = "github.com"
  org = "mycompany"
  token = "ghp_xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx"

# GitLab
[[GitAuths]]
  host = "gitlab.com"
  org = "my-group"
  token = "glpat-xxxxxxxxxxxxx"
```

3. Start the service:
```bash
sudo systemctl start athens
sudo systemctl enable athens
```

4. Check the status:
```bash
sudo systemctl status athens
```

## Package Structure

```
/usr/local/bin/
  ├── athens                      # Main Athens proxy binary
  └── git-credential-athens       # Git credential helper

/etc/athens/
  └── config.toml                 # Configuration file

/var/lib/athens/                  # Data directory
/var/log/athens/                  # Log directory

/lib/systemd/system/
  └── athens.service              # Systemd service file
```

## Usage

### Basic Usage

```bash
# Start the service
sudo systemctl start athens

# Stop the service
sudo systemctl stop athens

# Restart the service
sudo systemctl restart athens

# View logs
sudo journalctl -u athens -f
```

### Using Athens Proxy

Configure your Go environment to use Athens:

```bash
export GOPROXY=http://localhost:3000
export GOPRIVATE=github.com/yourorg/*
```

Or configure in your `~/.gitconfig`:

```ini
[url "http://localhost:3000"]
    insteadOf = https://proxy.golang.org
```

## Uninstallation

### Remove but keep configuration and data
```bash
sudo apt-get remove athens-multipat
```

### Complete removal including data
```bash
sudo apt-get purge athens-multipat
```

## Troubleshooting

### Check if service is running
```bash
sudo systemctl status athens
```

### View logs
```bash
sudo journalctl -u athens -n 100
```

### Verify configuration
```bash
sudo cat /etc/athens/config.toml
```

### Check permissions
```bash
ls -la /var/lib/athens
ls -la /etc/athens
id athens
```

## Security Notes

- The Athens service runs as a dedicated `athens` user with restricted permissions
- Configuration files containing tokens should have restricted permissions (600 or 640)
- Store tokens securely and rotate them regularly
- Use fine-grained GitHub tokens with minimal required permissions

## Development

### Building Locally

```bash
# Install dependencies
sudo apt-get install dpkg-dev

# Build the package
./scripts/build-deb.sh 0.0.1 amd64

# The package will be in build/deb/
```

### Testing

```bash
# Install the built package
sudo dpkg -i build/deb/athens-multipat_0.0.1_amd64.deb

# Test it
sudo systemctl start athens
curl http://localhost:3000/healthz
```

## Contributing

See [CONTRIBUTING.md](../CONTRIBUTING.md) for general contribution guidelines.

## License

See [LICENSE](../LICENSE) for license information.
