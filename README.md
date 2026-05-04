# EDA User Logger

A Nokia **EDA** app that turns the EDA cluster into a system-of-record for **who did what**. It continuously records:

- every **configuration change** made in EDA — the user who made it, when, from which IP address, and a human-readable diff of what changed on each device
- every **sign-in and sign-out** to the EDA GUI
- every **administrative change** in Keycloak (user / group / role management)

All events are written to monthly log files (`Transaction-YYYY-MM.log`) that are accessible directly from the EDA UI or a simple HTTP endpoint — no scraping, no parsing, no extra tooling. A typical line looks like this:

```
2026-04-20T08:41:00 UTC | Event=EDA-Login | User=admin | IPADDR=10.244.0.55 | The user signed-in to the EDA GUI.
2026-04-20T07:26:09 UTC | Event=Transaction-101 | User=kubernetes | Modified=EDA | Namespace=eda | TargetNode resource named leaf2 has been created.
```

Designed for compliance archives, SIEM feeds, change-management audits, and "who broke the fabric last Tuesday?" conversations.

---

## Install (from the EDA UI)

There's nothing to configure. The app starts logging the moment it's installed.

**Step 1 — Add this catalog to your EDA cluster (one-time):**

1. In the EDA UI, go to **System Administration**.
2. Under **APP Management**, open **Catalogs**.
3. Click **Create** and paste the YAML below.
4. **Commit**.

```yaml
apiVersion: appstore.eda.nokia.com/v1
kind: Catalog
metadata:
  name: community-apps
  namespace: eda-system
spec:
  remoteURL: https://github.com/kkayhan/edaapp-userlogger.git
  skipTLSVerify: false
  title: Community EDA Apps
```

**Step 2 — Install from the Store:**

Open the **App Store** in the EDA UI. "EDA User Logger" will appear under *Integrations*. Click **Install**. That's it — no settings to fill in, no credentials to configure.

The controller starts immediately, enables Keycloak event auditing on your behalf, and begins writing the first log file within one poll cycle (default: 5 minutes).

---

## Where the logs are

### HTTP endpoint

Logs are served over the EDA HttpProxy at `https://<your-eda-host>/core/httpproxy/v1/userlogger/logs/`.

**Step 1 — list the available log files.** A `GET` on `/logs/` returns a JSON array of every file currently on disk, with sizes and timestamps:

```bash
curl -sk https://<your-eda-host>/core/httpproxy/v1/userlogger/logs/
```

```json
[
  {"name": "Transaction-2026-04.log", "size_bytes": 18432, "modified": "2026-04-30T23:59:00Z"},
  {"name": "Transaction-2026-05.log", "size_bytes":  4221, "modified": "2026-05-04T08:14:12Z"}
]
```

**Step 2 — download a specific file.** Append the `name` from the listing to the URL:

```bash
curl -sk https://<your-eda-host>/core/httpproxy/v1/userlogger/logs/Transaction-2026-05.log
```

### Helper script

[`logs/pull-logs.sh`](logs/pull-logs.sh) wraps both steps so you can grab everything in one command:

```bash
# Download every log file into the current directory
./pull-logs.sh https://<your-eda-host>

# Download every log file into ./audit-archive
./pull-logs.sh https://<your-eda-host> ./audit-archive

# Download a single named file
./pull-logs.sh https://<your-eda-host> ./audit-archive Transaction-2026-05.log
```

---

## What it does NOT do

- Does **not** forward logs to external systems (syslog / SIEM / S3). Pull logs over HTTP into whatever system you already run.
- Does **not** require (or accept) any credentials — it reads existing Kubernetes secrets inside the cluster.
- Does **not** filter log access per user. Anyone authenticated to EDA can read the audit log.

---

## License

MIT. See [`LICENSE`](LICENSE) for details.
