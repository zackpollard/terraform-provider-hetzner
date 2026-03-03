# Activates a Linux installation on next boot.
resource "hetzner_boot_linux" "example" {
  server_number = 12345
  dist          = "Rescue System"
  lang          = "en"
}
