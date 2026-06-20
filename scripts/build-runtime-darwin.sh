#!/usr/bin/env bash
set -euo pipefail

target="${1:?target is required, e.g. darwin-arm64}"
out="${2:?output tar.zst path is required}"
root="${RUNNER_TEMP:-/tmp}/magickgo-runtime-${target}"
build="${RUNNER_TEMP:-/tmp}/magickgo-build-${target}"
prefix="${build}/prefix"
imagemagick_version="${IMAGEMAGICK_VERSION:-7.1.2-8}"
imagemagick_url="${IMAGEMAGICK_URL:-https://imagemagick.org/archive/releases/ImageMagick-${imagemagick_version}.tar.xz}"

rm -rf "${root}" "${build}"
mkdir -p "${root}"/{bin,lib,etc,share}
mkdir -p "${build}" "${prefix}"

brew install jpeg-xl zstd ghostscript fontconfig libheif webp libpng libtiff librsvg freetype openjpeg little-cms2 libraw pango libtool || true
brew_prefix="$(brew --prefix)"
ghostscript_prefix="$(brew --prefix ghostscript 2>/dev/null || true)"
fontconfig_prefix="$(brew --prefix fontconfig 2>/dev/null || true)"

curl -fsSL "${imagemagick_url}" -o "${build}/imagemagick.tar.xz"
tar -C "${build}" --strip-components=1 -xJf "${build}/imagemagick.tar.xz"

export PKG_CONFIG_PATH="${brew_prefix}/lib/pkgconfig:${brew_prefix}/opt/jpeg-xl/lib/pkgconfig:${brew_prefix}/opt/libheif/lib/pkgconfig:${brew_prefix}/opt/webp/lib/pkgconfig:${brew_prefix}/opt/libpng/lib/pkgconfig:${brew_prefix}/opt/libtiff/lib/pkgconfig:${brew_prefix}/opt/librsvg/lib/pkgconfig:${brew_prefix}/opt/freetype/lib/pkgconfig:${brew_prefix}/opt/openjpeg/lib/pkgconfig:${brew_prefix}/opt/little-cms2/lib/pkgconfig:${brew_prefix}/opt/libraw/lib/pkgconfig:${brew_prefix}/opt/pango/lib/pkgconfig:${brew_prefix}/opt/fontconfig/lib/pkgconfig:${brew_prefix}/opt/libtool/lib/pkgconfig:${PKG_CONFIG_PATH:-}"
export LDFLAGS="-L${brew_prefix}/opt/libtool/lib ${LDFLAGS:-}"
export CPPFLAGS="-I${brew_prefix}/opt/libtool/include ${CPPFLAGS:-}"

(
  cd "${build}"
  ./configure \
    --prefix="${prefix}" \
    --disable-static \
    --enable-shared \
    --with-modules \
    --with-quantum-depth=16 \
    --enable-hdri \
    --with-bzlib \
    --with-fontconfig \
    --with-freetype \
    --with-heic \
    --with-jpeg \
    --with-jxl \
    --with-lcms \
    --with-lzma \
    --with-openjp2 \
    --with-pango \
    --with-png \
    --with-raw \
    --with-rsvg \
    --with-tiff \
    --with-webp \
    --with-xml \
    --with-zlib \
    --with-zstd
  make -j"$(sysctl -n hw.ncpu)"
  make install
)

cp -a "${prefix}/bin/magick" "${root}/bin/"
if [ -n "${ghostscript_prefix}" ] && [ -x "${ghostscript_prefix}/bin/gs" ]; then
  cp -a "${ghostscript_prefix}/bin/gs" "${root}/bin/"
elif [ -x "${brew_prefix}/bin/gs" ]; then
  cp -a "${brew_prefix}/bin/gs" "${root}/bin/"
fi

cp -pL "${prefix}"/lib/libMagickWand*.dylib "${root}/lib/"
cp -pL "${prefix}"/lib/libMagickCore*.dylib "${root}/lib/"

for dir in "${prefix}"/lib/ImageMagick "${prefix}"/lib/ImageMagick-*; do
  [ -d "${dir}" ] && cp -a "${dir}" "${root}/lib/"
done
for dir in "${prefix}"/etc/ImageMagick-*; do
  [ -d "${dir}" ] && cp -a "${dir}" "${root}/etc/"
done
for dir in "${prefix}"/share/ImageMagick-* "${brew_prefix}"/share/fonts "${fontconfig_prefix}"/etc/fonts; do
  [ -d "${dir}" ] && cp -a "${dir}" "${root}/share/"
done

if ! find "${root}/lib" -path '*/modules-*/coders/*' \( -name '*.dylib' -o -name '*.so' \) | grep -q .; then
  echo "ImageMagick coder modules were not copied into the runtime" >&2
  find "${root}/lib" -maxdepth 3 -type d >&2
  exit 1
fi

if ! find "${root}/lib" -path '*/config-*' -type d | grep -q .; then
  echo "ImageMagick module config directory was not copied into the runtime" >&2
  find "${root}/lib" -maxdepth 3 -type d >&2
  exit 1
fi

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

find "${root}/bin" "${root}/lib" -type f \( -perm -0100 -o -name '*.dylib' -o -name '*.so' \) -print0 |
  xargs -0 codesign --force --sign - --timestamp=none

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
