#!/bin/bash

# 构建 Docker 镜像
docker build -t teenet-signature-tool:latest .

# 导出并压缩
docker save teenet-signature-tool:latest | gzip > teenet-signature-tool.tar.gz

echo "✅ 完成: teenet-signature-tool.tar.gz"