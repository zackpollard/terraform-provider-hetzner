# Hetzner Robot API Specification

> Comprehensive reference for building Terraform resources/data sources against the Hetzner Robot Webservice API.
>
> **Base URL:** `https://robot-ws.your-server.de`
> **Authentication:** HTTP Basic Auth (username + password)
> **Content-Type (requests):** `application/x-www-form-urlencoded`
> **Response Format:** JSON (default), YAML (append `.yaml` to path)
> **Error Format:** `{ "error": { "status": <int>, "code": "<string>", "message": "<string>" } }`
> Invalid input errors also include `missing` and `invalid` arrays of parameter names.

---

## Table of Contents

1. [Server](#1-server)
2. [IP Addresses](#2-ip-addresses)
3. [Subnets](#3-subnets)
4. [Reset](#4-reset)
5. [Failover](#5-failover)
6. [Wake on LAN](#6-wake-on-lan)
7. [Boot Configuration](#7-boot-configuration)
8. [Reverse DNS (rDNS)](#8-reverse-dns-rdns)
9. [Traffic](#9-traffic)
10. [SSH Keys](#10-ssh-keys)
11. [Server Ordering](#11-server-ordering)
12. [Storage Box](#12-storage-box)
13. [Firewall](#13-firewall)
14. [vSwitch](#14-vswitch)
15. [Terraform Mapping Recommendations](#15-terraform-mapping-recommendations)

---

## 1. Server

Servers are the core resource. They cannot be created/destroyed via API (only ordered/cancelled).

### GET /server

Query all servers. **Rate limit:** 200/hr

| Request Param | Type   | Required | Description              |
|---------------|--------|----------|--------------------------|
| `server_ip`   | string | no       | Filter by server main IP |

**Response (array of):**

| Field             | Type   | Description                           |
|-------------------|--------|---------------------------------------|
| `server_ip`       | string | Main IPv4 address                     |
| `server_ipv6_net` | string | Main IPv6 network                     |
| `server_number`   | int    | Unique server ID (primary identifier) |
| `server_name`     | string | User-assigned name                    |
| `product`         | string | Product name (e.g. "DS 3000")         |
| `dc`              | string | Data center (e.g. "FSN1-DC14")        |
| `traffic`         | string | Free traffic quota (e.g. "5 TB")      |
| `status`          | string | `"ready"` or `"in process"`           |
| `cancelled`       | bool   | Whether server has been cancelled     |
| `paid_until`      | string | Date paid until (yyyy-MM-dd)          |
| `ip`              | array  | Array of assigned single IPs          |
| `subnet`          | array  | Array of assigned subnets             |

### GET /server/{server-number}

Query single server. **Rate limit:** 200/hr

Additional fields beyond list response:

| Field               | Type | Description                    |
|---------------------|------|--------------------------------|
| `reset`             | bool | Reset system available         |
| `rescue`            | bool | Rescue system available        |
| `vnc`               | bool | VNC installation available     |
| `windows`           | bool | Windows installation available |
| `plesk`             | bool | Plesk available                |
| `cpanel`            | bool | cPanel available               |
| `wol`               | bool | Wake on LAN available          |
| `hot_swap`          | bool | Hot swap capable               |
| `linked_storagebox` | int  | Linked storage box ID          |

### POST /server/{server-number}

Update server name. **Rate limit:** 200/hr

| Request Param | Type   | Required | Description     |
|---------------|--------|----------|-----------------|
| `server_name` | string | yes      | New server name |

### GET /server/{server-number}/cancellation

Query cancellation status. **Rate limit:** 200/hr

| Response Field               | Type         | Description                |
|------------------------------|--------------|----------------------------|
| `earliest_cancellation_date` | string       | yyyy-MM-dd                 |
| `cancelled`                  | bool         | Is cancelled               |
| `reservation_possible`       | bool         | Can reserve location       |
| `reserved`                   | bool         | Location will be reserved  |
| `cancellation_date`          | string/null  | Active cancellation date   |
| `cancellation_reason`        | array/string | Possible or active reasons |

### POST /server/{server-number}/cancellation

Cancel a server. **Rate limit:** 200/hr

| Request Param         | Type   | Required    | Description                  |
|-----------------------|--------|-------------|------------------------------|
| `cancellation_date`   | string | yes         | Date (yyyy-MM-dd) or `"now"` |
| `cancellation_reason` | string | no          | Reason                       |
| `reserve_location`    | string | conditional | `"true"` or `"false"`        |

### DELETE /server/{server-number}/cancellation

Withdraw cancellation. **Rate limit:** 200/hr. No response body.

**Terraform:** Data source `hetzner_server` (read-only). Resource `hetzner_server` for managing `server_name` only.
Import by `server_number`.

---

## 2. IP Addresses

Single IP addresses assigned to servers. IPs are purchased/cancelled separately from servers.

### GET /ip

Query all IPs. **Rate limit:** 5000/hr

| Request Param | Type   | Required | Description              |
|---------------|--------|----------|--------------------------|
| `server_ip`   | string | no       | Filter by server main IP |

**Response (array of):**

| Field              | Type        | Description                  |
|--------------------|-------------|------------------------------|
| `ip`               | string      | IP address                   |
| `server_ip`        | string      | Server main IP               |
| `server_number`    | int         | Server ID                    |
| `locked`           | bool        | Locking status               |
| `separate_mac`     | string/null | Separate MAC address or null |
| `traffic_warnings` | bool        | Traffic warnings enabled     |
| `traffic_hourly`   | int         | Hourly limit in MB           |
| `traffic_daily`    | int         | Daily limit in MB            |
| `traffic_monthly`  | int         | Monthly limit in GB          |

### GET /ip/{ip}

Query single IP. **Rate limit:** 5000/hr

Additional fields:

| Field       | Type   | Description       |
|-------------|--------|-------------------|
| `gateway`   | string | Gateway address   |
| `mask`      | int    | CIDR notation     |
| `broadcast` | string | Broadcast address |

### POST /ip/{ip}

Update traffic warnings. **Rate limit:** 5000/hr

| Request Param      | Type | Required | Description             |
|--------------------|------|----------|-------------------------|
| `traffic_warnings` | bool | no       | Enable/disable warnings |
| `traffic_hourly`   | int  | no       | Hourly limit in MB      |
| `traffic_daily`    | int  | no       | Daily limit in MB       |
| `traffic_monthly`  | int  | no       | Monthly limit in GB     |

### GET /ip/{ip}/mac

Query separate MAC address. **Rate limit:** 5000/hr

| Response Field | Type   | Description |
|----------------|--------|-------------|
| `ip`           | string | IP address  |
| `mac`          | string | MAC address |

### PUT /ip/{ip}/mac

Generate separate MAC. **Rate limit:** 10/hr. No parameters.

### DELETE /ip/{ip}/mac

Remove separate MAC. **Rate limit:** 10/hr. Returns MAC object with null value.

### GET /ip/{ip}/cancellation

Query IP cancellation data. **Rate limit:** 200/hr

| Response Field               | Type        | Description         |
|------------------------------|-------------|---------------------|
| `ip`                         | string      | IP address          |
| `server_number`              | string      | Server ID           |
| `earliest_cancellation_date` | string      | yyyy-MM-dd          |
| `cancelled`                  | bool        | Cancellation status |
| `cancellation_date`          | string/null | Scheduled date      |

### POST /ip/{ip}/cancellation

Cancel an IP. **Rate limit:** 200/hr

| Request Param       | Type   | Required | Description                  |
|---------------------|--------|----------|------------------------------|
| `cancellation_date` | string | yes      | Date (yyyy-MM-dd) or `"now"` |

### DELETE /ip/{ip}/cancellation

Revoke cancellation. **Rate limit:** 200/hr

**Terraform:** Data source `hetzner_ip` (read individual IP details). Resource `hetzner_ip` to manage traffic warnings
and separate MAC. Import by IP address string.

---

## 3. Subnets

Subnets assigned to servers. Similar to IPs but with subnet-specific fields.

### GET /subnet

Query all subnets. **Rate limit:** 5000/hr

| Request Param | Type   | Required | Description         |
|---------------|--------|----------|---------------------|
| `server_ip`   | string | no       | Filter by server IP |

**Response (array of):**

| Field              | Type   | Description          |
|--------------------|--------|----------------------|
| `ip`               | string | Subnet IP            |
| `mask`             | int    | CIDR notation        |
| `gateway`          | string | Subnet gateway       |
| `server_ip`        | string | Server main IP       |
| `server_number`    | int    | Server ID            |
| `failover`         | bool   | Is a failover subnet |
| `locked`           | bool   | Locking status       |
| `traffic_warnings` | bool   | Warnings enabled     |
| `traffic_hourly`   | int    | Hourly limit MB      |
| `traffic_daily`    | int    | Daily limit MB       |
| `traffic_monthly`  | int    | Monthly limit GB     |

### GET /subnet/{net-ip}

Query single subnet. **Rate limit:** 5000/hr. Returns subnet object.

### POST /subnet/{net-ip}

Update traffic warnings. **Rate limit:** 5000/hr

| Request Param      | Type | Required | Description      |
|--------------------|------|----------|------------------|
| `traffic_warnings` | bool | no       | Enable/disable   |
| `traffic_hourly`   | int  | no       | Hourly limit MB  |
| `traffic_daily`    | int  | no       | Daily limit MB   |
| `traffic_monthly`  | int  | no       | Monthly limit GB |

### GET /subnet/{net-ip}/mac

Query separate MAC. **Rate limit:** 5000/hr

| Response Field | Type   | Description                            |
|----------------|--------|----------------------------------------|
| `ip`           | string | Subnet IP                              |
| `mask`         | string | CIDR                                   |
| `mac`          | string | MAC address                            |
| `possible_mac` | object | Available MAC addresses to choose from |

### PUT /subnet/{net-ip}/mac

Set separate MAC. **Rate limit:** 10/hr

| Request Param | Type   | Required | Description                            |
|---------------|--------|----------|----------------------------------------|
| `mac`         | string | yes      | Target MAC address (from possible_mac) |

### DELETE /subnet/{net-ip}/mac

Remove separate MAC. **Rate limit:** 10/hr

### GET /subnet/{net-ip}/cancellation

Query cancellation. **Rate limit:** 200/hr

### POST /subnet/{net-ip}/cancellation

Cancel subnet. **Rate limit:** 200/hr

| Request Param       | Type   | Required | Description                  |
|---------------------|--------|----------|------------------------------|
| `cancellation_date` | string | yes      | Date (yyyy-MM-dd) or `"now"` |

### DELETE /subnet/{net-ip}/cancellation

Revoke cancellation. **Rate limit:** 200/hr

**Terraform:** Data source `hetzner_subnet`. Resource `hetzner_subnet` for traffic warnings and MAC. Import by subnet
IP.

---

## 4. Reset

Execute hardware resets on servers. This is an action, not a persistent resource.

### GET /reset

Query all servers' reset options. **Rate limit:** 500/hr

| Response Field    | Type   | Description           |
|-------------------|--------|-----------------------|
| `server_ip`       | string | Server main IP        |
| `server_ipv6_net` | string | Server IPv6 net       |
| `server_number`   | int    | Server ID             |
| `type`            | array  | Available reset types |

### GET /reset/{server-number}

Query reset options for one server. **Rate limit:** 500/hr

Additional field:

| Field              | Type   | Description                     |
|--------------------|--------|---------------------------------|
| `operating_status` | string | Current server operating status |

### POST /reset/{server-number}

Execute reset. **Rate limit:** 50/hr

| Request Param | Type   | Required | Description                       |
|---------------|--------|----------|-----------------------------------|
| `type`        | string | yes      | Reset type (from available types) |

Reset types include: `sw` (software), `hw` (hardware), `man` (manual).

**Terraform:** Not suitable as a resource (it is a one-shot action). Could be a provisioner or null_resource trigger,
but generally skip for Terraform.

---

## 5. Failover

Failover IPs can be routed between servers. Key for HA setups.

### GET /failover

Query all failover IPs. **Rate limit:** 100/hr

**Response (array of):**

| Field              | Type   | Description                    |
|--------------------|--------|--------------------------------|
| `ip`               | string | Failover IP/subnet address     |
| `netmask`          | string | Failover netmask               |
| `server_ip`        | string | Server main IP (owner)         |
| `server_ipv6_net`  | string | Server IPv6                    |
| `server_number`    | int    | Server ID (owner)              |
| `active_server_ip` | string | Current routing destination IP |

### GET /failover/{failover-ip}

Query single failover. **Rate limit:** 100/hr

### POST /failover/{failover-ip}

Switch failover routing. **Rate limit:** 50/hr

| Request Param      | Type   | Required | Description                       |
|--------------------|--------|----------|-----------------------------------|
| `active_server_ip` | string | yes      | Destination server IP to route to |

### DELETE /failover/{failover-ip}

Delete failover routing (set active_server_ip to null). **Rate limit:** 50/hr

**Terraform:** Resource `hetzner_failover` to manage `active_server_ip` routing. Import by failover IP. Data source
`hetzner_failover` to read current routing.

---

## 6. Wake on LAN

Send Wake on LAN packets. One-shot action.

### GET /wol/{server-number}

Query WoL status. **Rate limit:** 500/hr

| Response Field    | Type   | Description |
|-------------------|--------|-------------|
| `server_ip`       | string | Server IP   |
| `server_ipv6_net` | string | Server IPv6 |
| `server_number`   | int    | Server ID   |

### POST /wol/{server-number}

Send WoL packet. **Rate limit:** 10/hr. No parameters.

**Terraform:** Skip - one-shot action, not suitable for Terraform state.

---

## 7. Boot Configuration

Boot configuration supports multiple installation types: rescue, linux, vnc, windows.

### GET /boot/{server-number}

Query all boot configurations. **Rate limit:** 500/hr

Returns nested objects for each boot type (rescue, linux, vnc, windows) with their respective fields.

---

### 7.1 Rescue System

#### GET /boot/{server-number}/rescue

Query rescue options. **Rate limit:** 500/hr

| Response Field    | Type         | Description                                               |
|-------------------|--------------|-----------------------------------------------------------|
| `server_ip`       | string       | Server IP                                                 |
| `server_ipv6_net` | string       | Server IPv6                                               |
| `server_number`   | int          | Server ID                                                 |
| `os`              | array/string | Available OS options when inactive; active OS when active |
| `arch`            | array/int    | Available architectures (64, 32 @deprecated)              |
| `active`          | bool         | Whether rescue is currently active                        |
| `password`        | string/null  | Generated password (only on activation)                   |
| `authorized_key`  | array        | SSH key fingerprints                                      |
| `host_key`        | array        | Host key fingerprints                                     |
| `keyboard`        | string       | Keyboard layout                                           |

#### POST /boot/{server-number}/rescue

Activate rescue system. **Rate limit:** 500/hr

| Request Param    | Type         | Required | Description                     |
|------------------|--------------|----------|---------------------------------|
| `os`             | string       | yes      | OS: `"linux"` or `"vkvm"`       |
| `arch`           | int          | no       | Architecture, default 64        |
| `authorized_key` | string/array | no       | SSH key fingerprint(s)          |
| `keyboard`       | string       | no       | Keyboard layout, default `"us"` |

**Response:** Rescue object with generated `password`.

#### DELETE /boot/{server-number}/rescue

Deactivate rescue. **Rate limit:** 500/hr

#### GET /boot/{server-number}/rescue/last

Last rescue activation data. **Rate limit:** 500/hr

---

### 7.2 Linux Installation

#### GET /boot/{server-number}/linux

Query linux install options. **Rate limit:** 500/hr

| Response Field    | Type         | Description                       |
|-------------------|--------------|-----------------------------------|
| `server_ip`       | string       | Server IP                         |
| `server_ipv6_net` | string       | Server IPv6                       |
| `server_number`   | int          | Server ID                         |
| `dist`            | array/string | Available or active distributions |
| `lang`            | array/string | Available or active languages     |
| `arch`            | array/int    | Available architectures           |
| `active`          | bool         | Active status                     |
| `password`        | string/null  | Generated password                |
| `authorized_key`  | array        | SSH key fingerprints              |
| `host_key`        | array        | Host key fingerprints             |

#### POST /boot/{server-number}/linux

Activate linux install. **Rate limit:** 500/hr

| Request Param    | Type         | Required | Description              |
|------------------|--------------|----------|--------------------------|
| `dist`           | string       | yes      | Distribution             |
| `lang`           | string       | yes      | Language                 |
| `arch`           | int          | no       | Architecture, default 64 |
| `authorized_key` | string/array | no       | SSH key fingerprint(s)   |

**Response:** Linux object with generated `password`.

#### DELETE /boot/{server-number}/linux

Deactivate. **Rate limit:** 500/hr

#### GET /boot/{server-number}/linux/last

Last activation data. **Rate limit:** 500/hr

---

### 7.3 VNC Installation

#### GET /boot/{server-number}/vnc

Query VNC options. **Rate limit:** 500/hr

| Response Field    | Type         | Description                       |
|-------------------|--------------|-----------------------------------|
| `server_ip`       | string       | Server IP                         |
| `server_ipv6_net` | string       | Server IPv6                       |
| `server_number`   | int          | Server ID                         |
| `dist`            | array/string | Available or active distributions |
| `lang`            | array/string | Available or active languages     |
| `arch`            | array/int    | Available architectures           |
| `active`          | bool         | Active status                     |
| `password`        | string/null  | Generated password                |

#### POST /boot/{server-number}/vnc

Activate VNC install. **Rate limit:** 500/hr

| Request Param | Type   | Required | Description              |
|---------------|--------|----------|--------------------------|
| `dist`        | string | yes      | Distribution             |
| `lang`        | string | yes      | Language                 |
| `arch`        | int    | no       | Architecture, default 64 |

#### DELETE /boot/{server-number}/vnc

Deactivate. **Rate limit:** 500/hr

---

### 7.4 Windows Installation

#### GET /boot/{server-number}/windows

Query Windows options. **Rate limit:** 500/hr

| Response Field    | Type         | Description                          |
|-------------------|--------------|--------------------------------------|
| `server_ip`       | string       | Server IP                            |
| `server_ipv6_net` | string       | Server IPv6                          |
| `server_number`   | int          | Server ID                            |
| `dist`            | array/string | Available or active Windows versions |
| `lang`            | array/string | Available or active languages        |
| `active`          | bool         | Active status                        |
| `password`        | string/null  | Generated password                   |

#### POST /boot/{server-number}/windows

Activate Windows install. **Rate limit:** 500/hr

| Request Param | Type   | Required | Description     |
|---------------|--------|----------|-----------------|
| `dist`        | string | yes      | Windows version |
| `lang`        | string | yes      | Language        |

#### DELETE /boot/{server-number}/windows

Deactivate. **Rate limit:** 500/hr

---

**Terraform:** Resource `hetzner_boot_rescue` for rescue system activation. Data source `hetzner_boot` for querying
available options. The linux/vnc/windows installations are typically one-shot and may not map well as persistent
resources, but rescue is commonly toggled.

---

## 8. Reverse DNS (rDNS)

Manage PTR records for IP addresses.

### GET /rdns

Query all rDNS entries. **Rate limit:** 500/hr

| Request Param | Type   | Required | Description         |
|---------------|--------|----------|---------------------|
| `server_ip`   | string | no       | Filter by server IP |

**Response (array of):**

| Field | Type   | Description      |
|-------|--------|------------------|
| `ip`  | string | IP address       |
| `ptr` | string | PTR record value |

### GET /rdns/{ip}

Query single rDNS entry. **Rate limit:** 500/hr

### PUT /rdns/{ip}

Create new rDNS entry. **Rate limit:** 500/hr. Returns 201 Created.

| Request Param | Type   | Required | Description      |
|---------------|--------|----------|------------------|
| `ptr`         | string | yes      | PTR record value |

### POST /rdns/{ip}

Create or update rDNS entry. **Rate limit:** 500/hr. Returns 201 on create, 200 on update.

| Request Param | Type   | Required | Description      |
|---------------|--------|----------|------------------|
| `ptr`         | string | yes      | PTR record value |

### DELETE /rdns/{ip}

Delete rDNS entry. **Rate limit:** 500/hr. No response body.

**Terraform:** Resource `hetzner_rdns`. ID is the IP address. Import by IP address string. CRUD maps directly: Create =
PUT, Read = GET, Update = POST, Delete = DELETE.

---

## 9. Traffic

Query traffic statistics. Read-only, time-bounded queries.

### POST /traffic

Query traffic data. **Rate limit:** 200/hr

| Request Param   | Type   | Required | Description                     |
|-----------------|--------|----------|---------------------------------|
| `ip[]`          | array  | no       | One or more IP addresses        |
| `subnet[]`      | array  | no       | One or more subnets             |
| `from`          | string | yes      | Start date/time                 |
| `to`            | string | yes      | End date/time                   |
| `type`          | string | yes      | `"day"`, `"month"`, or `"year"` |
| `single_values` | string | no       | `"true"` for grouped data       |

Date formats by type:

- Day: `YYYY-MM-DDTHH`
- Month: `YYYY-MM-DD`
- Year: `YYYY-MM`

**Response:**

| Field           | Type   | Description         |
|-----------------|--------|---------------------|
| `type`          | string | Query type          |
| `from`          | string | Start date          |
| `to`            | string | End date            |
| `data`          | object | Keyed by IP/subnet  |
| `data.<ip>.in`  | float  | Inbound traffic GB  |
| `data.<ip>.out` | float  | Outbound traffic GB |
| `data.<ip>.sum` | float  | Total traffic GB    |

**Terraform:** Data source `hetzner_traffic` for querying. Not a resource (read-only metrics).

---

## 10. SSH Keys

Manage SSH public keys stored in the Robot account.

### GET /key

List all SSH keys. **Rate limit:** 500/hr

**Response (array of):**

| Field         | Type   | Description                                |
|---------------|--------|--------------------------------------------|
| `name`        | string | Key name                                   |
| `fingerprint` | string | Key fingerprint (unique ID)                |
| `type`        | string | Algorithm: `"RSA"`, `"ECDSA"`, `"ED25519"` |
| `size`        | int    | Key size in bits                           |
| `data`        | string | Public key in OpenSSH format               |
| `created_at`  | string | Creation date                              |

### POST /key

Create SSH key. **Rate limit:** 200/hr. Returns 201 Created.

| Request Param | Type   | Required | Description                         |
|---------------|--------|----------|-------------------------------------|
| `name`        | string | yes      | Key name                            |
| `data`        | string | yes      | Public key (OpenSSH or SSH2 format) |

### GET /key/{fingerprint}

Query single key. **Rate limit:** 500/hr

### POST /key/{fingerprint}

Update key name. **Rate limit:** 200/hr

| Request Param | Type   | Required | Description  |
|---------------|--------|----------|--------------|
| `name`        | string | yes      | New key name |

### DELETE /key/{fingerprint}

Delete key. **Rate limit:** 200/hr. No response body.

**Terraform:** Resource `hetzner_ssh_key`. ID = fingerprint. Import by fingerprint string. Full CRUD: Create = POST
/key, Read = GET /key/{fp}, Update = POST /key/{fp}, Delete = DELETE /key/{fp}. Note: `data` (the key material) is
immutable after creation - changing it requires destroy+recreate.

---

## 11. Server Ordering

Order new servers, marketplace servers, and addons. **These are financial transactions.**

### 11.1 New Server Products

#### GET /order/server/product

List available server products.

#### GET /order/server/product/{product-id}

Query specific product details.

### 11.2 New Server Transactions

#### GET /order/server/transaction

List server order transactions.

#### POST /order/server/transaction

Create a new server order.

Likely parameters (inferred from context):

- `product_id` - Server product to order
- `authorized_key` - SSH key fingerprints for installation
- `password` - Installation password
- `dist` - Linux distribution
- `lang` - Language
- `location` - Preferred data center
- `addon` - Additional services
- `test` - Test mode flag

#### GET /order/server/transaction/{id}

Query transaction status.

### 11.3 Server Market (Used Servers)

#### GET /order/server_market/product

List marketplace products.

#### GET /order/server_market/product/{product-id}

Query marketplace product details.

#### POST /order/server_market/transaction

Purchase marketplace server.

#### GET /order/server_market/transaction/{id}

Query transaction status.

### 11.4 Server Addons

#### GET /order/server_addon/{server-number}/product

List available addons for a server.

#### POST /order/server_addon/transaction

Order an addon.

#### GET /order/server_addon/transaction/{id}

Query addon transaction status.

### 11.5 Currency

#### GET /order/currency

List supported currencies.

**Terraform:** Generally skip for Terraform. Ordering involves financial transactions and long provisioning times that
don't fit the Terraform lifecycle well. Data sources `hetzner_server_product` and `hetzner_server_market_product` could
be useful for reference. Actual ordering should remain manual.

---

## 12. Storage Box

Manage Hetzner Storage Boxes (NAS/backup storage).

### GET /storagebox

List all storage boxes. **Rate limit:** 500/hr

**Response (array of):**

| Field             | Type     | Description                  |
|-------------------|----------|------------------------------|
| `storagebox_id`   | int      | Storage box ID               |
| `storagebox_name` | string   | Name                         |
| `disk_quota`      | int      | Total capacity GB            |
| `disk_usage`      | int      | Used space GB                |
| `status`          | string   | Operational status           |
| `paid_until`      | string   | Expiration date (yyyy-MM-dd) |
| `locked`          | bool     | Access restricted            |
| `server`          | int/null | Linked server ID             |

### GET /storagebox/{storagebox-id}

Query single storage box. **Rate limit:** 500/hr

Additional fields:

| Field                   | Type | Description           |
|-------------------------|------|-----------------------|
| `webdav`                | bool | WebDAV enabled        |
| `samba`                 | bool | SMB/CIFS enabled      |
| `ssh`                   | bool | SSH access enabled    |
| `external_reachability` | bool | Remote access enabled |
| `zfs`                   | bool | ZFS features enabled  |

### POST /storagebox/{storagebox-id}

Update settings. **Rate limit:** 500/hr

| Request Param           | Type   | Required | Description                  |
|-------------------------|--------|----------|------------------------------|
| `storagebox_name`       | string | no       | New name                     |
| `webdav`                | bool   | no       | Enable/disable WebDAV        |
| `samba`                 | bool   | no       | Enable/disable SMB           |
| `ssh`                   | bool   | no       | Enable/disable SSH           |
| `external_reachability` | bool   | no       | Enable/disable remote access |
| `zfs`                   | bool   | no       | Enable/disable ZFS           |

Error codes: `INVALID_INPUT`, `STORAGEBOX_NOT_FOUND`, `UPDATE_FAILED`

### POST /storagebox/{storagebox-id}/password

Reset password. **Rate limit:** 200/hr

---

### 12.1 Snapshots

#### GET /storagebox/{storagebox-id}/snapshot

List snapshots. **Rate limit:** 500/hr

| Response Field | Type        | Description              |
|----------------|-------------|--------------------------|
| `name`         | string      | Snapshot name/identifier |
| `timestamp`    | datetime    | Creation time            |
| `comment`      | string/null | Comment                  |
| `size`         | int         | Size in GB               |

#### POST /storagebox/{storagebox-id}/snapshot

Create snapshot. **Rate limit:** 500/hr

| Request Param | Type   | Required | Description      |
|---------------|--------|----------|------------------|
| `comment`     | string | no       | Snapshot comment |

#### DELETE /storagebox/{storagebox-id}/snapshot/{snapshot-name}

Delete snapshot. **Rate limit:** 10/hr

#### POST /storagebox/{storagebox-id}/snapshot/{snapshot-name}

Restore from snapshot. **Rate limit:** 10/hr

#### POST /storagebox/{storagebox-id}/snapshot/{snapshot-name}/comment

Update snapshot comment. **Rate limit:** 500/hr

| Request Param | Type   | Required | Description      |
|---------------|--------|----------|------------------|
| `comment`     | string | yes      | New comment text |

---

### 12.2 Snapshot Plans

#### GET /storagebox/{storagebox-id}/snapshotplan

Query snapshot schedule. **Rate limit:** 500/hr

| Response Field | Type   | Description                 |
|----------------|--------|-----------------------------|
| `status`       | string | `"enabled"` or `"disabled"` |
| `minute`       | int    | Minute (0-59)               |
| `hour`         | int    | Hour (0-23)                 |
| `day_of_week`  | int    | Day of week (0=Sun, 6=Sat)  |
| `day_of_month` | int    | Day of month (1-31)         |
| `month`        | int    | Month (1-12)                |

#### POST /storagebox/{storagebox-id}/snapshotplan

Update schedule. **Rate limit:** 500/hr

| Request Param  | Type   | Required | Description                 |
|----------------|--------|----------|-----------------------------|
| `status`       | string | no       | `"enabled"` or `"disabled"` |
| `minute`       | int    | no       | 0-59                        |
| `hour`         | int    | no       | 0-23                        |
| `day_of_week`  | int    | no       | 0-6                         |
| `day_of_month` | int    | no       | 1-31                        |
| `month`        | int    | no       | 1-12                        |

---

### 12.3 Sub-accounts

#### GET /storagebox/{storagebox-id}/subaccount

List sub-accounts. **Rate limit:** 500/hr

| Response Field          | Type        | Description            |
|-------------------------|-------------|------------------------|
| `username`              | string      | Login identifier       |
| `homedirectory`         | string      | Home directory path    |
| `samba`                 | bool        | SMB enabled            |
| `webdav`                | bool        | WebDAV enabled         |
| `ssh`                   | bool        | SSH enabled            |
| `external_reachability` | bool        | Remote access          |
| `readonly`              | bool        | Read-only mode         |
| `createdir`             | bool        | Can create directories |
| `comment`               | string/null | Comment                |

#### POST /storagebox/{storagebox-id}/subaccount

Create sub-account. **Rate limit:** 200/hr. Returns 201.

| Request Param           | Type   | Required | Description              |
|-------------------------|--------|----------|--------------------------|
| `username`              | string | yes      | Unique username          |
| `homedirectory`         | string | yes      | Home directory path      |
| `samba`                 | bool   | no       | Enable SMB               |
| `webdav`                | bool   | no       | Enable WebDAV            |
| `ssh`                   | bool   | no       | Enable SSH               |
| `external_reachability` | bool   | no       | Allow remote access      |
| `readonly`              | bool   | no       | Read-only restriction    |
| `createdir`             | bool   | no       | Allow directory creation |
| `comment`               | string | no       | Comment                  |

Error codes: `INVALID_INPUT`, `STORAGEBOX_NOT_FOUND`, `SUBACCOUNT_ALREADY_EXISTS`, `SUBACCOUNT_CREATE_FAILED`

#### PUT /storagebox/{storagebox-id}/subaccount/{username}

Update sub-account. **Rate limit:** 500/hr. Same parameters as create (all optional).

#### DELETE /storagebox/{storagebox-id}/subaccount/{username}

Delete sub-account. **Rate limit:** 10/hr

#### POST /storagebox/{storagebox-id}/subaccount/{username}/password

Reset sub-account password. **Rate limit:** 200/hr

**Terraform:**

- Resource `hetzner_storagebox` for managing settings (name, webdav, samba, ssh, etc.). Import by `storagebox_id`. Note:
  storage boxes are purchased externally; Terraform manages configuration only.
- Resource `hetzner_storagebox_subaccount` for managing sub-accounts. Import by `storagebox_id/username`.
- Resource `hetzner_storagebox_snapshot_plan` for managing snapshot schedules. Import by `storagebox_id`.
- Data source `hetzner_storagebox` and `hetzner_storagebox_snapshot`.

---

## 13. Firewall

Manage per-server firewall rules. Each server has its own firewall with input (incoming) and output (outgoing) rule
sets.

### GET /firewall/{server-number}

Query firewall configuration. **Rate limit:** not specified

**Response:**

| Field           | Type   | Description                                               |
|-----------------|--------|-----------------------------------------------------------|
| `server_ip`     | string | Server main IP                                            |
| `server_number` | int    | Server ID                                                 |
| `status`        | string | `"active"`, `"disabled"`, or `"in process"`               |
| `allowlist_hos` | bool   | Allow Hetzner services (rescue, DHCP, DNS, monitoring)    |
| `filter_ipv6`   | bool   | Whether firewall also filters IPv6 (IPv4 always filtered) |
| `port`          | string | Switch port: `"main"` or `"kvm"`                          |
| `rules`         | object | Contains `input` and `output` arrays                      |
| `rules.input`   | array  | Incoming traffic rules                                    |
| `rules.output`  | array  | Outgoing traffic rules                                    |

**Firewall Rule Object:**

| Field        | Type   | Description                                                                       |
|--------------|--------|-----------------------------------------------------------------------------------|
| `ip_version` | string | `"ipv4"`, `"ipv6"`, or empty (both)                                               |
| `name`       | string | Rule name. Allowed chars: `a-z A-Z . - + _ @` and space                           |
| `dst_ip`     | string | Destination IP/subnet in CIDR notation (e.g. `"1.2.3.4/32"`)                      |
| `src_ip`     | string | Source IP/subnet in CIDR notation                                                 |
| `dst_port`   | string | Destination port or range (e.g. `"443"` or `"32768-65535"`)                       |
| `src_port`   | string | Source port or range                                                              |
| `protocol`   | string | Protocol: `"tcp"`, `"udp"`, `"icmp"`, etc.                                        |
| `tcp_flags`  | string | TCP flags: `syn`, `fin`, `rst`, `psh`, `urg`; combine with `\|` (OR) or `&` (AND) |
| `action`     | string | `"accept"` or `"discard"`                                                         |

**Constraints:**

- Maximum 10 rules per direction (input/output).
- Rules are applied in order (top to bottom).
- When `ip_version` is omitted (applies to both IPv4 and IPv6), you **cannot** specify `dst_ip`, `src_ip`, or
  `protocol`.
- `protocol` requires `ip_version` to be set.
- `dst_ip`/`src_ip` require `ip_version` to be set.
- If output rules are left empty while firewall is active, the server cannot send outgoing packets.

### POST /firewall/{server-number}

Set firewall configuration. **Rate limit:** not specified

The POST replaces the entire firewall configuration. Parameters are submitted as form-encoded with array notation:

| Request Param                 | Type   | Required       | Description                  |
|-------------------------------|--------|----------------|------------------------------|
| `status`                      | string | yes            | `"active"` or `"disabled"`   |
| `allowlist_hos`               | bool   | no             | Allow Hetzner services       |
| `filter_ipv6`                 | bool   | no             | Filter IPv6 traffic          |
| `rules[input][N][ip_version]` | string | no             | Rule IP version              |
| `rules[input][N][name]`       | string | no             | Rule name                    |
| `rules[input][N][dst_ip]`     | string | no             | Destination IP/CIDR          |
| `rules[input][N][src_ip]`     | string | no             | Source IP/CIDR               |
| `rules[input][N][dst_port]`   | string | no             | Destination port(s)          |
| `rules[input][N][src_port]`   | string | no             | Source port(s)               |
| `rules[input][N][protocol]`   | string | no             | Protocol                     |
| `rules[input][N][tcp_flags]`  | string | no             | TCP flags                    |
| `rules[input][N][action]`     | string | yes (per rule) | `"accept"` or `"discard"`    |
| `rules[output][N][...]`       | string | no             | Same fields for output rules |

Where `N` is the 0-based rule index.

### DELETE /firewall/{server-number}

Clear all firewall rules (resets to default). **Rate limit:** not specified

---

### 13.1 Firewall Templates

Reusable firewall configurations.

#### GET /firewall/template

List all templates. **Rate limit:** not specified

**Response (array of):**

| Field           | Type   | Description                      |
|-----------------|--------|----------------------------------|
| `id`            | int    | Template ID                      |
| `name`          | string | Template name                    |
| `filter_ipv6`   | bool   | Filter IPv6                      |
| `allowlist_hos` | bool   | Allow Hetzner services           |
| `is_default`    | bool   | Is default template              |
| `rules`         | object | Same structure as firewall rules |

#### POST /firewall/template

Create template. **Rate limit:** not specified

| Request Param           | Type   | Required | Description                           |
|-------------------------|--------|----------|---------------------------------------|
| `name`                  | string | yes      | Template name                         |
| `filter_ipv6`           | bool   | no       | Filter IPv6                           |
| `allowlist_hos`         | bool   | no       | Allow Hetzner services                |
| `is_default`            | bool   | no       | Set as default                        |
| `rules[input][N][...]`  |        | no       | Input rules (same format as firewall) |
| `rules[output][N][...]` |        | no       | Output rules                          |

#### GET /firewall/template/{template-id}

Query single template. **Rate limit:** not specified

#### POST /firewall/template/{template-id}

Update template. **Rate limit:** not specified. Same parameters as create.

#### DELETE /firewall/template/{template-id}

Delete template. **Rate limit:** not specified

**Terraform:**

- Resource `hetzner_firewall` manages firewall config per server. ID = `server_number`. Import by server number. The
  entire config (status, rules, allowlist) is managed as one resource since the API replaces everything on POST.
- Resource `hetzner_firewall_template` manages templates. ID = template `id`. Full CRUD.
- Data source `hetzner_firewall` and `hetzner_firewall_template` for reading.

---

## 14. vSwitch

Virtual switches for Layer 2 networking between dedicated servers. Supports VLAN tagging.

### GET /vswitch

List all vSwitches. **Rate limit:** not specified

**Response (array of):**

| Field       | Type   | Description         |
|-------------|--------|---------------------|
| `id`        | int    | vSwitch ID          |
| `name`      | string | vSwitch name        |
| `vlan`      | int    | VLAN ID (4000-4091) |
| `cancelled` | bool   | Cancellation status |

### POST /vswitch

Create vSwitch. **Rate limit:** not specified

| Request Param | Type   | Required | Description                |
|---------------|--------|----------|----------------------------|
| `name`        | string | yes      | vSwitch name               |
| `vlan`        | int    | yes      | VLAN ID (range: 4000-4091) |

**Response:** Created vSwitch object with `id`.

### GET /vswitch/{vswitch-id}

Query single vSwitch with full details. **Rate limit:** not specified

**Response:**

| Field           | Type   | Description              |
|-----------------|--------|--------------------------|
| `id`            | int    | vSwitch ID               |
| `name`          | string | vSwitch name             |
| `vlan`          | int    | VLAN ID                  |
| `cancelled`     | bool   | Cancellation status      |
| `server`        | array  | Associated servers       |
| `subnet`        | array  | Associated subnets       |
| `cloud_network` | array  | Connected cloud networks |

**Server entry:**

| Field             | Type   | Description                              |
|-------------------|--------|------------------------------------------|
| `server_number`   | int    | Server ID                                |
| `server_ip`       | string | Server main IPv4                         |
| `server_ipv6_net` | string | Server IPv6 network                      |
| `status`          | string | `"ready"`, `"in process"`, or `"failed"` |

**Subnet entry:**

| Field     | Type   | Description   |
|-----------|--------|---------------|
| `ip`      | string | Subnet IP     |
| `mask`    | int    | CIDR notation |
| `gateway` | string | Gateway IP    |

**Cloud network entry:**

| Field     | Type   | Description      |
|-----------|--------|------------------|
| `id`      | int    | Cloud network ID |
| `ip`      | string | Network IP       |
| `mask`    | int    | CIDR notation    |
| `gateway` | string | Gateway IP       |

### POST /vswitch/{vswitch-id}

Update vSwitch. **Rate limit:** not specified

| Request Param | Type   | Required | Description             |
|---------------|--------|----------|-------------------------|
| `name`        | string | no       | New name                |
| `vlan`        | int    | no       | New VLAN ID (4000-4091) |

### DELETE /vswitch/{vswitch-id}

Cancel/delete vSwitch. **Rate limit:** not specified

| Request Param       | Type   | Required | Description                                        |
|---------------------|--------|----------|----------------------------------------------------|
| `cancellation_date` | string | no       | Date (yyyy-MM-dd) or `"now"` (default: end of day) |

### POST /vswitch/{vswitch-id}/server

Add server to vSwitch. **Rate limit:** not specified

| Request Param | Type | Required | Description          |
|---------------|------|----------|----------------------|
| `server`      | int  | yes      | Server number to add |

### DELETE /vswitch/{vswitch-id}/server

Remove server from vSwitch. **Rate limit:** not specified

| Request Param | Type | Required | Description             |
|---------------|------|----------|-------------------------|
| `server`      | int  | yes      | Server number to remove |

**Terraform:**

- Resource `hetzner_vswitch` for full lifecycle management. Create/Read/Update/Delete. ID = vSwitch `id`. Import by ID.
- The `server` list within the vSwitch can be managed inline (as a set attribute) or as a separate resource
  `hetzner_vswitch_server` for flexibility.
- Data source `hetzner_vswitch` for reading.

---

## 15. Terraform Mapping Recommendations

### Resources (full CRUD lifecycle)

| Terraform Resource                 | API Section                | ID Field               | Import Key    | Notes                                          |
|------------------------------------|----------------------------|------------------------|---------------|------------------------------------------------|
| `hetzner_ssh_key`                  | SSH Keys                   | fingerprint            | fingerprint   | Key data immutable after create                |
| `hetzner_rdns`                     | rDNS                       | ip                     | IP address    | Simple CRUD                                    |
| `hetzner_firewall`                 | Firewall                   | server_number          | server_number | Replaces entire config on update               |
| `hetzner_firewall_template`        | Firewall Templates         | id                     | template ID   | Full CRUD                                      |
| `hetzner_vswitch`                  | vSwitch                    | id                     | vSwitch ID    | Includes server membership                     |
| `hetzner_failover`                 | Failover                   | ip                     | failover IP   | Manages active_server_ip routing               |
| `hetzner_server`                   | Server                     | server_number          | server_number | Only manages server_name                       |
| `hetzner_storagebox`               | Storage Box                | storagebox_id          | storagebox ID | Manages settings only (externally provisioned) |
| `hetzner_storagebox_subaccount`    | Storage Box Sub-accounts   | storagebox_id/username | composite     | Full CRUD                                      |
| `hetzner_storagebox_snapshot_plan` | Storage Box Snapshot Plans | storagebox_id          | storagebox ID | Manages schedule                               |
| `hetzner_boot_rescue`              | Boot/Rescue                | server_number          | server_number | Activate/deactivate rescue                     |

### Data Sources (read-only)

| Terraform Data Source        | API Section        | Lookup Key                 | Notes                 |
|------------------------------|--------------------|----------------------------|-----------------------|
| `hetzner_server`             | Server             | server_number or server_ip | Read server details   |
| `hetzner_servers`            | Server             | (list all)                 | List/filter servers   |
| `hetzner_ssh_key`            | SSH Keys           | fingerprint                | Read key details      |
| `hetzner_ssh_keys`           | SSH Keys           | (list all)                 | List all keys         |
| `hetzner_ip`                 | IP                 | ip                         | Read IP details       |
| `hetzner_subnet`             | Subnet             | net-ip                     | Read subnet details   |
| `hetzner_failover`           | Failover           | failover-ip                | Read failover routing |
| `hetzner_firewall`           | Firewall           | server_number              | Read firewall config  |
| `hetzner_firewall_template`  | Firewall Templates | template id                | Read template         |
| `hetzner_firewall_templates` | Firewall Templates | (list all)                 | List templates        |
| `hetzner_vswitch`            | vSwitch            | vswitch id                 | Read vSwitch details  |
| `hetzner_vswitches`          | vSwitch            | (list all)                 | List vSwitches        |
| `hetzner_storagebox`         | Storage Box        | storagebox_id              | Read storage box      |
| `hetzner_traffic`            | Traffic            | ip/subnet + time range     | Query traffic stats   |

### Skip for Terraform (poor fit)

| Feature                       | Reason                                                              |
|-------------------------------|---------------------------------------------------------------------|
| Server Ordering               | Financial transactions with long provisioning; not idempotent       |
| Server Cancellation           | Destructive financial operation; should be manual                   |
| IP/Subnet Cancellation        | Destructive financial operation                                     |
| Reset                         | One-shot action, no persistent state                                |
| Wake on LAN                   | One-shot action, no persistent state                                |
| Linux/VNC/Windows Boot        | One-shot installations; rescue is the only boot type worth managing |
| Storage Box Snapshots         | Point-in-time backups; snapshot *plans* are the manageable resource |
| Storage Box Password Reset    | One-shot action                                                     |
| Traffic Warnings on IP/Subnet | Could be managed but low value; consider as optional enhancement    |

### Import Strategies

All resources support import via their natural identifiers:

- **Simple ID:** `hetzner_ssh_key` by fingerprint, `hetzner_rdns` by IP, `hetzner_vswitch` by numeric ID
- **Server-scoped:** `hetzner_firewall` by server_number, `hetzner_boot_rescue` by server_number
- **Composite:** `hetzner_storagebox_subaccount` by `storagebox_id/username`

### API Authentication Notes

- Base URL: `https://robot-ws.your-server.de`
- HTTP Basic Auth with Robot webservice credentials (not Hetzner Cloud API token)
- 3 failed auth attempts = 10-minute IP block (implement retry with backoff)
- Rate limits vary by endpoint (50-5000/hr); implement rate limiting in the API client
- All requests over HTTPS only

### Special Behaviors

1. **Firewall POST replaces everything:** The firewall API does not support adding individual rules. POST replaces the
   entire configuration. The Terraform resource must track all rules and send the complete state on every update.

2. **vSwitch server status is async:** Adding a server to a vSwitch may return `"in process"` status. The provider
   should poll until `"ready"` or `"failed"`.

3. **SSH key fingerprint is computed:** The fingerprint is derived from the key data. It cannot be specified by the user
   and serves as the natural ID.

4. **Storage boxes are externally provisioned:** Like servers, storage boxes are purchased through Hetzner and cannot be
   created/destroyed via API. Terraform manages configuration only.

5. **Boot rescue password is ephemeral:** The password is only returned when rescue is activated. Store it in state but
   mark as sensitive.

6. **Failover IPs are pre-provisioned:** Failover IPs must be purchased separately. The Terraform resource only manages
   routing (`active_server_ip`).

7. **Firewall rule ordering matters:** Rules are applied top-to-bottom. The Terraform resource should use an ordered
   list, not a set.

8. **vSwitch VLAN range:** VLAN IDs must be between 4000 and 4091 inclusive.
