#!/bin/bash

set -o errexit
set -o nounset
set -o pipefail

# 获取脚本所在目录
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

# 切换到项目根目录
cd "${SCRIPT_DIR}"

echo "开始编译 config.proto 文件..."

# 检查并安装必要的protoc插件
if [ ! $(command -v protoc-gen-go) ]; then
    echo "安装 protoc-gen-go..."
    go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
    protoc-gen-go --version
fi

# 编译redisx包中的protobuf文件
PROTO_FILE="config.proto"
echo "正在编译 ${PROTO_FILE}..."

protoc \
  --proto_path=. \
  --go_out=. \
  --go_opt=paths=source_relative \
  "${PROTO_FILE}"

echo "编译完成！生成的文件位于 $SCRIPT_DIR 目录下。"