"""Buildpack run-image package checker.

Reports the installed versions of the OS packages Snyk flagged in
wso2-enterprise/choreo#40063 (openssl/libssl3, gnupg2/gpgv), so a Choreo
component can be verified from the Console alone -- no kubectl, no exec.

These packages come from the buildpack RUN IMAGE that forms the base of the
final app image; nothing here installs them. That is what makes this a valid
check of which run image the build actually used.

Deploy a FRESH build. Redeploying an image built before the run-image pin was
updated will still report the old versions.

Output goes to stdout at startup (visible in Choreo runtime logs) and is also
served as JSON on GET / , so the check still works if endpoint invocation is
unavailable.
"""

import json
import os
import re
import sys
from http.server import BaseHTTPRequestHandler, ThreadingHTTPServer

DPKG_STATUS = "/var/lib/dpkg/status"
OS_RELEASE = "/etc/os-release"

# First version of each package carrying the fix, per the Snyk report on #40063.
FIXED = {
    "libssl3": "3.0.2-0ubuntu1.25",
    "openssl": "3.0.2-0ubuntu1.25",
    "gpgv": "2.2.27-3ubuntu2.5",
}


def read_installed(path=DPKG_STATUS, names=tuple(FIXED)):
    """Return {package: version} for names, read from the dpkg status file.

    Same source dpkg-query reads, so the output lines up with a
    `dpkg-query -W` run against the same image.
    """
    found = {}
    package = None
    try:
        with open(path, encoding="utf-8", errors="replace") as handle:
            for line in handle:
                if line.startswith("Package: "):
                    package = line[9:].strip()
                elif line.startswith("Version: ") and package in names:
                    # Multi-arch can list a package twice; keep the first.
                    found.setdefault(package, line[9:].strip())
    except OSError as err:
        print(f"PKGCHECK error reading {path}: {err}", file=sys.stderr, flush=True)
    return found


def compare(installed, fixed):
    """Compare two versions that differ only in a trailing integer.

    Returns "ok", "VULNERABLE", or "unknown". Deliberately narrow: both
    Ubuntu revisions here are `<prefix><int>`, and reporting "unknown" on
    any other shape beats guessing with a lexical compare.
    """
    got = re.match(r"^(.*?)(\d+)$", installed)
    want = re.match(r"^(.*?)(\d+)$", fixed)
    if not got or not want or got.group(1) != want.group(1):
        return "unknown"
    return "ok" if int(got.group(2)) >= int(want.group(2)) else "VULNERABLE"


def os_pretty_name(path=OS_RELEASE):
    try:
        with open(path, encoding="utf-8", errors="replace") as handle:
            for line in handle:
                if line.startswith("PRETTY_NAME="):
                    return line.split("=", 1)[1].strip().strip('"')
    except OSError:
        pass
    return "unknown"


def build_report():
    installed = read_installed()
    packages = {}
    for name, fixed in sorted(FIXED.items()):
        version = installed.get(name)
        packages[name] = {
            "installed": version,
            "fixed_in": fixed,
            "status": compare(version, fixed) if version else "not-installed",
        }
    statuses = {entry["status"] for entry in packages.values()}
    if "VULNERABLE" in statuses:
        verdict = "VULNERABLE"
    elif statuses == {"ok"}:
        verdict = "ok"
    else:
        verdict = "unknown"
    return {"verdict": verdict, "os": os_pretty_name(), "packages": packages}


def log_report(report):
    """Print one line per package, mirroring `dpkg-query -W` output order."""
    print(f"PKGCHECK os {report['os']}", flush=True)
    for name, entry in report["packages"].items():
        installed = entry["installed"] or "not-installed"
        print(
            f"PKGCHECK {name} {installed} "
            f"(fixed_in={entry['fixed_in']} status={entry['status']})",
            flush=True,
        )
    print(f"PKGCHECK verdict {report['verdict']}", flush=True)


class Handler(BaseHTTPRequestHandler):
    def _send(self, status, payload):
        body = json.dumps(payload, indent=2).encode()
        self.send_response(status)
        self.send_header("Content-Type", "application/json")
        self.send_header("Content-Length", str(len(body)))
        self.end_headers()
        self.wfile.write(body)

    def do_GET(self):
        path = self.path.split("?", 1)[0].rstrip("/") or "/"
        if path == "/healthz":
            self._send(200, {"healthy": True})
        elif path == "/":
            self._send(200, build_report())
        else:
            self._send(404, {"error": "not found"})

    def log_message(self, fmt, *args):
        print(f"{self.address_string()} {fmt % args}", flush=True)


def main():
    report = build_report()
    log_report(report)

    port = int(os.environ.get("PORT", "8080"))
    print(f"listening on {port}", flush=True)
    ThreadingHTTPServer(("", port), Handler).serve_forever()


if __name__ == "__main__":
    main()
