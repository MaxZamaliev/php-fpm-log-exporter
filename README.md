<h1>Build on Linux:</h1>

mkdir ~/php-fpm-log-exporter<br>
cd ~/php-fpm-log-exporter<br>
git clone https://github.com/MaxZamaliev/php-fpm-log-exporter.git<br>
export GOPATH=\`pwd\`<br>
go get github.com/hpcloud/tail<br>
go get github.com/prometheus/client_golang/prometheus<br>
go get github.com/prometheus/client_golang/prometheus/promhttp<br>
go build php-fpm-log-exporter.go<br>


<h1>Install on CentOS 8:</h1>

<b>1. Copy binary file php-fpm-log-exporter to /usr/local/bin/</b>

<b>2. Create user:</b>

useradd -M -s /bin/false php-fpm_exporter

<b>3. Create file `/etc/systemd/system/php-fpm-log_exporter.service` with text:</b>

[Unit]<br>
Description=Prometheus PHP-FPM Exporter<br>
Wants=network-online.target<br>
After=network-online.target<br>
<br>
[Service]<br>
User=php-fpm_exporter<br>
Group=php-fpm_exporter<br>
Type=simple<br>
ExecStart=/usr/local/bin/php-fpm-log-exporter<br>
<br>
[Install]<br>
WantedBy=multi-user.target<br>

<b>4. Enable and start service:</b>

systemctl enable --now php-fpm-log-exporter

<b>5. Test service:</b>

curl http://localhost:9253/metrics
