# Scoundrel

A terminal-based card game built with Bubble Tea, playable over SSH.

## Rules

You explore a dungeon room by room. Each room contains 4 cards from a deck.

### Cards

- **Hearts (♥)**: Heal you for the face value (2-10)
- **Diamonds (♦)**: Weapons that deal damage equal to the face value (2-10)
- **Clubs (♣) & Spades (♠)**: Enemies that attack you (2-K, plus Ace)

### Gameplay

Each room, you:
1. **Select 3 cards in order** (press 1-4) - each card applies immediately when selected
2. **OR Run** (press R) - shuffle current room back into deck and draw 4 new cards
   - You cannot run from two rooms in a row
   - You cannot run after selecting any cards

The 4th unselected card carries over to the next room along with 3 new cards.

### Combat & Weapons

- **Aces are high** (value 14) and only appear in black suits
- **Without a weapon**: Take full enemy damage
- **With a weapon**: 
  - Initially can block any enemy
  - Damage taken = enemy value - weapon value (minimum 0)
  - After each use, weapon "degrades" - can only block enemies ≤ last enemy value - 1
  - If enemy is too strong for degraded weapon, take full damage

**Example:**
- Pick up 6♦: weapon value 6, can block any enemy
- Face 10♠: Take 10-6=4 damage, weapon now can only block ≤9
- Face J♠ (11): Can't use (11>9), take full 11 damage  
- Face 9♣: Can use! Take 9-6=3 damage, weapon now can only block ≤8

### Objective

Clear as many rooms as possible without dying! Start with 20 HP.

## Installation

```bash
go build -o scoundrel-server
```

## Deploy SSH Server

### Docker (Recommended)

**Important:** This will bind to port 22, replacing your system's SSH server. Make sure you have another way to access your server if needed, or adjust the port mapping.

Build and run using Docker Compose:

```bash
# Stop system SSH first (if running)
sudo systemctl stop sshd  # or ssh on some systems

# Start the game server
docker-compose up -d
```

Play the game:

```bash
ssh localhost
# or from another machine:
ssh your-server-address
```

**Or build and run manually:**

```bash
docker build -t scoundrel .
docker run -d -p 22:22 --name scoundrel-game scoundrel
```

**To run on a different port (keeping system SSH):**

Edit `docker-compose.yml` and change `"22:22"` to `"2222:22"`, then:

```bash
docker-compose up -d
ssh localhost -p 2222
```

### Native Deployment

#### Generate Host Key

```bash
mkdir -p .ssh
ssh-keygen -t ed25519 -f .ssh/id_ed25519 -N ""
```

#### Run the Server

```bash
./scoundrel-server
```

The server listens on port 22 by default (requires root/sudo).

**Run on different port without root:**

Edit `server.go` and change `port = 22` to `port = 2222`, then:

```bash
go build -o scoundrel-server
./scoundrel-server
```

#### Play the Game

```bash
ssh localhost -p 22
# or if using custom port:
ssh localhost -p 2222
```

Anyone can connect and play! No authentication required.

### Production Deployment

#### Option 1: Docker in Production (Replaces SSH)

```bash
# Stop and disable system SSH
sudo systemctl stop sshd
sudo systemctl disable sshd

# Run Scoundrel on port 22
docker run -d -p 22:22 --name scoundrel-game --restart unless-stopped scoundrel

# Or use Docker Compose
docker-compose up -d
```

**⚠️ Warning:** This replaces SSH access to your server. Ensure you have:
- Console access to the server (e.g., cloud provider console)
- Another remote access method, or
- Are running this on a dedicated game server

**Alternative: Run on custom port alongside SSH:**

```bash
# Keep system SSH on port 22, run Scoundrel on 2222
docker run -d -p 2222:22 --name scoundrel-game --restart unless-stopped scoundrel
```

#### Option 2: systemd service

```ini
[Unit]
Description=Scoundrel SSH Game
After=network.target

[Service]
Type=simple
User=root
WorkingDirectory=/opt/scoundrel
ExecStart=/opt/scoundrel/scoundrel-server
Restart=always

[Install]
WantedBy=multi-user.target
```

#### Option 3: Cloud Deployment

Deploy to any container platform:
- **Fly.io**: `flyctl launch` (update fly.toml to expose port 22)
- **Railway**: Connect repo, expose port
- **DigitalOcean App Platform**: Deploy as container
- **AWS ECS/Fargate**: Deploy with port 22 mapping

### Configuration

Edit `server.go` to change:
- `host`: Bind address (default: "0.0.0.0")
- `port`: SSH port (default: 22)
- Host key path (default: ".ssh/id_ed25519")

### Controls (in-game)

- `1-4`: Select card (applies immediately)
- `R`: Run from the room (shuffle and get a new room)
- `Q`: Quit game
- `N`: New game (when game over)

## Architecture

- `game.go`: Game logic and Bubble Tea UI
- `server.go`: SSH server using Charm Wish
- Each SSH connection gets its own game instance
- Anonymous access, no authentication needed

## Example

```
HP: 20 | Weapon: None (fists: 0) | Rooms: 0 | Cards left: 41

Current Room:

    [1]             [2]             [3]             [4]        
┌──────────────┐  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐
│ 5            │  │ Q            │  │ 7            │  │ A            │
│              │  │              │  │              │  │              │
│   ***   ***  │  │      **      │  │  ** **** **  │  │      **      │
│  ************│  │     ****     │  │ ************ │  │     ****     │
│  ************│  │    ******    │  │ ************ │  │    ******    │
│   ********** │  │   ********   │  │  ********** │  │   ********   │
│    ********  │  │  ********** │  │      **      │  │  ********** │
│     ******   │  │   ********   │  │     ****     │  │ ************ │
│      ****    │  │    ******    │  │    ******    │  │ ************ │
│       **     │  │     ****     │  │              │  │  ********** │
│              │  │      **      │  │              │  │      **      │
│              │  │              │  │              │  │     ****     │
│              │  │              │  │              │  │    ******    │
│            5 │  │            Q │  │            7 │  │            A │
└──────────────┘  └──────────────┘  └──────────────┘  └──────────────┘
```

Select cards 1, 2, 3:
1. Heal for 5 HP (now at 25 HP)
2. Pick up weapon with 12 damage (can block ≤14)
3. Defeat 7♣ enemy - take 0 damage, weapon now can block ≤6

Card 4 (A♠) carries to next room with 3 new cards.
