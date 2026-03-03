# Manages traffic warning settings for a subnet.
resource "hetzner_subnet" "example" {
  ip                    = "203.0.113.0"
  mask                  = 24
  traffic_warnings      = true
  traffic_warning_limit = 10000
}
