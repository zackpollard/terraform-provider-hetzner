# Terraform Provider for Hetzner Robot

A [Terraform](https://www.terraform.io) provider for managing dedicated server resources via
the [Hetzner Robot API](https://robot.your-server.de/doc/webservice/en.html).

> **Note:** This provider manages **dedicated server** infrastructure through the Hetzner Robot API.
> For Hetzner Cloud (VPS, load balancers, cloud storage boxes, etc.), use the
> [hcloud provider](https://registry.terraform.io/providers/hetznercloud/hcloud/latest).

## Requirements

- [Terraform](https://developer.hashicorp.com/terraform/downloads) >= 1.0
- [Go](https://golang.org/doc/install) >= 1.24 (for building from source)

## Authentication

The provider authenticates with the Hetzner Robot API using HTTP basic auth.
Create credentials at [robot.your-server.de](https://robot.your-server.de/) under Settings > Webservice.

```hcl
provider "hetzner" {
  username = var.robot_username
  password = var.robot_password
}
```

Or via environment variables:

```shell
export HETZNER_ROBOT_USERNAME="your-robot-username"
export HETZNER_ROBOT_PASSWORD="your-robot-password"
```

## Resources

| Resource | Description |
|----------|-------------|
| `hetzner_ssh_key` | SSH public keys for server access |
| `hetzner_rdns` | Reverse DNS (PTR) records for IPs |
| `hetzner_firewall` | Server firewall rules |
| `hetzner_firewall_template` | Reusable firewall rule templates |
| `hetzner_server` | Dedicated server settings (name, cancellation) |
| `hetzner_ip` | IP address traffic warning settings |
| `hetzner_subnet` | Subnet traffic warning settings |
| `hetzner_failover` | Failover IP routing |
| `hetzner_vswitch` | Virtual switches (vSwitch) |
| `hetzner_vswitch_server` | Server-to-vSwitch attachments |
| `hetzner_boot_rescue` | Rescue system boot configuration |
| `hetzner_boot_linux` | Linux installation boot configuration |
| `hetzner_boot_vnc` | VNC installation boot configuration |
| `hetzner_boot_windows` | Windows installation boot configuration |

## Data Sources

| Data Source | Description |
|-------------|-------------|
| `hetzner_ssh_key` | Read a single SSH key by fingerprint |
| `hetzner_ssh_keys` | List all SSH keys |
| `hetzner_rdns` | Read reverse DNS for an IP |
| `hetzner_firewall` | Read firewall config for a server |
| `hetzner_firewall_template` | Read a firewall template by ID |
| `hetzner_firewall_templates` | List all firewall templates |
| `hetzner_server` | Read a single server |
| `hetzner_servers` | List all servers |
| `hetzner_ip` | Read a single IP address |
| `hetzner_ips` | List all IP addresses |
| `hetzner_subnet` | Read a single subnet |
| `hetzner_subnets` | List all subnets |
| `hetzner_failover` | Read a failover IP |
| `hetzner_failovers` | List all failover IPs |
| `hetzner_vswitch` | Read a single vSwitch |
| `hetzner_vswitches` | List all vSwitches |
| `hetzner_boot_rescue` | Read rescue boot config |
| `hetzner_boot_linux` | Read Linux boot config |
| `hetzner_boot_vnc` | Read VNC boot config |
| `hetzner_boot_windows` | Read Windows boot config |
| `hetzner_reset` | Read server reset options |
| `hetzner_wol` | Read Wake-on-LAN status |

## Development

```shell
# Build
go build -v ./...

# Run unit tests
go test -v ./internal/provider/ -run TestUnit -count=1

# Run linter
golangci-lint run

# Generate documentation
make generate

# Run acceptance tests (requires credentials and real infrastructure)
export HETZNER_ROBOT_USERNAME="..."
export HETZNER_ROBOT_PASSWORD="..."
export HETZNER_TEST_SERVER_NUMBER="..."  # existing server number
make testacc
```
