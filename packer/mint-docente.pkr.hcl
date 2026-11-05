packer {
  required_plugins {
    virtualbox = {
      source  = "github.com/hashicorp/virtualbox"
      version = ">= 1.1.0"
    }
  }
}

variable "iso_path" {
  type    = string
  default = "../cache/linuxmint-22.3-xfce-64bit.iso"
}

variable "iso_checksum" {
  type    = string
  default = "sha256:45a835b5dddaf40e84d776549e0b19b3fbd49673b6cc6434ebddbfcd217df776"
}

# uuid del medio en isolinux/live.cfg; cambia con cada ISO
variable "casper_uuid" {
  type    = string
  default = "6e72f523-dc09-4880-8910-93ffa64401c5"
}

variable "vm_name" {
  type    = string
  default = "MintDocente"
}

variable "output_dir" {
  type    = string
  default = "../salida"
}

variable "headless" {
  type    = bool
  default = true
}

locals {
  user = "docente"
  pass = "docente"
}

source "virtualbox-iso" "mint" {
  vm_name              = var.vm_name
  guest_os_type        = "Ubuntu_64"
  iso_url              = var.iso_path
  iso_checksum         = var.iso_checksum
  firmware             = "bios"
  cpus                 = 2
  memory               = 4096
  disk_size            = 40960
  hard_drive_interface = "sata"
  hard_drive_discard   = true
  headless             = var.headless
  http_directory       = "http"

  # En BIOS el ISO arranca con ISOLINUX: <esc> abre el prompt "boot:" y
  # escribimos la línea del kernel con instalación automática de Ubiquity.
  # casper exige el uuid del medio (isolinux/live.cfg del ISO fijado).
  boot_wait = "6s"
  boot_command = [
    "<esc><wait2>",
    "/casper/vmlinuz initrd=/casper/initrd.lz boot=casper uuid=${var.casper_uuid} ",
    "username=mint hostname=mint automatic-ubiquity noprompt ",
    "url=http://{{ .HTTPIP }}:{{ .HTTPPort }}/preseed.cfg ",
    "debian-installer/locale=es_ES.UTF-8 keyboard-configuration/layoutcode=es ",
    "-- <enter>"
  ]

  communicator           = "ssh"
  ssh_username           = local.user
  ssh_password           = local.pass
  ssh_timeout            = "90m"
  ssh_handshake_attempts = 200
  shutdown_command       = "echo '${local.pass}' | sudo -S shutdown -P now"

  guest_additions_mode = "upload"
  guest_additions_path = "/tmp/VBoxGuestAdditions.iso"

  vboxmanage = [
    ["modifyvm", "{{ .Name }}", "--graphicscontroller", "vmsvga"],
    ["modifyvm", "{{ .Name }}", "--vram", "128"],
    ["modifyvm", "{{ .Name }}", "--clipboard-mode", "bidirectional"],
    ["modifyvm", "{{ .Name }}", "--drag-and-drop", "bidirectional"],
    ["modifyvm", "{{ .Name }}", "--audio-enabled", "on"],
    ["modifyvm", "{{ .Name }}", "--audio-out", "on"],
    ["modifyvm", "{{ .Name }}", "--usb-xhci", "on"],
    ["modifyvm", "{{ .Name }}", "--mouse", "usbtablet"],
    ["modifyvm", "{{ .Name }}", "--ioapic", "on"],
  ]

  format           = "ova"
  output_directory = var.output_dir
  export_opts = [
    "--manifest",
    "--vsys", "0",
    "--product", "Mint Docente",
    "--description", "Linux Mint 22.3 XFCE + OpenCode Desktop, lista para docentes. Usuario: docente / contraseña: docente",
  ]
}

build {
  sources = ["source.virtualbox-iso.mint"]

  provisioner "shell" {
    inline = ["rm -rf /var/tmp/mint-docente && mkdir -p /var/tmp/mint-docente"]
  }

  provisioner "file" {
    source      = "../files/"
    destination = "/var/tmp/mint-docente"
  }

  provisioner "file" {
    source      = "../test"
    destination = "/var/tmp/mint-docente/"
  }

  provisioner "shell" {
    execute_command   = "echo '${local.pass}' | sudo -S -E bash '{{ .Path }}'"
    expect_disconnect = true
    scripts = [
      "../provision/00-base.sh",
    ]
  }

  # 00-base reinicia para cargar el kernel nuevo y las Guest Additions.
  provisioner "shell" {
    execute_command = "echo '${local.pass}' | sudo -S -E bash '{{ .Path }}'"
    pause_before    = "30s"
    scripts = [
      "../provision/10-documentos.sh",
      "../provision/20-opencode.sh",
      "../provision/30-escritorio.sh",
    ]
  }

  # 80-recorte quita kernels viejos, compiladores y lo que no se usa, y reinicia.
  provisioner "shell" {
    execute_command   = "echo '${local.pass}' | sudo -S -E bash '{{ .Path }}'"
    expect_disconnect = true
    scripts           = ["../provision/80-recorte.sh"]
  }

  # Se verifica después del reinicio: la máquina arranca con lo que queda.
  provisioner "shell" {
    execute_command = "echo '${local.pass}' | sudo -S -E bash '{{ .Path }}'"
    pause_before    = "30s"
    inline          = ["bash /var/tmp/mint-docente/test/verificar.sh"]
  }

  provisioner "shell" {
    execute_command = "echo '${local.pass}' | sudo -S -E bash '{{ .Path }}'"
    scripts         = ["../provision/90-limpieza.sh"]
  }
}
