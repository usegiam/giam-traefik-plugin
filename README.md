# Traefik Plugin: `Giam`

A Traefik middleware plugin that enforces **Label-Based Access Control** (LBAC) on Prometheus/Thanos and Loki queries.  It transparently rewrites incoming PromQL/LogQL requests to inject per-team, per-environment or per-cluster label constraints—so you can gate access to metrics and logs without touching your dashboards or scrapers.

---

## Overview

When used in Traefik as a middleware, `giam`:

- Parses each incoming HTTP query of Grafana to Prometheus/Thanos or Loki.
- Enforce label matchers based on your LBAC policy.
- Supports Equal, Match Regex, Not Equal, and Not Match Regex rules in any combination.
- Integrates with your OAuth/OIDC provider or Grafana teams to map users → policies.

---

## Key Features

- **Zero Code Changes**
  Simply point your data-source traffic through Traefik + `giam-lbac`. No changes to alerts, dashboards, or instrumentation.

- **Full PromQL & LogQL Support**
  Handles counters, gauges, histograms, aggregations, ranges, regex matchers, and raw log queries.

- **Dynamic Policy Loading**
  Pull rules from YAML, Kubernetes ConfigMaps, or an HTTP endpoint.

- **Identity Provider Hooks**
  Automatically resolve teams/groups from OAuth2/OIDC tokens or Grafana session cookies.

- **Lightweight & High-Performance**
  Written in Go as a native Traefik plugin; adds only milliseconds of latency.

---

## Requirements

- **Traefik v2.5+** with [plugin support enabled](https://doc.traefik.io/traefik/v2.5/plugins/overview/).
- A running **Grafana** and **Prometheus/Thanos** or **Loki** instance.
- API Key that you can generate form our [portal](https://portal.usegiam.com). 
---

## Installation

- Please follow our official [documentation](https://amusing-lighter-a68.notion.site/Quick-Start-1cdb1b6682e380c9a618ea6db0342325)
