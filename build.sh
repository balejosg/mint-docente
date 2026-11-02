#!/bin/bash
# Construye salida/MintDocente-AAAAMMDD.ova de principio a fin.
#   ./build.sh            (sin ventana)
#   VER=1 ./build.sh      (mostrando la VM mientras se instala)
set -euo pipefail
cd "$(dirname "$0")"

ISO_URL=https://ftp.nl.freebsd.org/os/Linux/distr/linuxmint/iso/stable/22.3/linuxmint-22.3-xfce-64bit.iso
ISO_SHA256=45a835b5dddaf40e84d776549e0b19b3fbd49673b6cc6434ebddbfcd217df776
ISO=cache/linuxmint-22.3-xfce-64bit.iso
NOMBRE=MintDocente-$(date +%Y%m%d)

for c in VBoxManage packer curl sha256sum; do
  command -v "$c" >/dev/null || { echo "Falta '$c' (ver README.md)"; exit 1; }
done

mkdir -p cache
if ! echo "$ISO_SHA256  $ISO" | sha256sum -c --status 2>/dev/null; then
  echo "==> Descargando Linux Mint 22.3 XFCE"
  curl -fL -o "$ISO" "$ISO_URL"
  echo "$ISO_SHA256  $ISO" | sha256sum -c -
fi

[ -e "salida/$NOMBRE.ova" ] && { echo "Ya existe salida/$NOMBRE.ova"; exit 1; }
rm -rf salida/tmp

cd packer
packer init .
packer build -force \
  -var "iso_path=../$ISO" -var "iso_checksum=sha256:$ISO_SHA256" \
  -var "vm_name=$NOMBRE" -var "output_dir=../salida/tmp" \
  -var "headless=$([ "${VER:-0}" = 1 ] && echo false || echo true)" \
  .
cd ..
mv "salida/tmp/$NOMBRE.ova" "salida/$NOMBRE.ova"
rm -rf salida/tmp
sha256sum "salida/$NOMBRE.ova" | tee "salida/$NOMBRE.ova.sha256"
echo "==> Listo: salida/$NOMBRE.ova"
