# Manages settings for an existing dedicated server.
resource "hetzner_server" "example" {
  server_number = 12345
  server_name   = "my-server"
}
