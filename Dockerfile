FROM alpine
ADD https://github.com/efritz/derision/releases/download/0.2/derision /
RUN chmod +x derision

FROM scratch
COPY --from=0 /derision .
ENTRYPOINT ["./derision"]
