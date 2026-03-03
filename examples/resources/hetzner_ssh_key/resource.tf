resource "hetzner_ssh_key" "example" {
  name = "my-key"
  data = "ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAA..."
}
