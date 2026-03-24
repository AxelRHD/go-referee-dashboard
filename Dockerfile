FROM scratch
COPY bin/referee-dashboard /referee-dashboard
COPY static /static
USER 1000:1000
EXPOSE 3000
CMD ["/referee-dashboard", "serve"]
