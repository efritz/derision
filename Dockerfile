FROM scratch

EXPOSE 5000
COPY derision /
CMD ["/derision"]
