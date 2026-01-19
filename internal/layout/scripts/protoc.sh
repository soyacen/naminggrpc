#!/bin/bash

set -o errexit
set -o nounset
set -o pipefail

#!/bin/sh

set -o errexit
set -o nounset
set -o pipefail

if [ ! $(command -v protoc-gen-go) ]
then
	go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
	protoc-gen-go --version
fi

if [ ! $(command -v protoc-gen-go-grpc) ]
then
	go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
	protoc-gen-go-grpc --version
fi

if [ ! $(command -v protoc-gen-goose) ]
then
	go install github.com/soyacen/goose/cmd/protoc-gen-goose@latest
	protoc-gen-goose --version
fi

if [ ! $(command -v protoc-gen-validate-go) ]
then
	go install github.com/envoyproxy/protoc-gen-validate/cmd/protoc-gen-validate-go@latest
fi

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

# 使用find命令查找项目中所有的 .proto 文件，排除pkg/layout/third_party和internal/layout/third_party目录
PROTO_FILES=()
while IFS= read -r -d '' file; do
    if [[ "$file" != *"third_party"* ]]; then
        PROTO_FILES+=("$file")
    fi
done < <(find . -name "*.proto" -type f -print0)

if [ ${#PROTO_FILES[@]} -eq 0 ]; then
    echo "没有找到 .proto 文件"
    exit 0
fi

echo "发现 ${#PROTO_FILES[@]} 个 proto 文件..."

# 编译所有proto文件，使用项目根目录和third_party作为proto_path，这样导入路径可以正确解析
echo "正在编译 proto 文件..."
protoc \
  --proto_path=. \
  --proto_path=./third_party \
  --go_out=. \
  --go_opt=paths=source_relative \
  "${PROTO_FILES[@]}"

echo "编译完成！生成的文件位于对应的 proto 文件所在目录。"