Use the new raspbian installer, setting wifi pw, ssh key, username, etc.

```bash

sudo apt update
sudo apt install -y htop tmux unattended-upgrades fail2ban openssh-server git
sudo apt upgrade -y

cat <<EOF > /etc/apt/apt.conf.d/02periodic
APT::Periodic::Enable "1";
APT::Periodic::Update-Package-Lists "1";
APT::Periodic::Download-Upgradeable-Packages "1";
APT::Periodic::Unattended-Upgrade "1";
APT::Periodic::AutocleanInterval "1";
APT::Periodic::Verbose "2";
EOF

cat <<EOF > /etc/ssh/sshd_config
X11Forwarding no
PrintMotd no
AcceptEnv LANG LC_*
Subsystem sftp /usr/lib/openssh/sftp-server
AllowUsers pi
ChallengeResponseAuthentication no
PasswordAuthentication no
UsePAM no
PermitRootLogin no
EOF

git clone https://github.com/pimoroni/pirate-audio
cd pirate-audio/mopidy
bash install.sh
apt install -y mpc mpd mopidy-mpd

sudo python3 -m pip install Mopidy-MPD

sudo vi /etc/mopidy/mopidy.conf

# add config under [spotify], see 1pw discotoilet
# or get tokens here: https://mopidy.com/ext/spotify/
sudo systemctl restart mopidy

# see actual mopidy config
sudo mopidyctl config



```

http://discotoilet.local:6680/iris/settings

- spotify
- log in

You should be able to play tracks now.

You need to authenticate both mopidy and iris to spotify using oauth

https://www.spotify.com/nl/account/apps/

See here for approved apps, should have both `Mopidy extensions` and `Iris`.
