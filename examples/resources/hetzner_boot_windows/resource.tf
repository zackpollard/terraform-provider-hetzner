# Activates a Windows installation on next boot.
resource "hetzner_boot_windows" "example" {
  server_number = 12345
  dist          = "standard"
  lang          = "en_US"
}
