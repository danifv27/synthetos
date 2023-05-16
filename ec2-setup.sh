#!/bin/bash

main_dir=/opt/uxperi
mkdir -p $main_dir
cd $main_dir

# Install dependencies
wget https://dl.google.com/linux/direct/google-chrome-stable_current_x86_64.rpm

yum -y install ./google-chrome-stable_current_x86_64.rpm
echo 'export CHROME_BIN=/usr/bin/chromium-browser' >> /home/ec2-user/.bashrc


# Download binary file from URL
curl -LJO https://tools.adidas-group.com/artifactory/pc-maven/com/adidas/devops/gherkin/linux/amd64/uxperi

# Make the downloaded binary file executable
chmod +x uxperi


# Create snapshots tmp folder
mkdir -p /tmp/snapshots

# Make uxperi as a runnable service

# (Input text)
service_content="
[Unit]
Description=Uxperi service

[Service]
ExecStart=$main_dir/uxperi test --logging.level=debug --no-probes.enable
Restart=always
User=ec2-user
Environment=SC_TEST_TIMEOUT=40s
Environment=SC_TEST_AZURE_USERNAME=svc_chromedp_cucu@emea.adsint.biz
Environment=SC_TEST_AZURE_PASSWORD=###Use lastpass password here
Environment=SC_TEST_SNAPSHOTS_FOLDER=/tmp/snapshots
Environment=CHROME_BIN=/usr/bin/chromium-browser

[Install]
WantedBy=multi-user.target"

# Redirect the service content into a new file
echo "$service_content" > /etc/systemd/system/uxperi.service

systemctl daemon-reload
systemctl start uxperi.service

# Enforce to run at start
systemctl enable uxperi.service