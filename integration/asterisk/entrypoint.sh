#!/bin/sh

set -eo pipefail

/usr/sbin/asterisk -T -W -U asterisk -p -vvvdddf | /log_sender -addr integration:8080
# /usr/sbin/asterisk -T -W -U asterisk -p -vvvdddf

