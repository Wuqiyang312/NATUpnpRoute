# **NATUpnpRoute**

## 问题

1. 需要自己编译，配置需要编译前写入。
2. ~~NAT打孔有bug，可能出现http服务端口被占用问题。~~ (修复服务端端口被占用问题)

## 功能

仅需光猫开起Upnp实现NAT打孔(前提：光猫支持upnp且运营商内网为NAT1)，使得端口可以公网访问。

## 版本

测试版v0.0.2

## 开发环境

- 编程语言： `golang`
- 开发工具： `IDEA`
- 开发环境： `debain 12 amd64`
- 编译环境： `golang 1.19`
- 测试环境： `onecloud armbain armv7l`