# Attaches a server to a vSwitch.
resource "hetzner_vswitch_server" "example" {
  vswitch_id    = hetzner_vswitch.example.id
  server_number = 12345
}
