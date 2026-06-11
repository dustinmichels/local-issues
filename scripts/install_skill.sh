#!/bin/bash

echo "Installing skill to agents..."
mkdir -p ~/.agents/skills
rm -rf ~/.agents/skills/local-issues
echo "> cp -r skill/local-issues ~/.agents/skills/"
cp -r skill/local-issues ~/.agents/skills/

echo "Symlinking to .claude..."
mkdir -p ~/.claude/skills/
rm -f ~/.claude/skills/local-issues
echo "> ln -s ~/.agents/skills/local-issues ~/.claude/skills/local-issues"
ln -s ~/.agents/skills/local-issues ~/.claude/skills/local-issues

echo "Done!"