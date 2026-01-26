#!/bin/bash

# 公共 proto 编译库
# 包含两个脚本共用的功能函数

set -o errexit  # 遇到错误立即退出
set -o nounset  # 使用未定义变量时报错
set -o pipefail # 管道中任一命令失败则整个管道失败

# 安装 protoc-gen-go 插件
ensure_protoc_gen_go() {
    if ! command -v protoc-gen-go &> /dev/null; then
        echo "install protoc-gen-go..."
        go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
        protoc-gen-go --version
    fi
}

# 安装 protoc-gen-validate-go 插件
ensure_protoc_gen_validate_go() {
    if ! command -v protoc-gen-validate-go &> /dev/null; then
        echo "install protoc-gen-validate-go..."
        go install github.com/envoyproxy/protoc-gen-validate/cmd/protoc-gen-validate-go@latest
        protoc-gen-validate-go --version
    fi
}

# 获取项目根目录路径
get_project_root_dir() {
    echo "$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
}

# 查找所有 proto 文件并输出到 stdout
find_proto_files() {
    local search_path="${1:-./api}"
    
    local proto_files=()
    while IFS= read -r -d '' file; do
        proto_files+=("$file")
    done < <(find "$search_path" -name "*.proto" -type f -print0)

    if [ ${#proto_files[@]} -eq 0 ]; then
        echo "没有找到 .proto 文件"
        exit 0
    fi

    echo "发现 ${#proto_files[@]} 个 proto 文件..."
    
    # 将文件列表输出到 stdout
    printf '%s\n' "${proto_files[@]}"
}

# 编译 proto 文件的通用函数
compile_proto_files() {
    local extra_args=("${@}")
    
    local proto_files=()
    while IFS= read -r -d '' file; do
        proto_files+=("$file")
    done < <(find ./api -name "*.proto" -type f -print0)

    echo "正在编译 proto 文件..."
    protoc \
      --proto_path=. \
      --proto_path=./third_party \
      --go_out=. \
      --go_opt=paths=source_relative \
      "${extra_args[@]}" \
      "${proto_files[@]}"

    echo "编译完成！生成的文件位于对应的 proto 文件所在目录。"
}