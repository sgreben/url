FROM scratch
COPY binaries/linux_x86_64/url /url
ENTRYPOINT [ "/url" ]