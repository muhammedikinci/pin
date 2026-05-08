# VPS Quickstart

Pin is built for the small self-hosted path: one VPS, Docker installed, one tiny daemon, and pipeline runs you can trigger and watch remotely.

## 1. Install Pin

```bash
curl -fsSL https://raw.githubusercontent.com/muhammedikinci/pin/main/scripts/install.sh | sh
```

Or build from source:

```bash
go build -o pin ./cmd/cli
sudo install -m 0755 pin /usr/local/bin/pin
```

## 2. Create a Pipeline

From your project directory:

```bash
pin init
pin validate -f pin.yaml
```

For a Dockerized Go or Node app, start from:

```bash
cp examples/vps-go-deploy.yaml pin.yaml
# or
cp examples/vps-node-deploy.yaml pin.yaml
```

Use `host: true` for deploy jobs that must run directly on the VPS, such as `docker compose up -d`, `systemctl restart`, or copying files into `/opt`.

## 3. Run the Daemon

Start locally first:

```bash
PIN_TOKEN=change-me pin daemon --port 8081 --data-dir .pin
```

For systemd, adapt [docs/systemd/pin-daemon.service](systemd/pin-daemon.service):

```bash
sudo useradd --system --home /var/lib/pin --shell /usr/sbin/nologin pin
sudo usermod -aG docker pin
sudo mkdir -p /var/lib/pin /opt/myapp
sudo chown -R pin:docker /var/lib/pin /opt/myapp
sudo cp docs/systemd/pin-daemon.service /etc/systemd/system/pin-daemon.service
sudo systemctl daemon-reload
sudo systemctl enable --now pin-daemon
```

## 4. Trigger and Watch

From your laptop:

```bash
pin trigger -f pin.yaml --url http://your-vps:8081 --token change-me
pin watch --url http://your-vps:8081 --token change-me
pin runs --url http://your-vps:8081 --token change-me
pin logs <run_id> --url http://your-vps:8081 --token change-me
```

Runs are stored under `.pin/runs` by default, or under `<data-dir>/runs` when `--data-dir` is provided.

## 5. Put It Behind HTTPS

Keep Pin bound to localhost and proxy it with Nginx:

```bash
pin daemon --host 127.0.0.1 --port 8081 --token change-me
```

Use [docs/nginx/pin.conf](nginx/pin.conf) as a starting point. The token is still required by Pin even when Nginx handles TLS.

## 6. GitHub Webhook

Run the daemon with a pipeline file and webhook secret:

```bash
PIN_TOKEN=change-me PIN_GITHUB_SECRET=github-secret \
  pin daemon \
  --host 127.0.0.1 \
  --port 8081 \
  --data-dir /var/lib/pin \
  --github-pipeline /opt/myapp/pin.yaml \
  --github-branch main
```

In GitHub repository settings, create a webhook:

- Payload URL: `https://pin.example.com/webhooks/github`
- Content type: `application/json`
- Secret: `github-secret`
- Event: `push`

When a push to `main` arrives, Pin verifies `X-Hub-Signature-256`, queues the configured pipeline, records the run, and streams events through `/events`.

## Host Jobs

Container jobs need `image` or `dockerfile`. Host jobs use `host: true` instead:

```yaml
deploy:
  host: true
  condition: $BRANCH == "main"
  script:
    - docker compose up -d --build
    - docker compose ps
```

Host jobs run with the daemon user's permissions. Keep the systemd user narrow and only grant the access it needs.
