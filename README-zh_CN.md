## t-rss
t-rss是一个自动rss bt种子文件的程序，相比flexget丰富的功能，t-rss功能较为精简，同时体积更小、占用资源更少，支持自动将rss得到的种子文件添加至bt客户端（目前支持qBittorrent和Deluge（未完整测试过）），将来会加入从irc获取资源信息的功能（irc已经不想写了。）

## 安装
首先[下载](https://github.com/capric98/t-rss/releases)最新的pre-release or release中系统/架构对应的压缩包

解压后得到一个二进制文件，如果没有x属性自己加一下

写好配置文件直接运行就完了，命令行参数见`-help`，此处略

## 配置
带*的部分是可省略不配置的，但是`download_to`和`client`至少配置一个不然程序跑完rss啥也不会干
```yaml
Name0:                             # 任务名称，随意取
  rss:                             # RSS链接
  cookie:                          #*如果rss链接需要cookie才能访问，请将其粘贴在这里
  edit_tracker:                    #*编辑tracker
    delete:                        #*
      - share                      #*正则表达式，可以有多条，匹配的tracker会被删除
    add:                           #*
      - https://example.com/       #*
      - https://example2.com/      #*
      - https://example3.com/      #*
  content_size:                    #*体积过滤，默认单位MB，同时支持以下写法：
    min: 2048                      #*min: "2GB"  or  min: "2G"
    max: 9999                      #*max: "1TB"  blabla...
  quota:                           #*单次rss限额
    num: 65535                     #*采用数量限制，默认65535
    size: maxint64                 #*采用体积限制，默认0x7FFFFFFFFFFFFFFF，格式与前min/max相同
  regexp:                          #*正则表达式，可不配置
    accept:                        #*若配置该项目，则只有同时符合两者的种子被采用
      - Vol.*?Fin
    reject:                        #*拒绝符合以下正则表达式列表的种子
      - Test
  interval: 10                     #*RSS间隔，单位秒，非负整数，默认60s
  delay: 0                         #*RSS延迟添加时间，单位秒，非负整数，默认0s
  download_to: "/home/WatchDir/"   #*种子文件保存目录，可不设置
  client:                          #*自动添加至以下客户端，可不设置
    This_is_a_client:              # 一个任意的标签，不能是纯数字，不可重复
      type: qBittorrent            # 需要指定客户端类型(qBittorrent/Deluge)
      host: http://127.0.0.1:8080  # qBittorrent的webui地址，http/https不可少，可以是远程地址
      username: admin              # webui用户名
      password: adminadmin         # webui密码
      dlLimit: 10M                 #*下载速度限制，单位可以是M/MB等等。。
      upLimit: 10M                 #*上传速度限制，单位可以是M/MB等等。。
      paused: true                 #*是否以暂停状态添加种子
      savepath: "/home/Downloads/" #*下载目录
    This_is_another_client:        #*可以有多个客户端
      type: Deluge                 #*deluge初步支持，未经过详细测试
      host: 127.0.0.1:????         # Deluge的rpc地址及其端口
      username: xxxx               #*具体配置项参见 https://bit.ly/2tzS2Gi
      password: xxxx               #*或者 https://bit.ly/36yJFcI
      ...
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
在RSS目录下运行二进制文件即可，默认使用同目录下的config.yml作为配置文件，历史保留在同目录下的`.RSS-history`目录内；也可以nohup或者注册成服务什么的。。

[go-yaml](https://github.com/go-yaml/yaml)
[go-rencode](https://github.com/gdm85/go-rencode)

