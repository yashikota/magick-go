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

for dir in "${prefix}"/lib/ImageMagick "${prefix}"/lib/ImageMagick-*; do
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
      find "${root}/bin" "${root}/lib" -type f \( -perm -0100 -o -name '*.dylib' -o -name '*.so' \) -print0 |
        xargs -0 otool -L 2>/dev/null |
        awk '/^\t/ {print $1}' |
        sort -u
    )
  done
}
copy_deps

rewrite_dep() {
  local file="${1}"
  local dep="${2}"
  local base
  local replacement
  base="$(basename "${dep}")"
  [ -f "${root}/lib/${base}" ] || return 0
  case "${file}" in
    "${root}"/bin/*)
      replacement="@executable_path/../lib/${base}"
      ;;
    "${root}"/lib/*.dylib)
      replacement="@loader_path/${base}"
      ;;
    "${root}"/lib/ImageMagick*/*/*/*.dylib|"${root}"/lib/ImageMagick*/*/*/*.so)
      replacement="@loader_path/../../../${base}"
      ;;
    *)
      replacement="@loader_path/${base}"
      ;;
  esac
  install_name_tool -change "${dep}" "${replacement}" "${file}" || true
}

find "${root}/bin" "${root}/lib" -type f \( -perm -0100 -o -name '*.dylib' -o -name '*.so' \) | while read -r file; do
  case "${file}" in
    *.dylib)
      install_name_tool -id "@loader_path/$(basename "${file}")" "${file}" || true
      ;;
  esac
  otool -L "${file}" | awk '/\/opt\/homebrew|\/usr\/local/ {print $1}' | while read -r dep; do
    rewrite_dep "${file}" "${dep}"
  done
done

remaining_refs="$(
  find "${root}/bin" "${root}/lib" -type f \( -perm -0100 -o -name '*.dylib' -o -name '*.so' \) -print0 |
    xargs -0 otool -L 2>/dev/null |
    awk '/:$/ {file=$0} /^\t\/opt\/homebrew|^\t\/usr\/local/ {print file " " $1}'
)"
if [ -n "${remaining_refs}" ]; then
  echo "absolute Homebrew references remain in runtime" >&2
  echo "${remaining_refs}" >&2
  exit 1
fi

mkdir -p "$(dirname "${out}")"
tar -C "${root}" --zstd -cf "${out}" .
