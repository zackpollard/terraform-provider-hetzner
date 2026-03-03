resource "hetzner_firewall_template" "example" {
  name          = "web-server"
  filter_ipv6   = false
  allowlist_hos = true
  is_default    = false

  input = [{
    name       = "Allow SSH"
    ip_version = "ipv4"
    dst_port   = "22"
    protocol   = "tcp"
    action     = "accept"
    }, {
    name       = "Allow HTTPS"
    ip_version = "ipv4"
    dst_port   = "443"
    protocol   = "tcp"
    action     = "accept"
  }]
}
