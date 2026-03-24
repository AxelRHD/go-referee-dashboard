FROM scratch
COPY bin/referee-dashboard /referee-dashboard
COPY static /static
EXPOSE 3000
CMD ["/referee-dashboard", "serve"]
