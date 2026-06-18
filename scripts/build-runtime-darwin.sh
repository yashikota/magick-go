#!/usr/bin/env bash
set -euo pipefail

target="${1:?target is required, e.g. darwin-arm64}"
out="${2:?output tar.zst path is required}"
root="${RUNNER_TEMP:-/tmp}/magickgo-runtime-${target}"

rm -rf "${root}"
mkdir -p "${root}"/{bin,lib,etc,share}

brew install imagemagick zstd ghostscript fontconfig || true
prefix="$(brew --prefix)"

cp -a "${prefix}/bin/magick" "${root}/bin/"
[ -x "${prefix}/bin/gs" ] && cp -a "${prefix}/bin/gs" "${root}/bin/" || true

cp -pL "${prefix}"/lib/libMagickWand*.dylib "${root}/lib/"
cp -pL "${prefix}"/lib/libMagickCore*.dylib "${root}/lib/"

for dir in "${prefix}"/lib/ImageMagick-*; do
  [ -d "${dir}" ] && cp -a "${dir}" "${root}/lib/"
done
for dir in "${prefix}"/etc/ImageMagick-*; do
  [ -d "${dir}" ] && cp -a "${dir}" "${root}/etc/"
done
for dir in "${prefix}"/share/ImageMagick-* "${prefix}"/share/fonts "${prefix}"/etc/fonts; do
  [ -d "${dir}" ] && cp -a "${dir}" "${root}/share/"
done

copy_deps() {
  local changed=1
  while [ "${changed}" -eq 1 ]; do
    changed=0
    while read -r lib; do
      [ -f "${lib}" ] || continue
      case "${lib}" in
        /usr/lib/*|/System/*)
          continue
          ;;
      esac
      name="$(basename "${lib}")"
      dest="${root}/lib/${name}"
      if [ ! -e "${dest}" ]; then
        cp -pL "${lib}" "${dest}" || true
        changed=1
      fi
    done < <(
      find "${root}/bin" "${root}/lib" -type f \( -perm -0100 -o -name '*.dylib' \) -print0 |
        xargs -0 otool -L 2>/dev/null |
        awk '/^\t/ {print $1}' |
        sort -u
    )
  done
}
copy_deps

find "${root}/lib" -type f -name '*.dylib' | while read -r lib; do
  install_name_tool -id "@loader_path/$(basename "${lib}")" "${lib}" || true
  otool -L "${lib}" | awk '/\/opt\/homebrew|\/usr\/local/ {print $1}' | while read -r dep; do
    base="$(basename "${dep}")"
    [ -f "${root}/lib/${base}" ] && install_name_tool -change "${dep}" "@loader_path/${base}" "${lib}" || true
  done
done

if grep -R "/opt/homebrew\|/usr/local" "${root}/lib" >/dev/null 2>&1; then
  echo "absolute Homebrew references remain in runtime" >&2
  exit 1
fi

mkdir -p "$(dirname "${out}")"
tar -C "${root}" --zstd -cf "${out}" .
