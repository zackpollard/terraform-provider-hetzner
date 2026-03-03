resource "hetzner_rdns" "example" {
  ip  = "203.0.113.1"
  ptr = "server.example.com"
}
