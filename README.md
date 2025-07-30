# GScope

**GitHub Repository Analytics for Engineering Managers & Team Leaders**

GScope is a comprehensive analytics platform designed specifically for engineering managers and team leaders who need data-driven insights into their development teams. Transform your GitHub repositories into actionable performance metrics, track team contributions, and make informed decisions about resource allocation and team dynamics.

## Why GScope?

- **Unified Analytics**: Aggregate data from multiple GitHub repositories into a single dashboard
- **Team Performance**: Track individual and team contributions with detailed statistics
- **Smart Analysis**: Automatically analyze commits, pull requests, and code changes
- **Email Merging**: Consolidate multiple developer identities for accurate attribution
- **Flexible Reporting**: Daily, weekly, monthly, and yearly insights with customizable scoring
- **Automated Updates**: Schedule regular data synchronization with GitHub

## Key Benefits for Engineering Leaders

- **Team Performance Visibility**: Get clear insights into individual and team productivity metrics
- **Resource Planning**: Make data-driven decisions about team capacity and project allocation
- **Performance Reviews**: Objective data for developer evaluations and career discussions
- **Bottleneck Identification**: Spot workflow issues and collaboration gaps early
- **Cross-Project Insights**: Manage and compare performance across multiple teams and repositories
- **Executive Reporting**: Generate comprehensive reports for stakeholder updates

## Quick Start

1. **Clone and Setup**
   ```bash
   git clone https://github.com/alimgiray/gscope
   cd gscope
   go mod download
   ```

2. **Configure Environment**
   ```bash
   # Option 1: Use the setup script
   ./setup_env.sh
   
   # Option 2: Manual setup
   cp env.example .env
   # Edit .env with your configuration
   ```

3. **Configure GitHub OAuth App**
   
   **Important**: You need to create your own GitHub OAuth App because the callback URL must match your deployment environment.
   
   **For Local Development:**
   - Go to [GitHub OAuth Apps](https://github.com/settings/developers)
   - Click "New OAuth App"
   - Fill in the details:
     - **Application name**: `GScope` (or any name you prefer)
     - **Homepage URL**: `http://localhost:8080`
     - **Authorization callback URL**: `http://localhost:8080/auth/github/callback`
   - Click "Register application"
   - Copy the **Client ID** and **Client Secret** to your `.env` file
   
   **For Production Deployment:**
   - Create a new OAuth App with your production domain
   - Set **Authorization callback URL** to: `https://yourdomain.com/auth/github/callback`
   - Update your environment variables accordingly
   
   **OAuth App Scopes Required:**
   - `user:email` - Access to user's email addresses
   - `read:user` - Read access to user profile data  
   - `read:org` - Read access to organization membership
   - `repo` - Full access to repositories (includes PRs, issues, etc.)

4. **Run GScope**
   ```bash
   go run cmd/server/main.go
   ```

5. **Access Dashboard**
   Open `http://localhost:8080` and authenticate with GitHub

## Configuration

GScope uses environment variables for configuration. Copy `env.example` to `.env` and customize the values:

### Server Configuration
```bash
PORT=8080                    # Server port
GIN_MODE=release            # Gin mode (debug/release)
READ_TIMEOUT=15             # HTTP read timeout in seconds
WRITE_TIMEOUT=15            # HTTP write timeout in seconds
```

### Database Configuration
```bash
DB_PATH=./gscope.db         # SQLite database file path
```

### GitHub OAuth Configuration
```bash
GITHUB_CLIENT_ID=your_github_client_id_here
GITHUB_CLIENT_SECRET=your_github_client_secret_here
GITHUB_CALLBACK_URL=http://localhost:8080/auth/github/callback
```

### Session Configuration
```bash
SESSION_SECRET=your-super-secret-session-key-change-this-in-production
```

### Environment Variables Priority
1. **Environment variables** (highest priority)
2. **`.env` file** (loaded automatically)
3. **Default values** (fallback)

### Production Deployment
For production deployment, set environment variables directly:
```bash
export PORT=80
export GIN_MODE=release
export DB_PATH=/var/lib/gscope/gscope.db
export SESSION_SECRET=your-super-secure-random-string
export GITHUB_CLIENT_ID=your_production_client_id
export GITHUB_CLIENT_SECRET=your_production_client_secret
export GITHUB_CALLBACK_URL=https://yourdomain.com/auth/github/callback
```

### GitHub API Rate Limits
GScope uses GitHub's OAuth App tokens which have a rate limit of **5,000 requests per hour**. For high-volume usage or large repositories, consider:

- **GitHub Personal Access Tokens (PATs)**: Same rate limit but more suitable for server-to-server calls
- **GitHub Apps**: Higher rate limits for enterprise use cases
- **Enterprise GitHub**: Significantly higher rate limits

The application includes built-in rate limiting handling with exponential backoff and automatic retry logic.

## Features

- **Repository Management**: Clone, track, and analyze multiple GitHub repositories
- **Commit Analysis**: Detailed commit statistics with file-level insights
- **Pull Request Tracking**: Monitor PR reviews, merges, and collaboration patterns
- **People Analytics**: Track individual developer contributions and team dynamics
- **Custom Scoring**: Configure scoring systems based on your team's priorities
- **Automated Jobs**: Background workers for data synchronization and analysis
- **Export Reports**: Generate reports in multiple time ranges and formats

## Tech Stack

- **Backend**: Go 1.24 with Gin web framework
- **Database**: SQLite with optimized performance settings
- **Authentication**: GitHub OAuth integration
- **Workers**: Concurrent job processing system
- **Frontend**: Server-side rendered HTML with Tailwind CSS

**Perfect for engineering managers, team leads, and CTOs who need objective insights to build high-performing development teams.**