FROM debian:bookworm-slim

RUN apt-get update && apt-get install -y \
    ca-certificates \
		&& rm -rf /var/lib/apt/lists/*

ADD dist/e-ink-dashboard /usr/local/bin/
RUN mkdir /e-ink-dashboard \
		&& chmod 0755 /usr/local/bin/e-ink-dashboard

WORKDIR /e-ink-dashboard
ENTRYPOINT ["/usr/local/bin/e-ink-dashboard"]
