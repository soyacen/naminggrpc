#!/bin/bash

# 编译项目中所有 proto 文件的脚本
# 该脚本会自动安装所需的 protoc 插件，然后编译所有 .proto 文件

set -o errexit  # 遇到错误立即退出
set -o nounset  # 使用未定义变量时报错
set -o pipefail # 管道中任一命令失败则整个管道失败

# 引入公共库
source "$(dirname "${BASH_SOURCE[0]}")/common_proto_lib.sh"

# 检查并安装 protoc-gen-go 插件
ensure_protoc_gen_go

# 检查并安装 protoc-gen-validate-go 插件
ensure_protoc_gen_validate_go

# 检查并安装 protoc-gen-goose 插件
if ! command -v protoc-gen-goose &> /dev/null; then
    echo "install protoc-gen-goose..."
    go install github.com/soyacen/goose/cmd/protoc-gen-goose@latest
    protoc-gen-goose --version
fi

# 获取脚本所在目录（项目根目录）
PROJECT_ROOT_DIR=$(get_project_root_dir)

# 切换到项目根目录
cd "${PROJECT_ROOT_DIR}"

echo "开始编译项目中所有的 proto 文件..."

# 使用公共库函数查找 proto 文件（兼容方式）
proto_files=()
while IFS= read -r file; do
    proto_files+=("$file")
done < <(find_proto_files | grep -v "^[^/]* 个 proto 文件\|^没有找到")

if [ ${#proto_files[@]} -eq 0 ]; then
    echo "没有找到 .proto 文件"
    exit 0
fi

echo "发现 ${#proto_files[@]} 个 proto 文件..."

# 编译所有proto文件，使用项目根目录和third_party作为proto_path，这样导入路径可以正确解析
echo "正在编译 proto 文件..."
protoc \
  --proto_path=. \
  --proto_path=./third_party \
  --go_out=. \
  --go_opt=paths=source_relative \
  --goose_out=. \
  --goose_opt=paths=source_relative \
  "${proto_files[@]}"

echo "编译完成！生成的文件位于对应的 proto 文件所在目录。"