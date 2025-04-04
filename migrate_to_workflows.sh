#!/bin/bash

# 定义基础目录
SRC_DIR="/home/devbox/Storage/internal/logic/pipelines"
DST_DIR="/home/devbox/Storage/internal/logic/workflows"

# 创建目标目录结构
mkdir -p "$DST_DIR"/{core,api,sync,notification,integration}
mkdir -p "$DST_DIR"/core/{hooks,metrics}
mkdir -p "$DST_DIR"/api/runner
mkdir -p "$DST_DIR"/notification/{sender,receiver}
mkdir -p "$DST_DIR"/integration/apifox

# 复制并更新文件
find "$SRC_DIR" -type f -name "*.go" | while read -r file; do
    # 计算目标文件路径
    rel_path=${file#$SRC_DIR/}
    dst_file="$DST_DIR/$rel_path"
    dst_dir=$(dirname "$dst_file")
    
    # 确保目标目录存在
    mkdir -p "$dst_dir"
    
    # 复制并更新引用
    sed 's|Storage/internal/logic/pipelines|Storage/internal/logic/workflows|g' "$file" > "$dst_file"
    echo "已迁移: $file -> $dst_file"
done

# 复制README
if [ -f "$SRC_DIR/README.md" ]; then
    sed 's|pipelines/|workflows/|g' "$SRC_DIR/README.md" > "$DST_DIR/README.md"
    echo "已迁移: $SRC_DIR/README.md -> $DST_DIR/README.md"
fi

echo "迁移完成!"
