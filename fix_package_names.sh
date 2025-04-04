#!/bin/bash

# 遍历所有服务目录
for dir in internal/logic/*/; do
    if [ -d "$dir" ]; then
        service_name=$(basename "$dir")
        # 修改每个 .go 文件的包名
        for file in "$dir"*.go; do
            if [ -f "$file" ]; then
                # 将 package serviceXXX 替换为 package serviceXXXlogic
                sed -i "s/^package ${service_name}/package ${service_name}logic/" "$file"
            fi
        done
    fi
done

# 处理特殊目录
for dir in internal/logic/{tools,workflows,pipelines}/; do
    if [ -d "$dir" ]; then
        dir_name=$(basename "$dir")
        # 修改每个 .go 文件的包名
        find "$dir" -type f -name "*.go" -exec sed -i "s/^package ${dir_name}/package ${dir_name}logic/" {} \;
    fi
done
