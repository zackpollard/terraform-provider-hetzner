# Routes a failover IP to a specific server.
resource "hetzner_failover" "example" {
  ip            = "203.0.113.100"
  active_server = "203.0.113.1"
}
