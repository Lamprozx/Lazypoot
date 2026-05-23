FROM --platform=linux/arm64 ubuntu:24.04

RUN apt update && apt install -y \
    proot curl wget xz-utils tar sudo git

RUN mkdir -p /data/data/com.termux/files/usr \
    /data/data/com.termux/files/home \
    /storage/emulated/0

ENV PREFIX=/data/data/com.termux/files/usr
ENV HOME=/data/data/com.termux/files/home
ENV TMPDIR=/data/data/com.termux/files/usr/tmp

RUN mkdir -p $TMPDIR

RUN printf '#!/bin/bash\necho arm64-v8a\n' > /usr/local/bin/getprop \
 && chmod +x /usr/local/bin/getprop
