#!/usr/bin/env bash

function print_next_version() {
    local image_name=$1
    local root_vol=$2
    local chart_path=$3

    result=$(docker run -v $root_vol:/go/src -w /go/src --rm $image_name helm-release $chart_path --print --silent)
    local OS=$(echo `uname`|tr '[:upper:]' '[:lower:]')

    local sed_opt=

    case "$OS" in
        linux*) sed_opt='-r';;
        darwin) sed_opt='-E';;
     esac

     result=$(echo $result | sed $sed_opt 's/(.*)\+.*/\1/')
     result="export APP_VERSION=$result"
     echo $result
}

function set_chart_version(){
    local image_name=$1
    local root_vol=$2
    local chart_path=$3

    eval $(print_next_version $image_name $root_vol $chart_path)
    docker run -v $root_vol:/go/src -w /go/src --rm $image_name helm-release $chart_path --tag=$APP_VERSION
}