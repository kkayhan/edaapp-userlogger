# EDA User Logger

Automatically logs all EDA transactions and Keycloak authentication events into monthly audit log files. Provides a read-only HTTP API for viewing and downloading logs.

## What It Logs

- **EDA Transactions**: Every configuration change -- who changed what, when, from which IP, with a flattened diff of the config changes
- **Keycloak Login/Logout Events**: GUI sign-in and sign-out events with username and source IP
- **Keycloak Admin Events**: User/group/client-role/realm modifications made through Keycloak

### Sample Log Output

```
2026-04-16T12:20:22 UTC | Event=Transaction-99 | User=admin | IPADDR=10.0.0.5 | Modified=Fabric | Namespace=default | Fabric resource named fabric1 has been updated.
   interface-ethernet/ethernet-1-1/admin-state: enable -> disable

2026-04-16T12:22:40 UTC | Event=EDA-Login | User=admin | IPADDR=10.244.0.27 | The user signed-in to the EDA GUI.
```

## Prerequisites

- EDA v25.12.x or later
- Access to `ghcr.io` from the EDA cluster nodes

## Installation

### Step 1: Add the Catalog

Apply this Catalog CR to register the app in your EDA App Store:

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

```bash
kubectl apply -f catalog.yaml
```

### Step 2: Install from the App Store

Open the EDA GUI, navigate to the **App Store**. The **EDA User Logger** will appear under the *Integrations* category. Click **Install**.

Or install via CLI:

```yaml
apiVersion: appstore.eda.nokia.com/v1
kind: AppInstaller
metadata:
  name: install-userlogger
  namespace: eda-system
spec:
  operation: install
  dryRun: false
  apps:
    - appId: edauserlogger.eda.tt.com
      catalog: community-apps
      version:
        type: semver
        value: "v0.1.0+260416121951"
```

```bash
kubectl apply -f install.yaml
```

### Step 3: Post-Install Setup

After the app is installed, patch the webhook failure policies and create the default configuration:

```bash
# Patch webhooks
kubectl patch validatingwebhookconfiguration \
  eda-edauserlogger-validating-webhook-configuration \
  --type='json' -p='[{"op":"replace","path":"/webhooks/0/failurePolicy","value":"Ignore"}]'

kubectl patch mutatingwebhookconfiguration \
  eda-edauserlogger-mutating-webhook-configuration \
  --type='json' -p='[{"op":"replace","path":"/webhooks/0/failurePolicy","value":"Ignore"}]'

# Create default config
cat <<'EOF' | kubectl apply -f -
apiVersion: edauserlogger.eda.tt.com/v1alpha1
kind: UserLogConfig
metadata:
  name: default
spec:
  pollIntervalSeconds: 300
  retentionMonths: 0
EOF
```

### Step 4: Verify

```bash
# Check the pod is running
kubectl -n eda-system get pods -l eda.nokia.com/app=eda-edauserlogger

# Check health
curl -sk https://<eda-address>/core/httpproxy/v1/userlogger/healthz
```

## Usage

### Log Endpoints

All endpoints are accessible via the EDA HttpProxy:

```
https://<eda-address>/core/httpproxy/v1/userlogger/
```

| Endpoint | Description |
|----------|-------------|
| `/healthz` | Health status and last poll time |
| `/logs/` | List all log files with sizes and timestamps (JSON) |
| `/logs/<filename>` | Download a specific log file (plain text) |

**Examples:**

```bash
EDA=https://<eda-address>
BASE=$EDA/core/httpproxy/v1/userlogger

# Health check
curl -sk $BASE/healthz

# List log files
curl -sk $BASE/logs/

# Download current month's log
curl -sk $BASE/logs/EDA-user-events-2026-05.log
```

### CRD Status

```bash
kubectl get userlogconfig default -o yaml
```

Reports: `health`, `subsystems` (edaApi, keycloakEvents), `lastPollTime`, `lastTransactionId`, `transactionsProcessed`, `kcEventsProcessed`, `logFiles`, `version`.

## Configuration

### UserLogConfig CRD

```bash
kubectl edit userlogconfig default
```

| Field | Default | Range | Description |
|-------|---------|-------|-------------|
| `pollIntervalSeconds` | 300 | 60-3600 | Polling interval in seconds |
| `retentionMonths` | 0 | 0+ | Months of logs to keep (0 = unlimited) |

### App Settings

Adjustable via the EDA App Store settings:

| Setting | Default | Description |
|---------|---------|-------------|
| `controllerCpuLimit` | 200m | CPU limit |
| `controllerMemoryLimit` | 128Mi | Memory limit |
| `logStorageSize` | 500Mi | PVC size for log storage |

## Uninstalling

Via CLI:

```yaml
apiVersion: appstore.eda.nokia.com/v1
kind: AppInstaller
metadata:
  name: uninstall-userlogger
  namespace: eda-system
spec:
  operation: delete
  dryRun: false
  apps:
    - appId: edauserlogger.eda.tt.com
      catalog: community-apps
```

Or remove through the EDA App Store UI.

## Troubleshooting

**Pod in ImagePullBackOff**: Verify the cluster can reach `ghcr.io`. The images are public -- no authentication needed.

**Health shows degraded/error**: Check `kubectl get userlogconfig default -o yaml` and pod logs.

**No logs**: The controller auto-discovers the latest transaction ID on first start and only logs new events going forward. Make a change in EDA and wait for the next poll cycle.

**HttpProxy 404**: Verify the HttpProxy CR exists: `kubectl get httpproxies.core.eda.nokia.com userlogger`
