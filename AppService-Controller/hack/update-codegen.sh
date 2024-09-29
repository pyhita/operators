#!/usr/bin/env bash

set -o errexit
set -o nounset
set -o pipefail

# 此处为变量，是 generate-groups.sh 的文件夹
CODEGEN_PKG="/Users/kante.yang/workspaces/go/pkg/mod/k8s.io/code-generator@v0.31.1"
# generate the code with:
# --output-base    because this script should also be able to run inside the vendor dir of
#                  k8s.io/kubernetes. The output-base is needed for the generators to output into the vendor dir
#                  instead of the $GOPATH directly. For normal projects this can be dropped.
# 执行代码生成功能，生成 deepcopy,client,informer,lister
# AppService-Controller 是 go.mod 中的项目地址
# stable:v1beta1 根据自己设计的组与版本填写
# output-base 为输出目录
# go-header-file 为每个文件添加一个头文件，就是开源协议
#generate-groups.sh
"${CODEGEN_PKG}/kube_codegen.sh" "deepcopy,client,informer,lister" \
  AppService-Controller/pkg/generated \
  AppService-Controller/pkg/apis \
  stable:v1beta1 \
  --output-base ../pkg/ \
  --go-header-file ./boilerplate.go.txt

# To use your own boilerplate text append:
#   --go-header-file "${SCRIPT_ROOT}"/hack/custom-boilerplate.go.txt