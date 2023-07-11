FROM goreleaser/goreleaser:latest as builder

WORKDIR /build/
ADD . /build/

RUN goreleaser build --single-target --snapshot

FROM gcr.io/distroless/base

LABEL org.opencontainers.image.source=https://github.com/cirruslabs/cirrus-webhooks-server

COPY --from=builder /build/dist/cws_linux_*/cws /bin/cws

ENTRYPOINT ["/bin/cws"]
