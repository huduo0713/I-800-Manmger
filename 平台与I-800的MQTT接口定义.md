

# 下发算法-request

**topic**:

`/sys/i800/{deviceId}/request`

**payload**:

```
{
  "cmdId": "uuid-1234",
  "version":"1.0",
  "method": "algorithm.add",
  "timestamp": "2023-10-27 10:06:00", // 指令下发时间
  "params": {
  	"algorithmId": "uuid-1234",	// 算法id
  	"algorithmName": "夜间节能策略算法",       // 算法名称
    "algorithmVersion": "1.0",              // 算法版本
    "algorithmVersionId": "uuid-1234"	// 算法版本id
    "algorithmDataUrl": "http://113.249.91.53:9001/haikang/algorithmZip/5fe37a4d248b413d8e62057bc6adb11c",     // http下载链接， 算法包下载到/usr/runtime/algorithm/{algorithmId}/{algorithmVersionId}下
    "fileSize":5242880,   // 文件总大小（字节），用于设备预估进度和空间
    "lastModifyTime": "2025-08-28 16:09:36",
    "md5": "a1b2c3d4e5f67890abcdef1234567890" // 文件校验和，使用md5shm
  }
}
```

# 下发算法-reply

**topic**:

`/sys/i800/{deviceId}/reply`

**payload**:

```
{
  "cmdId": "uuid-1234",
  "version":"1.0",
  "method": "algorithm.add",
  "timestamp": "2023-10-27 10:06:00", // 指令下发时间
  "code": 1,  // 0-成功；其他-失败
  "message": "checksum failed",
  "data": null
}

```



# 删除算法-request

**topic**:

``/sys/i800/{deviceId}/request``

**payload**:

```
{
  "cmdId": "uuid-1234",
  "version":"1.0",
  "method": "algorithm.delete",
  "timestamp": "2023-10-27 10:06:00", // 指令下发时间
  "params": {
  	"algorithmId": "uuid-1234"  // 算法id
  }
}
```

# 删除算法-reply

**topic**:

``/sys/i800/{deviceId}/reply``

**payload**:

```
{
  "cmdId": "uuid-1234",
  "version":"1.0",
  "method": "algorithm.delete",
  "timestamp": "2023-10-27 10:06:00", // 指令下发时间
  "message": "success",
  "code": 0,    // 0-成功；其他-失败 备注：当删除的算法不存在时，返回0， message警告提示算法不存在
  "data": null
}
```







# 查询算法-request

查询所有

**topic**:

``/sys/i800/{deviceId}/request``

**payload**:

```
{
  "cmdId": "uuid-1234",
  "version":"1.0",
  "method": "algorithm.show",
  "timestamp": "2023-10-27 10:06:00", // 指令下发时间
  "params": null

}
```

# 查询算法-reply

**topic**:

```/sys/i800/{deviceId}/reply```

**payload**:

```
{
  "cmdId": "uuid-1234",
  "version":"1.0",
  "method": "algorithm.show",
  "timestamp": "2023-10-27 10:06:00", // 指令下发时间
  "message": "success",
  "code": 0,    // 0-成功；其他-失败
  "data": [
  	{
  		"algorithmName": "夜间节能策略算法",       // 算法名称
    	"algorithmId": "uuid123",  // 算法标识
    	"algorithmVersion": "1.0"              // 算法版本
    	"runStatus": 1           // 算法运行状态， 1-运行；0-停止
  	}
  ]
}
```



# 启停算法-request

**topic**:

``/sys/i800/{deviceId}/request``

**payload**:

```
{
  "cmdId": "uuid-1234",
  "version":"1.0",
  "method": "algorithm.config",
  "timestamp": "2023-10-27 10:06:00", // 指令下发时间
   "params": {
  	"algorithmId": "uuid-1234",  // 算法标识
  	"runStatus": 0   // 设置为0标识关闭；设置为1标识开启  , 根据/usr/runtime/algorithm/{algorithmId}/{algorithmVersionId}/config.yaml中的algo.runStatus的值进行判断, 0表示关闭；1表示开启；
  }
}
```

# 启停算法-reply

**topic**:

```/sys/i800/{deviceId}/reply```

**payload**:

```
{
  "cmdId": "uuid-1234",
  "version":"1.0",
  "method": "algorithm.config",
  "timestamp": "2023-10-27 10:06:00", // 指令下发时间
  "message": "success",
  "code": 0,    // 0-成功；其他-失败
  "data": null
}
```



# 设备注册-register

对于频繁、周期性的通信，直接在topic层判别

**topic**:

``/sys/i800/{deviceId}/event/register``

**payload**:

```
{
  "cmdId": "uuid-1234",
  "version":"1.0",
  "method": "event.register",   // 这个字段可以不判定，和topic是一致的
  "timestamp": "2023-10-27 10:06:00", // 事件上报时间
   "data": {
    "deviceModule": "I-800-RK",
  	"deviceId": "F2-D5-4F-C7-2B-AD",  // 设备ID
  	"heartBeat": 10,   // 心跳周期 单位秒
  	"IP": "192.168.11.204",
    "runtimeStatus": 1,   //查询本地的1231端口是否在LISTEN, 如果在监听，则返回1，表示runtime进程正常，否则返回0
    "opcuaServerPort": 4840
  }
}
```

