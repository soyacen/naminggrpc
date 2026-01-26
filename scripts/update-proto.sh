#!/bin/bash

# Script to copy proto files from pkg to internal/layout/third_party/pkg preserving directory structure

set -o errexit  # 遇到错误立即退出
set -o nounset  # 使用未定义变量时报错
set -o pipefail # 管道中任一命令失败则整个管道失败

# 获取脚本所在目录（项目根目录）
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT_DIR="$(dirname "${SCRIPT_DIR}")"

# 定义源目录和目标目录
SRC_DIR="${PROJECT_ROOT_DIR}/pkg"
DST_DIR="${PROJECT_ROOT_DIR}/internal/layout/third_party/pkg"

# 创建目标目录
mkdir -p "${DST_DIR}"

# 查找并复制所有 .proto 文件，保留目录结构
find "${SRC_DIR}" -name "*.proto" -type f | while read -r proto_file; do
  # 计算相对路径
  rel_path="${proto_file#"${SRC_DIR}/"}"
  dst_file="${DST_DIR}/${rel_path}"
  
  # 创建目标文件的目录
  mkdir -p "$(dirname "${dst_file}")"
  
  # 复制文件
  cp "${proto_file}" "${dst_file}"
  echo "Copied: ${rel_path}"
done

echo "All proto files copied successfully!"