#!/bin/bash

set -o errexit
set -o nounset
set -o pipefail

# 获取脚本所在目录（项目根目录）
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"

# 切换到项目根目录
cd "${SCRIPT_DIR}"

echo "开始编译项目中所有的 proto 文件..."

# 检查并安装必要的protoc插件
if ! command -v protoc-gen-go &> /dev/null; then
    echo "安装 protoc-gen-go..."
    go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
    protoc-gen-go --version
fi

# 使用find命令查找项目中所有的 .proto 文件
PROTO_FILES=()
while IFS= read -r -d '' file; do
    PROTO_FILES+=("$file")
done < <(find . -name "*.proto" -type f -print0)

if [ ${#PROTO_FILES[@]} -eq 0 ]; then
    echo "没有找到 .proto 文件"
    exit 0
fi

echo "发现 ${#PROTO_FILES[@]} 个 proto 文件..."

# 编译所有proto文件，使用项目根目录作为proto_path，这样导入路径可以正确解析
echo "正在编译 proto 文件..."
protoc \
  --proto_path=. \
  --go_out=. \
  --go_opt=paths=source_relative \
  "${PROTO_FILES[@]}"

echo "编译完成！生成的文件位于对应的 proto 文件所在目录。"