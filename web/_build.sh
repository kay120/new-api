#!/bin/bash
export PATH="$HOME/.nvm/versions/node/v22.22.0/bin:$PATH"
cd /Users/kay/code/work/ai-gateway/new-api/web
exec npx vite build --mode development 2>&1 | tail -40
