FROM alpine
COPY zabbix_exporter /bin/
ENTRYPOINT ["zabbix_exporter"]
