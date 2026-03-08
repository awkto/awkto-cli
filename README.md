# awkto-cli

CLI tool for managing Kea DHCP leases/reservations and DNS records via awkto APIs.

## Install

```bash
curl -fsSL https://gist.githubusercontent.com/awkto/166a3e3d96a8109005b804fcfdb489db/raw/install.sh | bash
```

Or download a specific version from [Releases](https://github.com/awkto/awkto-cli/releases).

## Configuration

Set these environment variables (e.g. in `~/.bashrc` or `~/.zshrc`):

```bash
export AWKTO_KEA_URL="https://kea.example.com:8080"
export AWKTO_KEA_TOKEN="your-kea-token"
export AWKTO_DNS_URL="https://dns.example.com"
export AWKTO_DNS_TOKEN="your-dns-token"
export AWKTO_SUBNET_ID="1"  # optional, defaults to 1
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
