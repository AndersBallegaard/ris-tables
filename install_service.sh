#!/bin/bash

sudo rm -rf /usr/bin/ris_tables
sudo cp bin/ris_tables-linux-amd64 /usr/bin/ris_tables

sudo rm -rf /usr/lib/systemd/system/ris-table.service
sudo cp ris-table.service /usr/lib/systemd/system/


sudo systemctl enable --now ris-table.service
