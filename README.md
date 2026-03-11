# awkto-cli

CLI tool for managing Kea DHCP leases/reservations and DNS records via awkto APIs.

## Install

```bash
curl -fsSL https://gist.githubusercontent.com/awkto/166a3e3d96a8109005b804fcfdb489db/raw/install.sh | bash
```

Or download a specific version from [Releases](https://github.com/awkto/awkto-cli/releases).

## Configuration

### Quick Setup

Add and configure your servers using the CLI:

```bash
# Add a DNS server
awkto server add mydns --type dns --url https://dns.example.com --token your-dns-token

# Add a KEA DHCP server
awkto server add mykea --type kea --url https://kea.example.com:8080 --token your-kea-token --subnet-id 1

# Set default servers (optional, if you have multiple)
awkto server default dns mydns
awkto server default kea mykea
```

This creates a config file at `~/.awkto/config.yaml`. You can use a custom location by setting:

```bash
export AWKTO_CONFIG=/path/to/your/config.yaml
```

### Config File Format

The config file (`~/.awkto/config.yaml`) has this structure:

```yaml
defaults:
  dns: mydns
  kea: mykea
servers:
  mydns:
    type: dns
    url: https://dns.example.com
    token: your-dns-token
  mykea:
    type: kea
    url: https://kea.example.com:8080
    token: your-kea-token
    subnet_id: "1"
```

## Usage

### DNS Records

```bash
awkto dns list
awkto dns list -filter A
awkto dns create -name www -type A -values 192.168.1.1 -ttl 300
awkto dns edit -name www -type A -values 192.168.1.2
awkto dns delete -name www -type A
```

### DHCP Leases

```bash
awkto lease list
awkto lease delete -ip 10.33.11.50
awkto lease delete -mac 52:54:00:ab:cd:ef
awkto lease promote -ip 10.33.11.50 -hostname myserver
```

### DHCP Reservations

```bash
awkto reserve list
awkto reserve create -ip 10.33.11.50 -mac 52:54:00:ab:cd:ef -hostname myserver
awkto reserve edit -ip 10.33.11.50 -hostname newname
awkto reserve delete -ip 10.33.11.50
```

## Build from source

```bash
go build -o awkto .
```
