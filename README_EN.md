# Webhook Telegram Bot

![Production Ready](https://img.shields.io/badge/production-ready-green)
![Docker](https://img.shields.io/badge/docker-supported-blue)
![AI Powered](https://img.shields.io/badge/AI-GPT--4.1--nano-orange)

Professional Telegram bot for automatic notifications about new posts from Discourse forums to Telegram groups with AI-generated summaries.

## ✨ Features

### 🔗 Discourse Integration
- Processing webhooks for topic and post creation
- Support for all types of Discourse content
- Combining topic and post data into one message
- Webhook signature verification for security

### 🤖 AI Content Analysis
- **OpenAI GPT-4.1-nano** for generating smart summaries
- Post content analysis with context awareness
- HTML tag cleanup for better analysis
- Special templates for different content types

### 📱 Flexible Telegram Delivery
- Send to main chat or specific topics
- **Category mapping** to different Telegram topics
- Support for emoji prefixes for user roles
- Smart message formatting with HTML

### 🎯 Filtering and Control
- **Monitor specific categories** or all at once
- **Ignore unwanted categories**
- **Block bots** (discobot, chatbot and others)
- **Premium sections** with subscription notifications

### 🚀 Production Ready
- Docker containerization with healthcheck
- GitHub Actions for automatic builds
- Systemd integration for autostart
- Logging and monitoring
- Graceful shutdown and error handling

## 📋 Message Format

```
👤 👑 Administrator created a new post: Topic Title

📋 Author shares a solution to Docker container configuration problem and explains step-by-step process of fixing database connection errors.

🔗 Topic link (https://forum.example.com/t/topic/123)

🏷 Tags: #docker, #database, #troubleshooting

💎 This section is available by subscription only.
```

### Role Prefixes
- 👑 Administrator
- 🛡️ Moderator  
- ⭐ Staff
- 🔥 Leader
- 👤 Regular user

## ⚙️ Configuration (.env)

### 📡 Telegram Settings
```bash
# Basic bot settings
TELEGRAM_BOT_TOKEN=1234567890:ABC-DEF1234567890          # Token from @BotFather
TELEGRAM_CHAT_ID=-1001234567890                          # Group/channel ID  
TELEGRAM_THREAD_ID=0                                     # Topic ID (0 = main chat)
```

### 🔗 Webhook Settings
```bash
WEBHOOK_SECRET=your_super_secret_key_here                # Secret key for signature verification
WEBHOOK_PORT=8080                                        # Webhook server port
WEBHOOK_PATH=/webhook                                     # Endpoint path
WEBHOOK_DOMAIN=https://your-server.com                   # Server domain (optional)
```

### 🤖 AI Settings
```bash
OPENAI_API_KEY=sk-xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx  # OpenAI API key
OPENAI_MODEL=gpt-4.1-nano                                   # GPT model (recommended gpt-4.1-nano)
```

### 🏷 Categories and Filtering
```bash
BASE_URL=https://your-forum.com                          # Your forum address

# Category monitoring (empty = all categories)
MONITORED_CATEGORIES=1,2,3,4,5                           # Category IDs to track
IGNORED_CATEGORIES=10,11,12                               # Category IDs to ignore (priority)

# User ignoring
IGNORED_USERS=-2,-1,1234                                  # User IDs to ignore
                                                          # -2 = discobot, -1 = system, etc.

# Premium sections
PREMIUM_CATEGORIES=4,5,6                                  # Category IDs with subscription
```

### 📍 Category Mapping to Telegram Topics
```bash
# Programming to topic 1
TELEGRAM_THREAD_ID_1=123456                              # Topic message ID
THREAD_CATEGORIES_1=1,2,3                                # Categories: Go, Python, JavaScript

# Design to topic 2  
TELEGRAM_THREAD_ID_2=234567
THREAD_CATEGORIES_2=4,5                                  # Categories: UI/UX, Graphics

# Business to topic 3
TELEGRAM_THREAD_ID_3=345678
THREAD_CATEGORIES_3=6,7,8                                # Categories: Marketing, Sales, Strategy

# Hardware to topic 4
TELEGRAM_THREAD_ID_4=456789  
THREAD_CATEGORIES_4=9,10                                 # Categories: Hardware, Reviews

# Off-topic to topic 5
TELEGRAM_THREAD_ID_5=567890
THREAD_CATEGORIES_5=11,12,13                             # Categories: General, Random, Fun
```

## 🚀 Installation

### Quick Installation (Recommended)
```bash
# Download and run installer
curl -fsSL https://raw.githubusercontent.com/dignezzz/webhook_tg_bot/main/install.sh -o install.sh
chmod +x install.sh
./install.sh install
```

### Bot Management
```bash
whtg                    # Interactive menu
whtg install            # Install bot  
whtg update             # Update to latest version
whtg start              # Start bot
whtg stop               # Stop bot
whtg restart            # Restart bot
whtg status             # Show status
whtg logs               # Show logs
whtg uninstall          # Remove bot
```

### Docker Compose (Manual Installation)
```yaml
version: '3.8'
services:
  webhook-bot:
    image: ghcr.io/dignezzz/webhook_tg_bot:latest
    container_name: webhook_tg_bot
    restart: unless-stopped
    ports:
      - "8080:8080"
    env_file:
      - .env
    healthcheck:
      test: ["CMD", "curl", "-f", "http://localhost:8080/health"]
      interval: 30s
      timeout: 10s
      retries: 3
```

## 🔧 Discourse Setup

### 1. Creating Webhook
1. Go to **Admin → API → Webhooks**
2. Click **New Webhook**
3. Fill in parameters:

```
Payload URL: http://your-server.com:8080/webhook
Content Type: application/json  
Secret: your_webhook_secret_from_env
Which events: Topic Event + Post Event
```

### 2. Getting Category IDs
Go to admin panel: `https://your-forum.com/admin/customize/site_texts`
Or check category URL: `https://your-forum.com/c/category-name/5` (where 5 is the ID)

### 3. Getting User IDs  
Go to admin panel: `https://your-forum.com/admin/users/USERNAME`
URL will show ID: `https://your-forum.com/admin/users/123/username`

## 📱 Telegram Setup

### 1. Creating Bot
1. Message [@BotFather](https://t.me/botfather)
2. Use command `/newbot`
3. Copy token to `TELEGRAM_BOT_TOKEN`

### 2. Group Setup
1. Create group or channel
2. Add bot as administrator
3. Get Chat ID:
   - Add [@userinfobot](https://t.me/userinfobot) to group
   - Or use API: `https://api.telegram.org/bot<TOKEN>/getUpdates`

### 3. Topics Setup (Optional)
1. Enable topics in group settings
2. Create topics for different categories
3. Get Thread ID for each topic:
   - Forward message from topic to bot
   - Check logs to get Thread ID

## 🔍 Monitoring and Debugging

### Status Check
```bash
# Healthcheck endpoint
curl http://localhost:8080/health

# Status via management script
whtg status
```

### Logs
```bash
# Via management script  
whtg logs

# Direct via Docker
docker-compose logs -f webhook-bot

# System logs
journalctl -u webhook-tg-bot -f
```

### Webhook Testing
```bash
curl -X POST http://localhost:8080/webhook \
  -H "Content-Type: application/json" \
  -H "X-Discourse-Event-Signature: sha256=..." \
  -d '{
    "topic": {
      "id": 1,
      "title": "Test Topic", 
      "category_id": 1,
      "created_by": {"username": "test"}
    }
  }'
```

## 📊 Architecture

```
┌─────────────────┐    ┌──────────────────┐    ┌─────────────────┐
│   Discourse     │───▶│  Webhook Server  │───▶│   Telegram Bot  │
│    Forum        │    │                  │    │                 │
└─────────────────┘    └──────────────────┘    └─────────────────┘
                              │                          │
                              ▼                          ▼
                       ┌──────────────┐         ┌─────────────────┐
                       │ OpenAI GPT   │         │ Telegram Groups │
                       │  (Summary)   │         │   & Topics      │
                       └──────────────┘         └─────────────────┘
```

### Components
- **HTTP Server**: Processing webhooks from Discourse
- **Storage**: Temporary storage for data merging
- **AI Provider**: Summary generation via OpenAI
- **Telegram Bot**: Sending messages to groups/topics
- **Config**: Flexible configuration system

## 🔐 Security

### Recommendations
- ✅ Use strong secret key (`WEBHOOK_SECRET`)
- ✅ Configure HTTPS in production  
- ✅ Restrict webhook port access via firewall
- ✅ Regularly update Docker image (`whtg update`)
- ✅ Monitor logs for suspicious activity

### Signature Verification
Bot automatically verifies HMAC-SHA256 signature of each webhook to protect against fake requests.

## 🔄 Auto-updates

### CI/CD Pipeline
- GitHub Actions automatically builds new Docker image on each commit
- Images are published to GitHub Container Registry
- Use `whtg update` to update to latest version

### Rollback
```bash
# Rollback to previous version  
docker-compose down
docker pull ghcr.io/dignezzz/webhook_tg_bot:previous-tag
docker-compose up -d
```

## 🤝 Support

- **GitHub**: [dignezzz/webhook_tg_bot](https://github.com/dignezzz/webhook_tg_bot)
- **Issues**: [Report an issue](https://github.com/dignezzz/webhook_tg_bot/issues)
- **Discussions**: [Discussions and questions](https://github.com/dignezzz/webhook_tg_bot/discussions)

## 📄 License

MIT License - see [LICENSE](LICENSE) file.

---

**💡 Tip**: Start with basic configuration, then gradually add category mapping and filters as needed.
