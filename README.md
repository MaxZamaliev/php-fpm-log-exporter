Build on Linux:

mkdir ~/php-fpm-log-exporter
cd ~/php-fpm-log-exporter
git clone https://github.com/MaxZamaliev/php-fpm-log-exporter.git
export GOPATH=`pwd`
go get github.com/hpcloud/tail
go get github.com/prometheus/client_golang/prometheus
go get github.com/prometheus/client_golang/prometheus/promhttp
go build php-fpm-log-exporter.go


Install on CentOS 8:

1. Copy binary file php-fpm-log-exporter to /usr/local/bin/

2. Create user:

useradd -M -s /bin/false php-fpm_exporter

3. Create file `/etc/systemd/system/php-fpm-log_exporter.service` with text:

[Unit]
Description=Prometheus PHP-FPM Exporter
Wants=network-online.target
After=network-online.target

[Service]
User=php-fpm_exporter
Group=php-fpm_exporter
Type=simple
ExecStart=/usr/local/bin/php-fpm-log-exporter

[Install]
WantedBy=multi-user.target

4. Enable and start service:

systemctl enable --now php-fpm-log-exporter

5. Test service:

curl http://localhost:9253/metrics
