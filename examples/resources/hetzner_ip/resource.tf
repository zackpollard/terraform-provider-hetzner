# Manages traffic warning settings for an IP address.
resource "hetzner_ip" "example" {
  ip                    = "203.0.113.1"
  traffic_warnings      = true
  traffic_warning_limit = 5000
}
