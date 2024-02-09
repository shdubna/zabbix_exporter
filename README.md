# Zabbix Server Exporter for Prometheus
[![Go Report Card](https://goreportcard.com/badge/github.com/shdubna/zabbix_exporter)](https://goreportcard.com/report/github.com/shdubna/zabbix_exporter)
[![GitHub CodeQL](https://github.com/shdubna/zabbix_exporter/workflows/CodeQL/badge.svg)](https://github.com/shdubna/zabbix_exporter/actions?query=workflow%3CodeQL)
[![GitHub Release](https://github.com/shdubna/zabbix_exporter/workflows/Release/badge.svg)](https://github.com/shdubna/zabbix_exporter/actions?query=workflow%3ARelease)
[![GitHub license](https://img.shields.io/github/license/shdubna/zabbix_exporter.svg)](https://github.com/shdubna/zabbix_exporter/blob/main/LICENSE)
[![GitHub tag](https://img.shields.io/github/v/tag/shdubna/zabbix_exporter?label=latest)](https://github.com/shdubna/zabbix_exporter/releases)


This is a simple server that periodically scrapes [Zabbix server/proxy](https://www.zabbix.com/) stats and exports them via HTTP for [Prometheus](https://prometheus.io/)
consumption.

## How to use

1. [Configure zabbix server/proxy](https://www.zabbix.com/documentation/current/manual/appendix/items/remote_stats) to allow export internal stats:
   - add options ```StatsAllowedIP``` to configuration of server/proxy;
   - add environment var ```ZBX_STATSALLOWEDIP``` if you use official docker image;
2. Download zabbix_exporter from [release page](https://github.com/shdubna/zabbix_exporter/releases) or use [docker image](https://github.com/shdubna/testci/pkgs/container/zabbix_exporter).
3. Run zabbix exporter:
   - via binary
   ```bash
   ./zabbix_exporter --zabbix_addr <address of zabbix server/proxy>  --zabbix_port <port of zabbix server>
   ```
   - or via docker image:
   ```bash
   docker run -d -p 9051:9051 ghcr.io/shdubna/zabbix_exporter:latest --zabbix_addr <address of zabbix server/proxy>  --zabbix_port <port of zabbix server>
   ```
4. Check metrics via ```/metrics``` endpoint.
