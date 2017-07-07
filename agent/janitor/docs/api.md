### sessions
`GET` `/proxy/sessions`

```json
{
  "stress-default-zgz-datamanmesos": {    // 应用
    "192.168.1.3": {                      //  来源IP
      "id": "2-stress-default-zgz-datamanmesos",  // 后端server
      "ip": "192.168.1.3",
      "port": 31002,
      "scheme": "",
      "version": "1496706111228860282",
      "weihgt": 100,
      "updated_at": "2017-06-18T14:13:47.709747733+08:00"    // 最后访问时间
    }
  },
  "xxx-default-zgz-datamanmesos": {
    "127.0.0.1": {
      "id": "1-xxx-default-zgz-datamanmesos",
      "ip": "192.168.1.3",
      "port": 31005,
      "scheme": "",
      "version": "1496706111228860282",
      "weihgt": 100,
      "updated_at": "2017-06-18T14:18:50.272170991+08:00"
    },
    "192.168.1.3": {
      "id": "1-xxx-default-zgz-datamanmesos",
      "ip": "192.168.1.3",
      "port": 31005,
      "scheme": "",
      "version": "1496706111228860282",
      "weihgt": 100,
      "updated_at": "2017-06-18T14:19:30.433291286+08:00"
    }
  }
}
```

### configs
`GET`  `/proxy/configs`

```json
{
  "listenAddr": ":80",
  "tlsListenAddr": "",
  "tlsCertFile": "",
  "tlsKeyFile": "",
  "domain": "bbklab.net",
  "advertiseIP": "192.168.1.137"
}
```

### upstreams

#### list
`GET` `/proxy/upstreams`

```json
[
  {
    "name": "stress-default-zgz-datamanmesos",     // upstream名（应用）
    "alias": "g.cn",                               // 对外的访问URL，HTTP代理 (可选)
    "listen": ":81",                               // 监听端口，4层代理 (可选)  
    "backends": [                                  // 后端server列表
      {
        "id": "1-stress-default-zgz-datamanmesos",
        "ip": "192.168.1.3",
        "port": 31001,
        "scheme": "",
        "version": "1496706111228860282",
        "weihgt": 100
      },
      {
        "id": "2-stress-default-zgz-datamanmesos",
        "ip": "192.168.1.3",
        "port": 31002,
        "scheme": "",
        "version": "1496706111228860282",
        "weihgt": 100
      }
    ]
  },
  {
    "name": "xxx-default-zgz-datamanmesos",
    "alias": "m.cn",
    "listen": ":82",
    "backends": [
      {
        "id": "1-xxx-default-zgz-datamanmesos",
        "ip": "192.168.1.3",
        "port": 31005,
        "scheme": "",
        "version": "1496706111228860282",
        "weihgt": 100
      }
    ]
  }
]
```

#### add / update
> 增加或修改一个upstream 和 backend，已存在则修改，不存在则添加。  
`PUT` `/proxy/upstreams`

```json
{
  "upstream": {   
    "name": "stress-default-zgz-datamanmesos",    // upstream名（应用） (添加后不可修改)
    "alias": "g.cn",                              // 对外的访问URL，HTTP代理 (可选)
    "listen": ":81"                               // 监听端口，4层代理 (可选)   (添加后不可修改)
  },
  "backend": {                                    // 一个指定的后端server
    "id": "1-stress-default-zgz-datamanmesos",    // 后端server ID (添加后不可修改)
    "ip": "192.168.1.3",
    "port": 31001,
    "version": "1496706111228860282",
    "weihgt": 100
  }
}
```

#### remove
`DELETE` `/proxy/upstreams`

```json
{
  "upstream": {
    "name": "stress-default-zgz-datamanmesos"
  },
  "backend": {
    "id": "1-stress-default-zgz-datamanmesos"
  }
}
```

### statistics
`GET` `/proxy/stats`

```json
{
  "counter": {      // 计数器
    "global": {     // 全局
      "rx_bytes": 1855,   // 累计入流量
      "tx_bytes": 1723,   // 累计出流量
      "requests": 16,     // 累计请求数量
      "fails": 0,         // 累计失败数量
      "rx_rate": 0,       // 实时入流量速率
      "tx_rate": 0,       // 实时出流量速率
      "requests_rate": 0,   // 实时请求速率
      "fails_rate": 0,      // 实时失败速率
      "uptime": "16m28.736555285s"  // 运行时长
    },
    "upstream": {             // upstream 统计数据
      "stress-default-zgz-datamanmesos": {   // upstream 数据 
        "1-stress-default-zgz-datamanmesos": {  // backend 数据 
          "active_clients": 1,                  // 当前在线连接数
          "rx_bytes": 0,
          "tx_bytes": 0,
          "requests": 2,
          "rx_rate": 0,
          "tx_rate": 0,
          "requests_rate": 0,
          "uptime": "3m34.211797489s"
        },
        "2-stress-default-zgz-datamanmesos": {
          "active_clients": 0,
          "rx_bytes": 243,
          "tx_bytes": 545,
          "requests": 4,
          "rx_rate": 0,
          "tx_rate": 0,
          "requests_rate": 0,
          "uptime": "7m52.235305312s"
        }
      },
      "xxx-default-zgz-datamanmesos": {
        "1-xxx-default-zgz-datamanmesos": {
          "active_clients": 2,
          "rx_bytes": 1612,
          "tx_bytes": 1178,
          "requests": 13,
          "rx_rate": 0,
          "tx_rate": 0,
          "requests_rate": 0,
          "uptime": "7m29.000419101s"
        }
      }
    }
  },
  "httpd": ":80",
  "httpdTLS": "",
  "tcpd": {            // 4层TCP代理  
    ":81": {           // 监听端口
      "active_clients": 1,  // 当前在线连接数
      "listen": ":81",
      "serving": true,
      "uptime": "15m59.771092122s"
    },
    ":82": {
      "active_clients": 2,
      "listen": ":82",
      "serving": true,
      "uptime": "11m57.742399309s"
    }
  }
}
```

获取upstream（应用）的统计数据  
`GET` `/proxy/stats/stress-default-zgz-datamanmesos`  

获取backend server（Task）的统计数据  
`GET` `/proxy/stats/stress-default-zgz-datamanmesos/2-stress-default-zgz-datamanmesos`  
