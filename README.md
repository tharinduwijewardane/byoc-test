# byoc-test

Example user apps for the Choreo BYOC feature.

## `python/` — buildpack run-image package check

A minimal Python service that reports the OS package versions it inherited
from the **Google buildpack run image**, so you can verify which run image a
Choreo build actually used — from the Console alone, with no `kubectl` and no
`exec` into the pod.

Built for [wso2-enterprise/choreo#40063](https://github.com/wso2-enterprise/choreo/issues/40063),
where Snyk flagged unpatched `openssl`/`libssl3` (CVE-2026-45447) and
`gnupg2`/`gpgv` in customer images built on a run image pinned in 2024.

### Using it

1. Create a **Service** component from this repo with build context `/python`,
   buildpack **Python**. Any Google-buildpack language would inherit the same
   run image; Ballerina would not — it uses a separate Choreo-managed Alpine
   run image and is unaffected.
2. Build and deploy. **The build must be fresh** — redeploying an image built
   before the run-image pin changed will still report the old versions.
3. Read the result either way:
   - **Runtime logs** in the Console — the report is printed at startup, so
     this works even if endpoint invocation fails.
   - **`GET /`** — the same report as JSON.

### Output

```
PKGCHECK os Ubuntu 22.04.x LTS
PKGCHECK gpgv 2.2.27-3ubuntu2.5 (fixed_in=2.2.27-3ubuntu2.5 status=ok)
PKGCHECK libssl3 3.0.2-0ubuntu1.26 (fixed_in=3.0.2-0ubuntu1.25 status=ok)
PKGCHECK openssl 3.0.2-0ubuntu1.26 (fixed_in=3.0.2-0ubuntu1.25 status=ok)
PKGCHECK verdict ok
```

A build on the old pinned run image reports `3.0.2-0ubuntu1.16` /
`2.2.27-3ubuntu2.1` and `verdict VULNERABLE`.

`status` is `unknown` when a version no longer matches the
`<prefix><integer>` shape these Ubuntu revisions use — the comparison reports
uncertainty rather than guessing.

### Endpoints

| Method | Path | Purpose |
|--------|------|---------|
| GET | `/` | Package report as JSON |
| GET | `/healthz` | Health check |

The versions come from `/var/lib/dpkg/status`, the same source `dpkg-query`
reads, so output lines up with:

```
dpkg-query -W -f='${binary:Package} ${Version}\n' libssl3 openssl gpgv
```
