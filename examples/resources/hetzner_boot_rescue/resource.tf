# Activates the rescue system on next boot.
resource "hetzner_boot_rescue" "example" {
  server_number = 12345
  os            = "linux"
}
