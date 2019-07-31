## GoRSS
GoRSS是一个自动rss bt种子文件的程序，相比flexget丰富的功能，GoRSS功能较为精简，同时体积更小、占用资源更少，支持自动将rss得到的种子文件添加至bt客户端（目前仅支持qBittorrent），将来会加入从irc获取资源信息的功能（咕咕咕）

## 安装
由于新手上路不太清楚如何较好把Go项目发到GitHub，因此建议不熟悉Go语言的朋友直接下载最新的release zip文件解压运行。解压后得到目录结构如下：
```bash
|- RSS
   |- config.yml # 配置文件
   |- GoRSS      # 软件本体，二进制格式可直接运行
   |- .RSS-saved # 缓存目录，用来记录种子是否下载过，在linux下默认隐藏
```

由于目前程序还在开发过程中，因此配置文件是直接hardcode进去的，将来会有自定义配置文件位置的特性。。
## 配置
带*的部分是可省略不配置的，但是`download_to`和`client`至少配置一个不然程序跑完rss啥也不会干。。
```yaml
Name0:                             # 任务名称，随意取
  rss:                             # RSS链接
  strict: no                       #*严格模式，某些rss给出文件大小均为0，在strict: yes下拒绝(?)
  content_size:                    #*体积过滤，默认单位MB，同时支持以下写法：
    min: 2048                      #*min: "2GB"  or  min: "2G"
    max: 9999                      #*max: "1TB"  blabla...
  regexp:                          #*正则表达式，可不配置
    accept:                  #*接受符合以下正则表达式列表的种子，仅在strict: yes时起作用
      - Vol.*?Fin
    reject:                  #*拒绝符合以下正则表达式列表的种子
      - Test
  interval: 10                     #*RSS间隔，单位秒，非负整数，默认60s
  download_to: "/home/WatchDir/"   #*种子文件保存目录，可不设置
  client:                          #*自动添加至以下客户端，可不设置
    qBittorrent:
      host: http://127.0.0.1:8080  # webui地址，http/https不可少，可以是远程地址
      username: admin              # webui用户名
      password: adminadmin         # webui密码
      dlLimit: "10M"               #*下载速度限制，注意双引号，单位可以是M/MB等等。。
      upLimit: "10M"               #*上传速度限制，注意双引号，单位可以是M/MB等等。。
      paused: "true"               #*是否以暂停状态添加种子，注意双引号不可少
      savepath: "/home/Downloads/" #*下载目录
    qBittorrent:                   #*可以有多个客户端
      ...
    #因为一些比较呆的原因，暂时不支持直接往deluge添加，请使用deluge的监视目录功能添加种子文件
Name1:                             #*一个配置文件里可以有多个任务
  rss: http://example.com/rss.xml
  content_size:
    max: 1000
    min: 100
  download_to: "/home/tmp"
Name2:
  ......
```
### 运行
在RSS目录下运行：`./GoRSS`即可，也可以nohup或者注册成服务什么的...只不过目前可能软件还不太稳定或者需要有一些功能改进。
## 依赖
[gofeed](https://github.com/mmcdole/gofeed)

[go-yaml](https://github.com/go-yaml/yaml)

[color](https://github.com/fatih/color)

