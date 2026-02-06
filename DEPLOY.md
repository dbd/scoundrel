# Deployment Guide

## Systemd Service Setup

The color issue when running as a systemd service is now fixed. The server properly detects terminal capabilities from SSH sessions.

### Install as Systemd Service

1. **Build the server:**
```bash
go build -o scoundrel-server
```

2. **Copy files to server location:**
```bash
sudo mkdir -p /opt/scoundrel
sudo cp scoundrel-server /opt/scoundrel/
sudo mkdir -p /opt/scoundrel/.ssh
sudo ssh-keygen -t ed25519 -f /opt/scoundrel/.ssh/id_ed25519 -N ""
```

3. **Create systemd service file:**
```bash
sudo nano /etc/systemd/system/scoundrel.service
```

```ini
[Unit]
Description=Scoundrel SSH Game Server
After=network.target

[Service]
Type=simple
User=root
WorkingDirectory=/opt/scoundrel
ExecStart=/opt/scoundrel/scoundrel-server
Restart=on-failure
RestartSec=5

# These environment variables help with terminal detection
Environment="TERM=xterm-256color"

[Install]
WantedBy=multi-user.target
```

4. **Enable and start the service:**
```bash
sudo systemctl daemon-reload
sudo systemctl enable scoundrel
sudo systemctl start scoundrel
```

5. **Check status:**
```bash
sudo systemctl status scoundrel
sudo journalctl -u scoundrel -f
```

### Stop System SSH (if replacing SSH)

```bash
# Stop and disable system SSH
sudo systemctl stop sshd
sudo systemctl disable sshd

# Start scoundrel on port 22
sudo systemctl start scoundrel
```

### Troubleshooting

**No colors showing:**
- The server now forces TrueColor output for all SSH sessions
- Rebuild and copy the latest version:
  ```bash
  go build -o scoundrel-server
  sudo cp scoundrel-server /opt/scoundrel/
  sudo systemctl restart scoundrel
  ```
- Make sure your SSH client supports colors (most modern terminals do)
- Try connecting with: `TERM=xterm-256color ssh yourserver.com`

**View logs:**
```bash
sudo journalctl -u scoundrel -n 50
sudo journalctl -u scoundrel -f --no-pager
```

**Restart service:**
```bash
sudo systemctl restart scoundrel
```

**File permissions:**
```bash
sudo chown -R root:root /opt/scoundrel
sudo chmod 600 /opt/scoundrel/.ssh/id_ed25519
sudo chmod 644 /opt/scoundrel/.ssh/id_ed25519.pub
```

## Docker Deployment (Alternative)

If you prefer Docker, colors work automatically:

```bash
docker-compose up -d
```

Docker handles terminal forwarding correctly by default.

## Firewall Configuration

If using port 22:

```bash
# UFW
sudo ufw allow 22/tcp

# firewalld
sudo firewall-cmd --permanent --add-service=ssh
sudo firewall-cmd --reload

# iptables
sudo iptables -A INPUT -p tcp --dport 22 -j ACCEPT
```

## Testing

Connect and verify colors work:

```bash
ssh yourserver.com
# or
ssh yourserver.com -p 22
```

You should see:
- Gold/yellow title
- Cyan headings  
- White text
- Green examples
- Red/black card suits with colored backgrounds
