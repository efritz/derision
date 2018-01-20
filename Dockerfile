FROM alpine
ADD https://github.com/efritz/derision/releases/download/0.1/derision /
RUN chmod +x derision

FROM scratch
COPY --from=0 /derision .
CMD ["./derision"]
