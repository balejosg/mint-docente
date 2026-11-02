#!/bin/bash
# Prepara el anfitrión de construcción (Debian 13): VirtualBox 7.2 + convivencia con KVM.
# Uso: sudo ./instalar-virtualbox-host.sh
set -euo pipefail
[ "$(id -u)" -eq 0 ] || { echo "Ejecútalo con sudo"; exit 1; }
USUARIO=${SUDO_USER:-run0}
curl -fsSL https://www.virtualbox.org/download/oracle_vbox_2016.asc \
  | gpg --dearmor --yes -o /usr/share/keyrings/oracle-virtualbox-2016.gpg
echo "deb [arch=amd64 signed-by=/usr/share/keyrings/oracle-virtualbox-2016.gpg] https://download.virtualbox.org/virtualbox/debian trixie contrib" \
  > /etc/apt/sources.list.d/virtualbox.list
apt-get update
apt-get install -y virtualbox-7.2 "linux-headers-$(uname -r)"
usermod -aG vboxusers "$USUARIO"
# Kernel ≥ 6.12: KVM toma VT-x al cargar y VirtualBox no puede arrancar VMs.
echo "options kvm enable_virt_at_load=0" > /etc/modprobe.d/kvm-virtualbox.conf
modprobe -r kvm_intel kvm && modprobe kvm_intel || echo "AVISO: no se pudo recargar kvm (¿VM KVM en uso?). Reinicia el equipo."
VBoxManage --version
echo "Listo. Cierra sesión y vuelve a entrar (grupo vboxusers)."
