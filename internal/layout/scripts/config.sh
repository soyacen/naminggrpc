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

# 编译所有proto文件，使用项目根目录和third_party作为proto_path，这样导入路径可以正确解析
echo "正在编译 proto 文件..."
protoc \
  --proto_path=. \
  --proto_path=./third_party \
  --go_out=. \
  --go_opt=paths=source_relative \
  config/*.proto
echo "编译完成！"