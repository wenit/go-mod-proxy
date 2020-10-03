#!/bin/bash

# 构建版本

# 当前目录
curr_dir=$(pwd)

# 获取脚本所在的目录
work_dir=$(dirname "$0")

cd ..
# 获取项目跟目录
root_dir=`pwd`

cd $work_dir


echo "[$(date)]开始打包......"
echo "[$(date)]当前工作目录：$work_dir"
echo "[$(date)]项目根目录  ：$root_dir"

version_file=$work_dir

# 获取版本号
version=$(cat $root_dir/VERSION)

#  MOD名称
module_name=`grep module ${root_dir}/go.mod|head -n 1|awk '{print $2}'`

array=(${module_name//\// })

# 可执行文件名称
app_name=$module_name
for var in ${array[@]}
do
   app_name=$var
done


# 编译包目录
release_dir=$root_dir/releases/$version
if [ ! -d "$release_dir" ]; then
    mkdir -p "$release_dir"
fi

# 架构
app_arch=amd64

# 打包
build_app(){
    app_os=$1
    echo "[$(date)]开始打包[${app_name}_${version}_${app_os}_${app_arch}]..."

    # 编译包目录
    release_os_dir=${release_dir}/${app_os}
    if [ ! -d "$release_os_dir" ]; then
        mkdir -p "$release_os_dir"
    fi

    app_suffix=""
    if [ "$app_os" = "windows" ];then
        app_suffix=".exe"
    fi
    CGO_ENABLED=0 GOARCH=$app_arch GOOS=$app_os go build -o ${release_os_dir}/${app_name}_${version}_${app_os}_${app_arch}${app_suffix} -ldflags="-w -s -X '${module_name}/internal/version.version=${version}' -X '${module_name}/internal/version.buildDate=$(date '+%Y-%m-%d %H:%M:%S')'"
}

# 全平台打包
build_app_all_platforms() {
    # windows打包
    build_app windows

    # linux打包
    build_app linux

    # darwin打包(mac os)
    build_app darwin
}

build_app_module(){
    #-------------------client-------------------
    # 可执行文件名称
    app_name=$app_name
    cd $root_dir/cmd/proxy
    echo "[$(date)]==================[$app_name]打包开始=================="
    build_app_all_platforms
    echo "[$(date)]==================[$app_name]打包结束=================="
    echo ""
}

# 清理打包目录
app_clean(){
    echo "[$(date)]清理打包目录：[$release_dir]"
    rm -rf $release_dir
}

# 清理打包目录
app_clean

# 打包
build_app_module



cd $work_dir
echo "[$(date)]所有模块打包完成."






