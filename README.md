# UnblockNeteaseMusic
解锁网易云音乐客户端变灰歌曲 (Golang)

[配套LUCI](https://github.com/7rah/luci-app-unblockneteasemusic)
# 特性
* 就是快
* 较为精准的歌曲匹配
* 全平台支持
* 低内存、高效率
* 暂时支持酷狗、酷我 、咪咕的源（酷我、咪咕支持无损音乐）
* 学习过程中的产物，随缘更新

# 运行
先为自己生成证书（windows需要自己下载openssl）（为了你的安全，请务必自己生成证书）
```
./createCertificate.sh
```

运行程序（由于m=1时 会自动修改hosts生效 所以需要sudo）
```
sudo ./UnblockNeteaseMusic
```

具体参数说明
```
./UnblockNeteaseMusic -h
    -c string
      	specify server cert,such as : "server.crt" (default "./server.crt")
    -e	replace song url
    -k string
      	specify server cert key ,such as : "server.key" (default "./server.key")
    -m int
      	specify running mode（1:hosts） ,such as : "1" (default 1)
    -o string
      	specify server source,such as : "kuwo:kugou" (default "kuwo:kugou")
    -p int
      	specify server port,such as : "80" (default 80)
    -sp int
      	specify server tls port,such as : "443" (default 443)
    -v	display version info

```

重要提示：

应用通过本机dns获取域名ip，请注意本地hosts文件

受限于歌曲md5的计算时间，耐心等待一会儿再点击下载歌曲吧

网易云APP能用就别升级，不保证新版本可以使用

IOS信任证书步骤：
1. 安装证书--设置-描述文件-安装
2. 通用-关于本机-证书信任设置-启动完全信任

已知：
1. windows版本的网易云音乐需要在应用内 设置代理 Http地址为「HttpProxy」下任意地址 端口 80
2. Linux 客户端 (1.2 版本以上需要在终端启动网易云客户端时增加 --ignore-certificate-errors 参数)
3. ios客户端需要信任根证书且运行UnblockNeteaseMusic时 加上 -e 参数
4. android客户端使用咪咕源下载歌曲时需要运行UnblockNeteaseMusic时 加上 -e 参数（其他情况无法使用时，尝试加上 -e 参数）
5. 咪咕源貌似部分宽带无法使用
# 感谢
[NodeJs版本](https://github.com/nondanee/UnblockNeteaseMusic)以及为它贡献的所有coder
