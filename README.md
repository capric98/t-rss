## t-rss
t-rss是一个自动rss bt种子文件的程序，相比flexget丰富的功能，t-rss功能较为精简，同时体积更小、占用资源更少，支持自动将rss得到的种子文件添加至bt客户端（目前支持qBittorrent和Deluge（未完整测试过）），将来会加入从irc获取资源信息的功能（irc已经不想写了。）

## 安装
首先[下载](https://github.com/capric98/t-rss/releases)最新的pre-release or release中系统/架构对应的压缩包

解压后得到一个二进制文件，如果没有x属性自己加一下

写好配置文件直接运行就完了，命令行参数见`-help`，此处略

## 配置
带*的部分是可省略不配置的，但`receiver`部分需要至少配置一个不然程序跑完rss啥也不会干

<details>
<summary>config.yml(格式修改中)</summary>

```yaml
GLOBAL:
  log_file: # empty -> stderr
  history:
    max_age: 30s # {int}s/m/h/d
    save_to: # ~/home/.t-rss/history/
  timeout: 1m

TASKS:
  Name_of_task0:
    rss:
      url: http(s)://example.com
      method: GET #*GET/POST
      headers:    #*if needed
        Cookie: something
        Key: Value
      interval: 10s # {int}s/m/h/d
    filter:
      content_size:
        min:
        max:
      regexp:
        accept:
          - A
        reject:
          - B
    quota:
      num: 65535
      size:
    edit:
      tracker:
        delete:
          - share
        add:
          - http(s)://example.com/
    receiver:
      save_to: /home/WatchDir/
      client:
        Name_of_client0:
          type: qBittorrent
          url: http(s)://example.com
          username: admin
          password: adminadmin
          dlLimit:
          upLimit:
          pause: true
          savepath: /home/Downloads
        Name_of_client1:
          type: Deluge
          host: 127.0.0.1:1234
          username:
          password:

  Name_of_task1:
    rss:
      url: http(s)://example.com
      receiver:
        save_path: /home/WatchDir/
  Name_of_task2:
    rss:
      url: http(s)://example.com
      receiver:
        save_path: /home/WatchDir/

```

</details>

### 运行
在RSS目录下运行二进制文件即可，默认使用同目录下的config.yml作为配置文件，历史保留在同目录下的`.t-rss_History`目录内；也可以nohup或者注册成服务什么的。。

### TODO
  * 重写client部分
  * 增加test覆盖率

[go-yaml](https://github.com/go-yaml/yaml)

[go-rencode](https://github.com/gdm85/go-rencode)

[logrus](https://github.com/sirupsen/logrus)

[go-colorable](https://github.com/mattn/go-colorable)
