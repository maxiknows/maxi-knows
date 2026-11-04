# Fail2ban

## Purpose

Fail2ban protects our Azure VM against repeated SSH login attempts by temporarily blocking suspicious IP addresses.

We checked the SSH logs and discovered several automated login attempts using invalid usernames.

## Installation

Our VM runs AlmaLinux 9.8, so we installed Fail2ban using `dnf`.

```bash
sudo dnf config-manager --set-enabled crb
sudo dnf install -y epel-release
sudo dnf install -y fail2ban fail2ban-firewalld fail2ban-systemd
```

We enabled firewalld:

```bash
sudo systemctl enable --now firewalld
```

## Configuration

File: `/etc/fail2ban/jail.local`

```ini
[sshd]
enabled = true
backend = systemd
maxretry = 5
findtime = 10m
bantime = 1h
```

This bans an IP for one hour after five matching failed SSH authentication attempts within ten minutes.

We enabled Fail2ban:

```bash
sudo systemctl enable --now fail2ban
```

## Monitoring

| Command | Purpose |
|---|---|
| `sudo fail2ban-client status sshd` | See detected failures and banned IPs |
| `sudo journalctl -u sshd -n 30` | See recent SSH activity |
| `sudo journalctl -u sshd --since "7 days ago" \| grep -c "Invalid user"` | Count invalid username attempts |
| `sudo journalctl -u sshd --since "7 days ago" \| grep "Accepted"` | See successful SSH logins |
| `sudo systemctl status fail2ban` | Check whether Fail2ban is running |

## Result

Fail2ban is installed and the SSH jail is active.

SSH password authentication and root login were already disabled. Fail2ban adds another layer of protection.

**Note:** Fail2ban protects SSH access to the VM, not the login functionality in the Maxi Knows application.
