#!/bin/bash
# ========================================
# 环境就绪检测脚本
# 检测 Makefile 所需的文件及目录是否齐全
# ========================================



RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

ICON_OK="✅"
ICON_FAIL="❌"
ICON_WARN="⚠️"
ICON_INFO="➡️"

BASE_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

total_required=0
pass_required=0
fail_required=0

total_recommended=0
pass_recommended=0
fail_recommended=0

echo "========================================"
echo "  环境就绪检测脚本"
echo "  工作目录: $BASE_DIR"
echo "========================================"
echo ""

# --------------------------------------------------
# 检测函数
# --------------------------------------------------
check_dir() {
    local path="$1"
    local label="$2"

    if [ -d "$BASE_DIR/$path" ]; then
        echo -e "  ${ICON_OK} ${label}                      存在"
        ((pass_required++))
        return 0
    else
        echo -e "  ${ICON_FAIL} ${label}                      缺失 ${ICON_WARN}"
        ((fail_required++))
        return 1
    fi
}

check_file() {
    local path="$1"
    local label="$2"
    local category="${3:-required}"  # required | recommended

    if [ -f "$BASE_DIR/$path" ]; then
        echo -e "  ${ICON_OK} ${label}            存在"
        if [ "$category" = "required" ]; then
            ((pass_required++))
        else
            ((pass_recommended++))
        fi
        return 0
    else
        if [ "$category" = "required" ]; then
            echo -e "  ${ICON_FAIL} ${label}            缺失 ${ICON_WARN}"
            ((fail_required++))
        else
            echo -e "  ${YELLOW}${ICON_WARN}${NC} ${label}            推荐但未找到"
            ((fail_recommended++))
        fi
        return 1
    fi
}

# --------------------------------------------------
# 1. 目录检查
# --------------------------------------------------
echo "[目录检查]"

check_dir "deploy"           "deploy/"
check_dir "dtm"              "dtm/"
check_dir "order-service"    "order-service/"
check_dir "stock-service"    "stock-service/"

echo ""

# --------------------------------------------------
# 2. 必需文件检查
# --------------------------------------------------
echo "[文件检查]"

check_file "docker-compose.local.yaml"   "docker-compose.local.yaml"   "required"
check_file "deploy/init-dtm-db.sh"       "deploy/init-dtm-db.sh"       "required"
check_file "dtm/Dockerfile.dtm"          "dtm/Dockerfile.dtm"          "required"
check_file "order-service/Dockerfile"    "order-service/Dockerfile"    "required"
check_file "stock-service/Dockerfile"    "stock-service/Dockerfile"    "required"

echo ""

# --------------------------------------------------
# 3. 推荐文件检查（非必需）
# --------------------------------------------------
echo "[推荐文件检查]"

check_file "order-service/go.mod"                   "order-service/go.mod"                   "recommended"
check_file "order-service/go.sum"                   "order-service/go.sum"                   "recommended"
check_file "order-service/config.yaml"              "order-service/config.yaml"              "recommended"
check_file "order-service/start-order-service.sh"   "order-service/start-order-service.sh"   "recommended"

echo ""

# --------------------------------------------------
# 汇总
# --------------------------------------------------
echo "========================================"
echo "  检测汇总"
echo "========================================"
echo ""
total_required=$((pass_required + fail_required))
echo "  必需项: $pass_required / $total_required 通过"
total_r=$((pass_recommended + fail_recommended))
if [ "$total_r" -gt 0 ]; then
    echo "  推荐项: $pass_recommended / $total_r 通过"
fi
echo ""

if [ "$fail_required" -eq 0 ]; then
    echo -e "  ${ICON_OK} 所有必需文件及目录已就绪，可以执行 Makefile 目标。"
    exit 0
else
    echo -e "  ${ICON_FAIL} 有 $fail_required 项必需文件/目录缺失，请先补充后再执行 Makefile。"
    exit 1
fi