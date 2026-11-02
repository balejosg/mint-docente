#!/bin/sh
# Ejecutado por Ubiquity (success_command) en el entorno del instalador, con el
# sistema instalado en /target. Deja preparado el acceso SSH temporal para Packer:
# openssh-server se instala en el PRIMER ARRANQUE (aquí no hay DNS).
set -x
cat > /target/etc/systemd/system/packer-ssh.service <<'UNIT'
[Unit]
Description=SSH temporal para Packer (lo retira 90-limpieza)
Wants=network-online.target
After=network-online.target

[Service]
Type=oneshot
ExecStart=/bin/sh -c 'until apt-get -o DPkg::Lock::Timeout=600 update && DEBIAN_FRONTEND=noninteractive apt-get -o DPkg::Lock::Timeout=600 install -y openssh-server; do sleep 10; done; systemctl enable --now ssh'

[Install]
WantedBy=multi-user.target
UNIT
mkdir -p /target/etc/systemd/system/multi-user.target.wants
ln -sf /etc/systemd/system/packer-ssh.service \
  /target/etc/systemd/system/multi-user.target.wants/packer-ssh.service
echo 'docente ALL=(ALL) NOPASSWD:ALL' > /target/etc/sudoers.d/99-packer
chmod 440 /target/etc/sudoers.d/99-packer
rm -f /target/root/postinstalacion.sh
