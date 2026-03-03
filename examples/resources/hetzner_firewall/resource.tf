resource "hetzner_firewall" "example" {
  server_number = 12345
  status        = "active"
  allowlist_hos = true
  filter_ipv6   = false

  input = [{
    name       = "Allow SSH"
    ip_version = "ipv4"
    dst_port   = "22"
    protocol   = "tcp"
    action     = "accept"
  }]
}
