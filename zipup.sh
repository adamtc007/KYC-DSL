#!/bin/bash

# zipup.sh - Create a clean zip of the repository for AI chat uploads
# This mirrors exactly what's tracked in git (excluding build artifacts, IDE files, etc.)

set -e

echo "🗂️  Creating clean zip of KYC-DSL repository..."

# Remove existing zip if it exists
rm -f kyc-dsl.zip

# Create temporary directory
TEMP_DIR=$(mktemp -d)
REPO_NAME="KYC-DSL"

echo "📁 Creating temporary copy in: $TEMP_DIR/$REPO_NAME"

# Copy only git-tracked files to temp directory
mkdir -p "$TEMP_DIR/$REPO_NAME"

# Use git ls-files to get exactly what's in the repository
git ls-files | while read -r file; do
    # Create directory structure if needed
    mkdir -p "$TEMP_DIR/$REPO_NAME/$(dirname "$file")"
    # Copy the file
    cp "$file" "$TEMP_DIR/$REPO_NAME/$file"
done

echo "📦 Files included in zip:"
git ls-files | sed 's/^/  ✅ /'

# Store current directory
CURRENT_DIR="$(pwd)"

# Create the zip file
cd "$TEMP_DIR"
zip -r "$CURRENT_DIR/kyc-dsl.zip" "$REPO_NAME"

# Return to original directory
cd "$CURRENT_DIR"

# Cleanup
rm -rf "$TEMP_DIR"

echo ""
echo "✅ Successfully created kyc-dsl.zip"
echo "📊 Zip file size: $(du -h kyc-dsl.zip | cut -f1)"
echo "📋 Contents: $(git ls-files | wc -l | tr -d ' ') files, $(git ls-files | xargs wc -l | tail -1 | awk '{print $1}') lines of code"
echo ""
echo "🚀 Ready to upload to Gemini, ChatGPT, or Claude chat sessions!"
echo "   The zip contains only clean source code - no build artifacts or IDE files."