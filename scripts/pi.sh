#!/bin/bash

scp build/* pi@192.168.1.148:/tmp
read -d '' setup << EOF
sudo mv /tmp/philter-linux-arm5 /usr/bin/philter
sudo mv /tmp/blacklist.txt /var/lib/philter/blacklist.txt
sudo mv /tmp/philter.service  /etc/systemd/system/philter.service
sudo chmod 644 /etc/systemd/system/philter.service
chmod +x /usr/bin/philter
sudo systemctl daemon-reload
sudo systemctl enable philter.service
sudo systemctl start philter.service
EOF

ssh pi@192.168.1.148 "$setup"
