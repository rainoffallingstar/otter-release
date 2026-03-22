#!/bin/bash
set -e

if [ -z "$1" ]; then
  echo "Usage: $0 <version>"
  echo "Example: $0 1.0.0"
  exit 1
fi

VERSION=$1

# 验证版本格式
if ! [[ $VERSION =~ ^[0-9]+\.[0-9]+\.[0-9]+$ ]]; then
  echo "Error: Version must be in format X.Y.Z"
  exit 1
fi

# 检查是否已存在标签
if git tag | grep -q "v${VERSION}"; then
  echo "Error: Tag v${VERSION} already exists"
  exit 1
fi

# 构建和测试
echo "Running release checks..."
bash scripts/verify_release_evidence.sh
go test ./...
go vet ./...
./scripts/build.sh $VERSION

# 创建标签
echo "Creating tag v${VERSION}..."
git tag -a "v${VERSION}" -m "Release v${VERSION}"

# 推送到远程
echo "Pushing to remote..."
git push origin "v${VERSION}"

echo "Release v${VERSION} created and pushed!"
echo "GitHub Actions will build and publish the release."
