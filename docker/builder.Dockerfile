FROM debian:bookworm

ENV DEBIAN_FRONTEND=noninteractive
ENV MAHOU_BUILDER=1

RUN apt-get update && apt-get install -y \
  autoconf \
  automake \
  build-essential \
  curl \
  fontconfig \
  ghostscript \
  libbz2-dev \
  libdjvulibre-dev \
  libfftw3-dev \
  libfontconfig1-dev \
  libfreetype6-dev \
  libheif-dev \
  libjbig-dev \
  libjpeg-dev \
  libjxl-dev \
  liblcms2-dev \
  liblqr-1-0-dev \
  libltdl-dev \
  liblzma-dev \
  libopenexr-dev \
  libopenjp2-7-dev \
  libpango1.0-dev \
  libpng-dev \
  libraqm-dev \
  librsvg2-dev \
  libraw-dev \
  libtiff-dev \
  libtool \
  libwebp-dev \
  libwmf-dev \
  libxml2-dev \
  libzip-dev \
  patchelf \
  pkg-config \
  tar \
  xz-utils \
  zlib1g-dev \
  zstd \
  && rm -rf /var/lib/apt/lists/*
