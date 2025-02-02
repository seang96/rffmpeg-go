FROM --platform=$BUILDPLATFORM node:18 as build
WORKDIR /build
ARG BRANCH=release-10.8.z
RUN git clone --depth 1 -b "$BRANCH" https://github.com/jellyfin/jellyfin-web .
ADD https://github.com/jellyfin/jellyfin-web/compare/${BRANCH}...ConfusedPolarBear:jellyfin-web:intros.patch intros.patch
RUN git apply intros.patch
RUN npm ci && npm run build:production

FROM ghcr.io/onedr0p/jellyfin:10.8.10@sha256:1ef614db6a4c589777eb48bc9004d573b9c09f0d6d573a509041c6060f3a956b

USER root
COPY rffmpeg-go /usr/local/bin/rffmpeg

RUN ln -s /usr/local/bin/rffmpeg /usr/local/bin/ffmpeg \
    && ln -s /usr/local/bin/rffmpeg /usr/local/bin/ffprobe

RUN mkdir -p /run/shm \
    && chown kah:kah /run/shm

RUN apt-get -qq update \
    && apt-get -qq install -y openssh-client \
    && apt-get purge -y --auto-remove -o APT::AutoRemove::RecommendsImportant=false \
    && apt-get autoremove -y \
    && apt-get clean \
    && \
    rm -rf \
        /tmp/* \
        /var/lib/apt/lists/* \
        /var/tmp/

RUN rm -rf /usr/share/jellyfin/web/
COPY --from=build /build/dist/ /usr/share/jellyfin/web/

USER kah
COPY ./apps/rffmpeg-go/rffmpeg.yml /etc/rffmpeg/rffmpeg.yml
COPY ./apps/rffmpeg-go/entrypoint.sh /entrypoint.sh
CMD ["/entrypoint.sh"]

LABEL org.opencontainers.image.source="https://github.com/jellyfin/jellyfin"