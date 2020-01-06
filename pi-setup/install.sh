#! /usr/bin/env bash

set -euo pipefail

USER=remmelt

echo Usage: sudo bash install.sh as user pi on a fresh raspbian installation

sed -i 's/raspberrypi/discotoilet/' /etc/hostname /etc/hosts

apt-get update
apt-get upgrade -y
apt-get install -y htop tmux unattended-upgrades fail2ban openssh-server

cat <<EOF > /etc/apt/apt.conf.d/02periodic
APT::Periodic::Enable "1";
APT::Periodic::Update-Package-Lists "1";
APT::Periodic::Download-Upgradeable-Packages "1";
APT::Periodic::Unattended-Upgrade "1";
APT::Periodic::AutocleanInterval "1";
APT::Periodic::Verbose "2";
EOF

adduser $USER
usermod -a -G adm,dialout,cdrom,sudo,audio,video,plugdev,games,users,input,netdev,gpio,i2c,spi $USER
cat $USER ALL=(ALL) PASSWD: ALL >> /etc/sudoers.d/010_pi-nopasswd

pkill -u pi
deluser -remove-home pi

cat <<EOF >> /etc/ssh/sshd_config
AllowUsers $USER
ChallengeResponseAuthentication no
PasswordAuthentication no
UsePAM no
EOF

mkdir ~$USER/.ssh/
chmod 0700 ~$USER/.ssh/
cat <<EOF > ~$USER/.ssh/authorized_keys
ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAACAQCrkq5rkorKGbqg0MclVexkF4OxFENSOqRR1S8PZV6I+hd3QlygIiv+mQmIBE96r7Ea9ar2VgHcnumLMsVpHgVqNAYpbsMCZqh0VXu8f3PHrv+DvbhnJJmhyejghxRPT9FWhWKj6VqcGs8U7i6EUHyimFHwl3t9R9DtyqH2hRf+GuNvLlo2WjptQlpHGTRSI3leXuWapKNKKKNmP8FSqK0EHojgwzgojeOD04qkvKpCaNuHzAQxlTvIXqChqzcQ2sffd64poZwOerxUYTyMcWR6bi24RSuVzN9UVRwnpK8D/WzD0wMMFhzeUcQmOdcbrHKs0ifJq+lGQEgKlqfdkIqbmHR0EsC0198fNjZUxuQoobvcBcESjPdNT6BCY6dMhIxrk9tugpOgii737NZMmSH/9zNt0QEsMpaEJeYkctclypQAveAnUytKie/dTfLNN8LEWUdlIbkO+0S+1nr9D6JFmBOvHiWxow25M40R3fElJkfvG1NQ5J/c76vrvw5OTndB24yhQfdoCkGdy8xjWCqCafpAYfPbD/IAzogk7MO9dlnHpQyIlG7FQhxRnJEMxDkLi841XnvuHGf0AvFwN0Ol3pTFAHR7Rc76RCWHShyG7CWRw0gjRietKtouxiCly4z91N0TUsBeBol8Yj8T71OyhRQ+rKd7LUCrt3Ul2ysidw== remmelt's mbp
EOF
chmod 0600 ~$USER/.ssh/authorized_keys
chown -R $USER:$USER ~$USER

git clone https://github.com/pimoroni/pirate-audio
cd pirate-audio/mopidy
bash install.sh
apt-get install -y mpc

echo "We're done here. Reboot!"
