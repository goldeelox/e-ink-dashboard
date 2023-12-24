FROM debian:bookworm-slim

ADD dist/e-ink-dashboard /usr/local/bin/
RUN mkdir /var/lib/e-ink-dashboard \
		&& chmod 0755 /usr/local/bin/e-ink-dashboard

ENTRYPOINT ["/usr/local/bin/e-ink-dashboard"]
