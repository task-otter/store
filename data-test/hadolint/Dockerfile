FROM alpine:3.20
USER 65534:65534
HEALTHCHECK --interval=30s --timeout=3s CMD kill -0 1 || exit 1
CMD ["sleep", "3600"]
