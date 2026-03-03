# Activates a VNC installation on next boot.
resource "hetzner_boot_vnc" "example" {
  server_number = 12345
  dist          = "CentOS-Stream-9"
  lang          = "en_US"
}
