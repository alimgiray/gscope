#!/bin/bash

# GScope Environment Setup Script

echo "🚀 Setting up GScope environment..."

# Check if .env file already exists
if [ -f ".env" ]; then
    echo "⚠️  .env file already exists. Backing up to .env.backup"
    cp .env .env.backup
fi

# Copy env.example to .env
if [ -f "env.example" ]; then
    cp env.example .env
    echo "✅ Created .env file from env.example"
    echo ""
    echo "📝 Please edit .env file with your configuration:"
    echo "   - Set GITHUB_CLIENT_ID and GITHUB_CLIENT_SECRET from GitHub OAuth app"
    echo "   - Change SESSION_SECRET to a secure random string"
    echo "   - Adjust PORT if needed (default: 8080)"
    echo ""
    echo "🔗 GitHub OAuth App Setup:"
    echo "   1. Go to https://github.com/settings/developers"
    echo "   2. Create a new OAuth App"
    echo "   3. Set Authorization callback URL to: http://localhost:8080/auth/github/callback"
    echo "   4. Copy Client ID and Client Secret to .env file"
    echo ""
    echo "🎯 Next steps:"
    echo "   1. Edit .env file with your values"
    echo "   2. Run: go run cmd/server/main.go"
    echo "   3. Visit: http://localhost:8080"
else
    echo "❌ env.example file not found!"
    exit 1
fi 