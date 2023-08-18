# 分布式对象存储系统--测试记录

## 1 单机对象存储

单机对象存储，仅支持PUT和GET操作，接口服务与数据存储服务高度耦合，都在同一台主机，甚至是同一个函数中处理。



启动单机服务

```sh
LISTEN_ADDRESS=:12345 STORAGE_ROOT=/tmp go run server.go

#创建存储目录
mkdir /tmp/objects
```

另起一个终端作为客户端，测试PUT和GET功能

```sh
#尝试GET对象（404，因为还没有上传）
curl -v localhost:12345/objects/test

*   Trying 127.0.0.1:12345...
* Connected to localhost (127.0.0.1) port 12345 (#0)
> GET /objects/test HTTP/1.1
> Host: localhost:12345
> User-Agent: curl/7.87.0
> Accept: */*
> 
* Mark bundle as not supporting multiuse
< HTTP/1.1 404 Not Found
< Date: Fri, 18 Aug 2023 14:23:57 GMT
< Content-Length: 0
< 
* Connection #0 to host localhost left intact
```

```sh
#上传对象（200，上传成功）
curl -v localhost:12345/objects/test -XPUT -d"this is a test object"

*   Trying 127.0.0.1:12345...
* Connected to localhost (127.0.0.1) port 12345 (#0)
> PUT /objects/test HTTP/1.1
> Host: localhost:12345
> User-Agent: curl/7.87.0
> Accept: */*
> Content-Length: 21
> Content-Type: application/x-www-form-urlencoded
> 
* Mark bundle as not supporting multiuse
< HTTP/1.1 200 OK
< Date: Fri, 18 Aug 2023 14:27:11 GMT
< Content-Length: 0
< 
* Connection #0 to host localhost left intact
```

```sh
#再次GET该对象（200，获取成功）
 curl -v localhost:12345/objects/test
 
*   Trying 127.0.0.1:12345...
* Connected to localhost (127.0.0.1) port 12345 (#0)
> GET /objects/test HTTP/1.1
> Host: localhost:12345
> User-Agent: curl/7.87.0
> Accept: */*
> 
* Mark bundle as not supporting multiuse
< HTTP/1.1 200 OK
< Date: Fri, 18 Aug 2023 14:28:24 GMT
< Content-Length: 21
< Content-Type: text/plain; charset=utf-8
< 
* Connection #0 to host localhost left intact
this is a test object
```



## 2 可扩展的分布式系统

### 2.1 安装RabbitMQ

请在Linux环境下安装，尽管Windows环境下也有相应的适配版本，但是总是不能正常开启服务（打不开[管理页面](http://localhost:15672)）。

```shell
sudo apt-get install rabbitmq-server
sudo rabbitmq-plugins enable rabbitmq_management
wget localhost:15672/cli/rabbitmqadmin

#创建exchange（也可以在管理页面手动创建）
python rabbitmqadmin declare exchange name=apiServers type=fanout
python rabbitmqadmin declare exchange name=dataServers type=fanout

#添加用户及权限
sudo rabbitmqctl add_user test test
sudo rabbitmqctl set_permissions -p / test ".*" ".*" ".*"
```



### 2.2 执行启动脚本

已将分布式各节点的启动命令放到`tools`目录下，

```sh
cd tools

#初始化环境（网络配置、目录配置）
/bin/bash ./init_test_env.sh

#执行启动脚本（启动服务）
/bin/bash ./start_test_env.sh
```



### 2.3 测试服务

```sh
#尝试上传一个对象
curl -v 10.29.2.2:12345/objects/test2_222 -XPUT -d"this object test2_222"

*   Trying 10.29.2.2:12345...
* Connected to 10.29.2.2 (10.29.2.2) port 12345 (#0)
> PUT /objects/test2_222 HTTP/1.1
> Host: 10.29.2.2:12345
> User-Agent: curl/7.87.0
> Accept: */*
> Content-Length: 21
> Content-Type: application/x-www-form-urlencoded
> 
2023/08/18 22:08:30 choose server: 10.29.1.6:12345
* Mark bundle as not supporting multiuse
< HTTP/1.1 200 OK
< Date: Fri, 18 Aug 2023 14:08:30 GMT
< Content-Length: 0
< 
* Connection #0 to host 10.29.2.2 left intact

#定位这个对象
curl 10.29.2.2:12345/locate/test2_222
"10.29.1.6:12345"

#GET这个对象
curl 10.29.2.2:12345/objects/test2_222
this object test2_222

#尝试换一个api服务节点GET（分布式节点均服务正常）
curl 10.29.2.1:12345/objects/test2_222
this object test2_222
```



## 3 元数据服务

### 3.1 安装ElasticSearch

建议安装`ElasticSearch 7.10`，到官网选择直接下载安装，安装过程参考https://zhuanlan.zhihu.com/p/336560713

```shell
#如下则ES安装成功
curl -X GET "localhost:9200"

{
  "name" : "LAPTOP-QRDIJITP",
  "cluster_name" : "elasticsearch",
  "cluster_uuid" : "o1a0MgTWQBmS-MzSG_nOEA",
  "version" : {
    "number" : "7.10.1",
    "build_flavor" : "default",
    "build_type" : "deb",
    "build_hash" : "1c34507e66d7db1211f66f3513706fdf548736aa",
    "build_date" : "2020-12-05T01:00:33.671820Z",
    "build_snapshot" : false,
    "lucene_version" : "8.7.0",
    "minimum_wire_compatibility_version" : "6.8.0",
    "minimum_index_compatibility_version" : "6.0.0-beta1"
  },
  "tagline" : "You Know, for Search"
}
```



ES的启动命令已内置到`start_test_env.sh`中，接下来创建metadata索引以及objects类型的映射。

【注意】：ElasticSearch在7.x版本之后，不再支持`string`type，index也不支持`not_analyzed`，故原书中的命令无法执行，相关命令更改如下：

```sh
curl -H"Content-Type: application/json" localhost:9200/metadata -XPUT -d'{"mappings":{"properties":{"name":{"type":"text","index":true,"fielddata":true},"version":{"type":"integer"},"size":{"type":"integer"},"hash":{"type":"text"}}}}'
```



```sh
curl -H"Content-Type: application/json" localhost:9200/metadata -XPUT -d'{"mappings":{"properties":{"name":{"type":"text","index":true,"fielddata":true},"version":{"type":"integer"},"size":{"type":"integer"},"hash":{"type":"text"}}}}'
```

### 3.2 元数据服务

#### 3.2.1 上传对象

接下来用`curl`命令作为客户端访问服务节点，PUT一个test3_333对象

```sh
#上传对象，但是没有指定hash值(上传失败)
curl -v 10.29.2.2:12345/objects/test3_333 -XPUT -d"this is object test3_333"

*   Trying 10.29.2.2:12345...
* Connected to 10.29.2.2 (10.29.2.2) port 12345 (#0)
> PUT /objects/test3_333 HTTP/1.1
> Host: 10.29.2.2:12345
> User-Agent: curl/7.87.0
> Accept: */*
> Content-Length: 24
> Content-Type: application/x-www-form-urlencoded
> 
2023/08/18 19:55:04 missing object hash in digest header...
* Mark bundle as not supporting multiuse
< HTTP/1.1 400 Bad Request	#400报错 因为未提供对象的hash值
< Date: Fri, 18 Aug 2023 11:55:04 GMT
< Content-Length: 0
< 
* Connection #0 to host 10.29.2.2 left intact

```

```sh
#上传对象且指定了hash值（上传成功）
 echo -n "this is object test3_333"|openssl dgst -sha256 -binary |base64
oqV4BrFLU0oUK5LHq38S0lgC1D2u7L16BXpZM3LHh4k=

curl -v 10.29.2.2:12345/objects/test3_333 -XPUT -d"this is object test3_333" -H "Digest: SHA-256=oqV4BrFLU0oUK5LHq38S0lgC1D2u7L16BXpZM3LHh4k="

*   Trying 10.29.2.2:12345...
* Connected to 10.29.2.2 (10.29.2.2) port 12345 (#0)
> PUT /objects/test3_333 HTTP/1.1
> Host: 10.29.2.2:12345
> User-Agent: curl/7.87.0
> Accept: */*
> Digest: SHA-256=oqV4BrFLU0oUK5LHq38S0lgC1D2u7L16BXpZM3LHh4k=
> Content-Length: 24
> Content-Type: application/x-www-form-urlencoded
> 
2023/08/18 20:11:00 choose server: 10.29.1.1:12345
2023/08/18 20:11:00 addVersion error: <nil>
* Mark bundle as not supporting multiuse
< HTTP/1.1 200 OK
< Date: Fri, 18 Aug 2023 12:11:00 GMT
< Content-Length: 0
< 
* Connection #0 to host 10.29.2.2 left intact
```



```sh
#再次上传test3_333对象，但内容不同
echo -n "this is object test3_333 version2"|openssl dgst -sha256 -binary |base64
Fv7VZb+YDG8EbA6qprIU4G1K4pt8SpIP6P4VcYIa+xQ=

curl -v 10.29.2.2:12345/objects/test3_333 -XPUT -d"this is object test3_333 version2" -H "Digest: SHA-256=Fv7VZb+YDG8EbA6qprIU4G1K4pt8SpIP6P4VcYIa+xQ="

*   Trying 10.29.2.2:12345...
* Connected to 10.29.2.2 (10.29.2.2) port 12345 (#0)
> PUT /objects/test3_333 HTTP/1.1
> Host: 10.29.2.2:12345
> User-Agent: curl/7.87.0
> Accept: */*
> Digest: SHA-256=Fv7VZb+YDG8EbA6qprIU4G1K4pt8SpIP6P4VcYIa+xQ=
> Content-Length: 33
> Content-Type: application/x-www-form-urlencoded
> 
2023/08/18 20:33:07 choose server: 10.29.1.4:12345
2023/08/18 20:33:07 addVersion error: <nil>
* Mark bundle as not supporting multiuse
< HTTP/1.1 200 OK
< Date: Fri, 18 Aug 2023 12:33:07 GMT
< Content-Length: 0
< 
* Connection #0 to host 10.29.2.2 left intact
```

目前已上传两个版本的`test3_333`对象，hash值分别为：

`oqV4BrFLU0oUK5LHq38S0lgC1D2u7L16BXpZM3LHh4k=`，

`Fv7VZb+YDG8EbA6qprIU4G1K4pt8SpIP6P4VcYIa+xQ=`。

#### 3.2.2 定位对象

```sh
 #定位test3_333对象所在的存储节点
 curl 10.29.2.1:12345/locate/oqV4BrFLU0oUK5LHq38S0lgC1D2u7L16BXpZM3LHh4k=
"10.29.1.1:12345"

curl 10.29.2.1:12345/locate/Fv7VZb+YDG8EbA6qprIU4G1K4pt8SpIP6P4VcYIa+xQ=
"10.29.1.4:12345"

#查看test3_333对象版本
curl 10.29.2.1:12345/versions/test3_333
{"Name":"test3_333","Version":1,"Size":24,"Hash":"oqV4BrFLU0oUK5LHq38S0lgC1D2u7L16BXpZM3LHh4k="}
{"Name":"test3_333","Version":2,"Size":33,"Hash":"Fv7VZb+YDG8EbA6qprIU4G1K4pt8SpIP6P4VcYIa+xQ="}
```



#### 3.2.3 GET指定版本的对象

在header中指定要获取对象的version，

```sh
curl 10.29.2.1:12345/objects/test3_333?version=1
this is object test3_333

 curl 10.29.2.1:12345/objects/test3_333
this is object test3_333 version2	#默认获取最新版本的对象
```



#### 3.2.4 删除对象

删除对象，再尝试GET就无法直接获取（因为默认最新version，对应ES端记录的hash为空）

但是如果自己指定version可以查到对应元数据（含hash），对应对象就找得到。

```sh
#删除对象
curl -v 10.29.2.1:12345/objects/test3_333 -XDELETE

*   Trying 10.29.2.1:12345...
* Connected to 10.29.2.1 (10.29.2.1) port 12345 (#0)
> DELETE /objects/test3_333 HTTP/1.1
> Host: 10.29.2.1:12345
> User-Agent: curl/7.87.0
> Accept: */*
> 
* Mark bundle as not supporting multiuse
< HTTP/1.1 200 OK
< Date: Fri, 18 Aug 2023 13:05:04 GMT
< Content-Length: 0
< 
* Connection #0 to host 10.29.2.1 left intact

#删除对象后再尝试GET对象（404）
curl -v 10.29.2.1:12345/objects/test3_333

*   Trying 10.29.2.1:12345...
* Connected to 10.29.2.1 (10.29.2.1) port 12345 (#0)
> GET /objects/test3_333 HTTP/1.1
> Host: 10.29.2.1:12345
> User-Agent: curl/7.87.0
> Accept: */*
> 
* Mark bundle as not supporting multiuse
< HTTP/1.1 404 Not Found
< Date: Fri, 18 Aug 2023 13:05:48 GMT
< Content-Length: 0
< 
* Connection #0 to host 10.29.2.1 left intact

#查看版本（删除后，hash置为""）
curl 10.29.2.1:12345/versions/test3_333
{"Name":"test3_333","Version":1,"Size":24,"Hash":"oqV4BrFLU0oUK5LHq38S0lgC1D2u7L16BXpZM3LHh4k="}
{"Name":"test3_333","Version":2,"Size":33,"Hash":"Fv7VZb+YDG8EbA6qprIU4G1K4pt8SpIP6P4VcYIa+xQ="}
{"Name":"test3_333","Version":3,"Size":0,"Hash":""}

#指定version，仍可以GET对象
curl 10.29.2.1:12345/objects/test3_333?version=1
this is object test3_333

curl 10.29.2.1:12345/objects/test3_333?version=2
this is object test3_333 version2
```





```shell
```



##  4 数据校验和去重

### 4.1 去重

#### 4.1.1 连续PUT多个 名字不同 而 内容相同 的对象

```shell
#内容都如下
echo -n "this object will have only 1 instance"|openssl dgst -sha256 -binary|base64
aWKQ2BipX94sb+h3xdTbWYAu1yzjn5vyFG2SOwUQIXY=

#不同名字
curl -v 10.29.2.1:12345/objects/test4_111 -XPUT -d"this object will have only 1 instance" -H "Digest: SHA-256=aWKQ2BipX94sb+h3xdTbWYAu1yzjn5vyFG2SOwUQIXY="

*   Trying 10.29.2.1:12345...
* Connected to 10.29.2.1 (10.29.2.1) port 12345 (#0)
> PUT /objects/test4_111 HTTP/1.1
> Host: 10.29.2.1:12345
> User-Agent: curl/7.87.0
> Accept: */*
> Digest: SHA-256=aWKQ2BipX94sb+h3xdTbWYAu1yzjn5vyFG2SOwUQIXY=
> Content-Length: 37
> Content-Type: application/x-www-form-urlencoded
> 
2023/08/18 18:34:39 choose server: 10.29.1.3:12345
* Mark bundle as not supporting multiuse
< HTTP/1.1 200 OK
< Date: Fri, 18 Aug 2023 10:34:39 GMT
< Content-Length: 0
< 
* Connection #0 to host 10.29.2.1 left intact


 curl -v 10.29.2.1:12345/objects/test4_222 -XPUT -d"this object will have only 1 instance" -H "Digest: SHA-256=aWKQ2BipX94sb+h3xdTbWYAu1yzjn5vyFG2SOwUQIXY="
 
*   Trying 10.29.2.1:12345...
* Connected to 10.29.2.1 (10.29.2.1) port 12345 (#0)
> PUT /objects/test4_222 HTTP/1.1
> Host: 10.29.2.1:12345
> User-Agent: curl/7.87.0
> Accept: */*
> Digest: SHA-256=aWKQ2BipX94sb+h3xdTbWYAu1yzjn5vyFG2SOwUQIXY=
> Content-Length: 37
> Content-Type: application/x-www-form-urlencoded
> 
* Mark bundle as not supporting multiuse
< HTTP/1.1 200 OK
< Date: Fri, 18 Aug 2023 10:35:24 GMT
< Content-Length: 0
< 
* Connection #0 to host 10.29.2.1 left intact
```



#### 4.1.2 查看上传结果

因为内容相同，所以仅保存一个对象

```shell
#定位对象查看上传结果
curl 10.29.2.1:12345/locate/aWKQ2BipX94sb+h3xdTbWYAu1yzjn5vyFG2SOwUQIXY=
"10.29.1.3:12345"

#查看磁盘
 ls /tmp/?/objects/aWKQ2BipX94sb+h3xdTbWYAu1yzjn5vyFG2SOwUQIXY=
'/tmp/3/objects/aWKQ2BipX94sb+h3xdTbWYAu1yzjn5vyFG2SOwUQIXY='
```



#### 4.1.3 尝试GET对象

二者名字不同，但内容相同，指向同一对象

```shell
 curl 10.29.2.2:12345/objects/test4_111
this object will have only 1 instance

curl 10.29.2.2:12345/objects/test4_222
this object will have only 1 instance
```



### 4.2 数据校验

尝试PUT一个hash值不正确的对象

```shell
curl -v 10.29.2.2:12345/objects/test4_111 -XPUT -d"this object will have only 1 instance" -H "Digest: SHA-256=incorrecthash"


*   Trying 10.29.2.2:12345...
* Connected to 10.29.2.2 (10.29.2.2) port 12345 (#0)
> PUT /objects/test4_111 HTTP/1.1
> Host: 10.29.2.2:12345
> User-Agent: curl/7.87.0
> Accept: */*
> Digest: SHA-256=incorrecthash
> Content-Length: 37
> Content-Type: application/x-www-form-urlencoded
> 
2023/08/18 19:20:45 choose server: 10.29.1.1:12345
2023/08/18 19:20:45 object hash mismatch,calculated=aWKQ2BipX94sb+h3xdTbWYAu1yzjn5vyFG2SOwUQIXY=,but requested=incorrecthash

* Mark bundle as not supporting multiuse
< HTTP/1.1 400 Bad Request
< Date: Fri, 18 Aug 2023 11:20:45 GMT
< Content-Length: 0
< 
* Connection #0 to host 10.29.2.2 left intact


```







## 5 数据冗余和即时修复

实现了数据分片和RS纠错技术

PUT一个对象，会分成多个分片保存在多个数据节点中。

```shell
# 获取sha256摘要
echo -n "this object will be separate to 4+2 shards"|openssl dgst -sha256 -binary|base64

>  return:
MBMxWHrPMsuOBaVYHkwScZQRyTRMQyiKp2oelpLZza8=

#访问服务节点，PUT一个test5对象
curl -v 10.29.2.1:12345/objects/test5 -XPUT -d"this object will be separate to 4+2 shards" -H "Digest: SHA-256=MBMxWHrPMsuOBaVYHkwScZQRyTRMQyiKp2oelpLZza8="

> return:
separate to 4+2 shards" -H "Digest: SHA-256=MBMxWHrPMsuOBaVYHkwScZQRyTRMQyiKp2oelpLZza8="
*   Trying 10.29.2.1:12345...
* Connected to 10.29.2.1 (10.29.2.1) port 12345 (#0)
> PUT /objects/test5 HTTP/1.1
> Host: 10.29.2.1:12345
> User-Agent: curl/7.87.0
> Accept: */*
> Digest: SHA-256=MBMxWHrPMsuOBaVYHkwScZQRyTRMQyiKp2oelpLZza8=
> Content-Length: 42
> Content-Type: application/x-www-form-urlencoded
> 
2023/08/10 23:48:51 choose servers: [10.29.1.3:12345 10.29.1.4:12345 10.29.1.5:12345 10.29.1.6:12345 10.29.1.2:12345 10.29.1.1:12345]
* Mark bundle as not supporting multiuse
< HTTP/1.1 200 OK
< Date: Thu, 10 Aug 2023 15:48:51 GMT
< Content-Length: 0
< 
* Connection #0 to host 10.29.2.1 left intact
```



```shell
#检查分片在各节点磁盘存储情况
ls /tmp/?/objects

> return:
/tmp/1/objects:
'MBMxWHrPMsuOBaVYHkwScZQRyTRMQyiKp2oelpLZza8=.5.wGW6r6pLkHAJC2GlYxfk45FdUTTv31c57INXIUjmhZ8='

/tmp/2/objects:
'MBMxWHrPMsuOBaVYHkwScZQRyTRMQyiKp2oelpLZza8=.4.i8xiyIwSO2cRJwnmkO4ieUV9v26H6e8tu5Y%2F3Op%2F4zE='

/tmp/3/objects:
'MBMxWHrPMsuOBaVYHkwScZQRyTRMQyiKp2oelpLZza8=.0.XVFHp5%2F5kZ89051XQo6UEkWW8OGzyXwLWS4Ln9f0Ncg='

/tmp/4/objects:
'MBMxWHrPMsuOBaVYHkwScZQRyTRMQyiKp2oelpLZza8=.1.DjgCAigrm%2FBMDzVlPdjPp+LZMHY9ktSKNX9A9eQShAQ='

/tmp/5/objects:
'MBMxWHrPMsuOBaVYHkwScZQRyTRMQyiKp2oelpLZza8=.2.pV2SP%2Fi3jK9KGs5BtQS++TJEecq8Z7%2FYaUnSRPU1IX8='

/tmp/6/objects:
'MBMxWHrPMsuOBaVYHkwScZQRyTRMQyiKp2oelpLZza8=.3.9cMmcwZQE+dlbz27iekkG2%2FL4raiYzUUSvcbfE9xUKw='
```



```shell
#查看各个分片
cat /tmp/3/objects/MBMxWHrPMsuOBaVYHkwScZQRyTRMQyiKp2oelpLZza8=.0.XVFHp5%2F5kZ89051XQo6UEkWW8OGzyXwLWS4Ln9f0Ncg=
> return:
this object

cat /tmp/4/objects/MBMxWHrPMsuOBaVYHkwScZQRyTRMQyiKp2oelpLZza8=.1.DjgCAigrm%2FBMDzVlPdjPp+LZMHY9ktSKNX9A9eQShAQ=
> return:
 will be se
 
 cat /tmp/5/objects/MBMxWHrPMsuOBaVYHkwScZQRyTRMQyiKp2oelpLZza8=.2.pV2SP%2Fi3jK9KGs5BtQS++TJEecq8Z7%2FYaUnSRPU1IX8=
> return:
parate to 4
 
  cat /tmp/6/objects/MBMxWHrPMsuOBaVYHkwScZQRyTRMQyiKp2oelpLZza8=.3.9cMmcwZQE+dlbz27iekkG2%2FL4raiYzUUSvcbfE9xUKw=
> return:
 +2 shards
 
 #因此这些数据分片可以组成
 this object will be separate to 4 +2 shards
```



```shell
#获取对象
curl 10.29.2.2:12345/objects/test5
>return:
this object will be separate to 4+2 shards	#可正常获取到对象内容

#定位对象
curl 10.29.2.1:12345/locate/MBMxWHrPMsuOBaVYHkwScZQRyTRMQyiKp2oelpLZza8=
> return:
{"0":"10.29.1.3:12345","1":"10.29.1.4:12345","2":"10.29.1.5:12345","3":"10.29.1.6:12345","4":"10.29.1.2:12345","5":"10.29.1.1:12345"}
```



```shell
#删除分片0，修改分片1内容，查看是否能恢复

#删除分片0
rm /tmp/3/objects/MBMxWHrPMsuOBaVYHkwScZQRyTRMQyiKp2oelpLZza8=.0.XVFHp5%2F5kZ89051XQo6UEkWW8OGzyXwLWS4Ln9f0Ncg=

#修改分片1
echo some_garbage > /tmp/4/objects/MBMxWHrPMsuOBaVYHkwScZQRyTRMQyiKp2oelpLZza8=.1.DjgCAigrm%2FBMDzVlPdjPp+LZMHY9ktSKNX9A9eQShAQ=

#查看分片1
cat /tmp/4/objects/MBMxWHrPMsuOBaVYHkwScZQRyTRMQyiKp2oelpLZza8=.1.DjgCAigrm%2FBMDzVlPdjPp+LZMHY9ktSKNX9A9eQShAQ=

> return:
some_garbage

#获取对象
curl 10.29.2.2:12345/objects/test5

> log:
2023/08/11 00:47:41 filenamePaths not found,name MBMxWHrPMsuOBaVYHkwScZQRyTRMQyiKp2oelpLZza8=.0, found []
2023/08/11 00:47:41 object hash mismatch, remove /tmp/4/objects/MBMxWHrPMsuOBaVYHkwScZQRyTRMQyiKp2oelpLZza8=.1.DjgCAigrm%2FBMDzVlPdjPp+LZMHY9ktSKNX9A9eQShAQ=

> return:
this object will be separate to 4+2 shards

#再次查看分片1
cat /tmp/4/objects/MBMxWHrPMsuOBaVYHkwScZQRyTRMQyiKp2oelpLZza8=.1.DjgCAigrm%2FBMDzVlPdjPp+LZMHY9ktSKNX9A9eQShAQ=
> return:
 will be se
 
 #再次检查数据节点上的内容
 ls /tmp/?/objects
 
 > return:
 /tmp/1/objects:
'MBMxWHrPMsuOBaVYHkwScZQRyTRMQyiKp2oelpLZza8=.5.wGW6r6pLkHAJC2GlYxfk45FdUTTv31c57INXIUjmhZ8='

/tmp/2/objects:
'MBMxWHrPMsuOBaVYHkwScZQRyTRMQyiKp2oelpLZza8=.4.i8xiyIwSO2cRJwnmkO4ieUV9v26H6e8tu5Y%2F3Op%2F4zE='

/tmp/3/objects:
'MBMxWHrPMsuOBaVYHkwScZQRyTRMQyiKp2oelpLZza8=.0.XVFHp5%2F5kZ89051XQo6UEkWW8OGzyXwLWS4Ln9f0Ncg='

/tmp/4/objects:
'MBMxWHrPMsuOBaVYHkwScZQRyTRMQyiKp2oelpLZza8=.1.DjgCAigrm%2FBMDzVlPdjPp+LZMHY9ktSKNX9A9eQShAQ='

/tmp/5/objects:
'MBMxWHrPMsuOBaVYHkwScZQRyTRMQyiKp2oelpLZza8=.2.pV2SP%2Fi3jK9KGs5BtQS++TJEecq8Z7%2FYaUnSRPU1IX8='

/tmp/6/objects:
'MBMxWHrPMsuOBaVYHkwScZQRyTRMQyiKp2oelpLZza8=.3.9cMmcwZQE+dlbz27iekkG2%2FL4raiYzUUSvcbfE9xUKw='

#数据已恢复正常

```

## 6 断点续传

 

```shell
#生成一段约100kb的随机文件
dd if=/dev/urandom of=/tmp/file bs=1000 count=100

> retuen:
记录了100+0 的读入
记录了100+0 的写出
100000 bytes (100 kB, 98 KiB) copied, 0.000597542 s, 167 MB/s

#计算随机文件的hash值
openssl dgst -sha256 -binary /tmp/file | base64

> return:
soudSwVx4hcqK/G1Gkqhs1U/EJ87GfsIj+L3voEvjwo=
```



```shell
#将随机文件分段上传为test666对象
curl -v 10.29.2.2:12345/objects/test666 -XPOST -H "Digest: SHA-256=soudSwVx4hcqK/G1Gkqhs1U/EJ87GfsIj+L3voEvjwo=" -H "Size: 100000"
*   Trying 10.29.2.2:12345...
* Connected to 10.29.2.2 (10.29.2.2) port 12345 (#0)
> POST /objects/test666 HTTP/1.1
> Host: 10.29.2.2:12345
> User-Agent: curl/7.87.0
> Accept: */*
> Digest: SHA-256=soudSwVx4hcqK/G1Gkqhs1U/EJ87GfsIj+L3voEvjwo=
> Size: 100000
> 
[2023-08-16 13:50:41 10.29.2.2:12345:go-distributed-oss/apiServer/objects.post(post.go):31]
hash:soudSwVx4hcqK/G1Gkqhs1U/EJ87GfsIj+L3voEvjwo=
* Mark bundle as not supporting multiuse
< HTTP/1.1 201 Created
< Location: /temp/eyJOYW1lIjoidGVzdDY2NiIsIlNpemUiOjEwMDAwMCwiSGFzaCI6InNvdWRTd1Z4NGhjcUslMkZHMUdrcWhzMVUlMkZFSjg3R2ZzSWorTDN2b0V2andvPSIsIlNlcnZlcnMiOlsiMTAuMjkuMS4zOjEyMzQ1IiwiMTAuMjkuMS42OjEyMzQ1IiwiMTAuMjkuMS4yOjEyMzQ1IiwiMTAuMjkuMS4xOjEyMzQ1IiwiMTAuMjkuMS40OjEyMzQ1IiwiMTAuMjkuMS41OjEyMzQ1Il0sIlV1aWRzIjpbIjUwNWM0YTFmLWZmOWItNDM3Zi05ZDJmLTdlMThlMjRiYmZjNiIsImExNDk2OGIyLTczNWEtNGU4Ni04NjM1LTAzMzY5NmI0N2ZhNyIsImFlNzljMDA1LTcxNTMtNDE2ZS05NTdjLWQ4MjUwNDMxYjFkNSIsIjY0OGU4ZTY3LWEwMGItNGRmZi1iZjM4LTEzY2YwZWQ1ZWY0YyIsIjJjZTYxMjZjLTRiMzUtNDAxNS1hYTkyLWQ3ZmQ1ZWJlN2MzNSIsImFhZDFlYjBmLTQ3MTgtNDIxNi1iZmM4LTM3Yzc0MjFhMTFhNSJdfQ==
< Date: Wed, 16 Aug 2023 05:50:42 GMT
< Content-Length: 0
< 
* Connection #0 to host 10.29.2.2 left intact

```



```sh
#注意：每次生成的token都带有dataServers和uuids信息
#uuid都是随机生成的，所以每次token都不一样（虽然有可能长得很像，但其实不一样）

#查看上传前的大小
curl -I 10.29.2.1:12345/temp/eyJOYW1lIjoidGVzdDY2NiIsIlNpemUiOjEwMDAwMCwiSGFzaCI6InNvdWRTd1Z4NGhjcUslMkZHMUdrcWhzMVUlMkZFSjg3R2ZzSWorTDN2b0V2andvPSIsIlNlcnZlcnMiOlsiMTAuMjkuMS4zOjEyMzQ1IiwiMTAuMjkuMS42OjEyMzQ1IiwiMTAuMjkuMS4yOjEyMzQ1IiwiMTAuMjkuMS4xOjEyMzQ1IiwiMTAuMjkuMS40OjEyMzQ1IiwiMTAuMjkuMS41OjEyMzQ1Il0sIlV1aWRzIjpbIjUwNWM0YTFmLWZmOWItNDM3Zi05ZDJmLTdlMThlMjRiYmZjNiIsImExNDk2OGIyLTczNWEtNGU4Ni04NjM1LTAzMzY5NmI0N2ZhNyIsImFlNzljMDA1LTcxNTMtNDE2ZS05NTdjLWQ4MjUwNDMxYjFkNSIsIjY0OGU4ZTY3LWEwMGItNGRmZi1iZjM4LTEzY2YwZWQ1ZWY0YyIsIjJjZTYxMjZjLTRiMzUtNDAxNS1hYTkyLWQ3ZmQ1ZWJlN2MzNSIsImFhZDFlYjBmLTQ3MTgtNDIxNi1iZmM4LTM3Yzc0MjFhMTFhNSJdfQ==
HTTP/1.1 200 OK
Content-Length: 0
Date: Wed, 16 Aug 2023 05:53:02 GMT

#生成第一分片
dd if=/tmp/file of=/tmp/first bs=1000 count=50
记录了50+0 的读入
记录了50+0 的写出
50000 bytes (50 kB, 49 KiB) copied, 0.000365006 s, 137 MB/s

#上传第一分片
url -v -XPUT --data-binary @/tmp/first 10.29.2.1:12345/temp/eyJOYW1lIjoidGVzdDY2NiIsIlNpemUiOjEwMDAwMCwiSGFzaCI6InNvdWRTd1Z4NGhjcUslMkZHMUdrcWhzMVUlMkZFSjg3R2ZzSWorTDN2b0V2andvPSIsIlNlcnZlcnMiOlsiMTAuMjkuMS4zOjEyMzQ1IiwiMTAuMjkuMS42OjEyMzQ1IiwiMTAuMjkuMS4yOjEyMzQ1IiwiMTAuMjkuMS4xOjEyMzQ1IiwiMTAuMjkuMS40OjEyMzQ1IiwiMTAuMjkuMS41OjEyMzQ1Il0sIlV1aWRzIjpbIjUwNWM0YTFmLWZmOWItNDM3Zi05ZDJmLTdlMThlMjRiYmZjNiIsImExNDk2OGIyLTczNWEtNGU4Ni04NjM1LTAzMzY5NmI0N2ZhNyIsImFlNzljMDA1LTcxNTMtNDE2ZS05NTdjLWQ4MjUwNDMxYjFkNSIsIjY0OGU4ZTY3LWEwMGItNGRmZi1iZjM4LTEzY2YwZWQ1ZWY0YyIsIjJjZTYxMjZjLTRiMzUtNDAxNS1hYTkyLWQ3ZmQ1ZWJlN2MzNSIsImFhZDFlYjBmLTQ3MTgtNDIxNi1iZmM4LTM3Yzc0MjFhMTFhNSJdfQ==
*   Trying 10.29.2.1:12345...
* Connected to 10.29.2.1 (10.29.2.1) port 12345 (#0)
> PUT /temp/eyJOYW1lIjoidGVzdDY2NiIsIlNpemUiOjEwMDAwMCwiSGFzaCI6InNvdWRTd1Z4NGhjcUslMkZHMUdrcWhzMVUlMkZFSjg3R2ZzSWorTDN2b0V2andvPSIsIlNlcnZlcnMiOlsiMTAuMjkuMS4zOjEyMzQ1IiwiMTAuMjkuMS42OjEyMzQ1IiwiMTAuMjkuMS4yOjEyMzQ1IiwiMTAuMjkuMS4xOjEyMzQ1IiwiMTAuMjkuMS40OjEyMzQ1IiwiMTAuMjkuMS41OjEyMzQ1Il0sIlV1aWRzIjpbIjUwNWM0YTFmLWZmOWItNDM3Zi05ZDJmLTdlMThlMjRiYmZjNiIsImExNDk2OGIyLTczNWEtNGU4Ni04NjM1LTAzMzY5NmI0N2ZhNyIsImFlNzljMDA1LTcxNTMtNDE2ZS05NTdjLWQ4MjUwNDMxYjFkNSIsIjY0OGU4ZTY3LWEwMGItNGRmZi1iZjM4LTEzY2YwZWQ1ZWY0YyIsIjJjZTYxMjZjLTRiMzUtNDAxNS1hYTkyLWQ3ZmQ1ZWJlN2MzNSIsImFhZDFlYjBmLTQ3MTgtNDIxNi1iZmM4LTM3Yzc0MjFhMTFhNSJdfQ== HTTP/1.1
> Host: 10.29.2.1:12345
> User-Agent: curl/7.87.0
> Accept: */*
> Content-Length: 50000
> Content-Type: application/x-www-form-urlencoded
> 
* Mark bundle as not supporting multiuse
< HTTP/1.1 200 OK
< Date: Wed, 16 Aug 2023 05:53:53 GMT
< Content-Length: 0
< 
* Connection #0 to host 10.29.2.1 left intact


#再次查看已上传的大小
 curl -I 10.29.2.1:12345/temp/eyJOYW1lIjoidGVzdDY2NiIsIlNpemUiOjEwMDAwMCwiSGFzaCI6InNvdWRTd1Z4NGhjcUslMkZHMUdrcWhzMVUlMkZFSjg3R2ZzSWorTDN2b0V2andvPSIsIlNlcnZlcnMiOlsiMTAuMjkuMS4zOjEyMzQ1IiwiMTAuMjkuMS42OjEyMzQ1IiwiMTAuMjkuMS4yOjEyMzQ1IiwiMTAuMjkuMS4xOjEyMzQ1IiwiMTAuMjkuMS40OjEyMzQ1IiwiMTAuMjkuMS41OjEyMzQ1Il0sIlV1aWRzIjpbIjUwNWM0YTFmLWZmOWItNDM3Zi05ZDJmLTdlMThlMjRiYmZjNiIsImExNDk2OGIyLTczNWEtNGU4Ni04NjM1LTAzMzY5NmI0N2ZhNyIsImFlNzljMDA1LTcxNTMtNDE2ZS05NTdjLWQ4MjUwNDMxYjFkNSIsIjY0OGU4ZTY3LWEwMGItNGRmZi1iZjM4LTEzY2YwZWQ1ZWY0YyIsIjJjZTYxMjZjLTRiMzUtNDAxNS1hYTkyLWQ3ZmQ1ZWJlN2MzNSIsImFhZDFlYjBmLTQ3MTgtNDIxNi1iZmM4LTM3Yzc0MjFhMTFhNSJdfQ==
HTTP/1.1 200 OK
Content-Length: 32000
Date: Wed, 16 Aug 2023 05:54:17 GMT

#生成第二分片（后68000字节）
 dd if=/tmp/file of=/tmp/second bs=1000 skip=32 count=68
记录了68+0 的读入
记录了68+0 的写出
68000 bytes (68 kB, 66 KiB) copied, 0.000391465 s, 174 MB/s

#上传第二分片
curl -v -XPUT --data-binary @/tmp/second -H "range: bytes=32000-" 10.29.2.1:12345/temp/eyJOYW1lIjoidGVzdDY2NiIsIlNpemUiOjEwMDAwMCwiSGFzaCI6InNvdWRTd1Z4NGhjcUslMkZHMUdrcWhzMVUlMkZFSjg3R2ZzSWorTDN2b0V2andvPSIsIlNlcnZlcnMiOlsiMTAuMjkuMS4zOjEyMzQ1IiwiMTAuMjkuMS42OjEyMzQ1IiwiMTAuMjkuMS4yOjEyMzQ1IiwiMTAuMjkuMS4xOjEyMzQ1IiwiMTAuMjkuMS40OjEyMzQ1IiwiMTAuMjkuMS41OjEyMzQ1Il0sIlV1aWRzIjpbIjUwNWM0YTFmLWZmOWItNDM3Zi05ZDJmLTdlMThlMjRiYmZjNiIsImExNDk2OGIyLTczNWEtNGU4Ni04NjM1LTAzMzY5NmI0N2ZhNyIsImFlNzljMDA1LTcxNTMtNDE2ZS05NTdjLWQ4MjUwNDMxYjFkNSIsIjY0OGU4ZTY3LWEwMGItNGRmZi1iZjM4LTEzY2YwZWQ1ZWY0YyIsIjJjZTYxMjZjLTRiMzUtNDAxNS1hYTkyLWQ3ZmQ1ZWJlN2MzNSIsImFhZDFlYjBmLTQ3MTgtNDIxNi1iZmM4LTM3Yzc0MjFhMTFhNSJdfQ==
*   Trying 10.29.2.1:12345...
* Connected to 10.29.2.1 (10.29.2.1) port 12345 (#0)
> PUT /temp/eyJOYW1lIjoidGVzdDY2NiIsIlNpemUiOjEwMDAwMCwiSGFzaCI6InNvdWRTd1Z4NGhjcUslMkZHMUdrcWhzMVUlMkZFSjg3R2ZzSWorTDN2b0V2andvPSIsIlNlcnZlcnMiOlsiMTAuMjkuMS4zOjEyMzQ1IiwiMTAuMjkuMS42OjEyMzQ1IiwiMTAuMjkuMS4yOjEyMzQ1IiwiMTAuMjkuMS4xOjEyMzQ1IiwiMTAuMjkuMS40OjEyMzQ1IiwiMTAuMjkuMS41OjEyMzQ1Il0sIlV1aWRzIjpbIjUwNWM0YTFmLWZmOWItNDM3Zi05ZDJmLTdlMThlMjRiYmZjNiIsImExNDk2OGIyLTczNWEtNGU4Ni04NjM1LTAzMzY5NmI0N2ZhNyIsImFlNzljMDA1LTcxNTMtNDE2ZS05NTdjLWQ4MjUwNDMxYjFkNSIsIjY0OGU4ZTY3LWEwMGItNGRmZi1iZjM4LTEzY2YwZWQ1ZWY0YyIsIjJjZTYxMjZjLTRiMzUtNDAxNS1hYTkyLWQ3ZmQ1ZWJlN2MzNSIsImFhZDFlYjBmLTQ3MTgtNDIxNi1iZmM4LTM3Yzc0MjFhMTFhNSJdfQ== HTTP/1.1
> Host: 10.29.2.1:12345
> User-Agent: curl/7.87.0
> Accept: */*
> range: bytes=32000-
> Content-Length: 68000
> Content-Type: application/x-www-form-urlencoded
> 
* We are completely uploaded and fine
* Mark bundle as not supporting multiuse
< HTTP/1.1 200 OK
< Date: Wed, 16 Aug 2023 05:55:56 GMT
< Content-Length: 0
< 
* Connection #0 to host 10.29.2.1 left intact

#get所上传的对象
#有的时候会因为hash错误而获取不到，"/"本来对应着"%2F"，但有时候会变成"%252F"，导致404
#不知道哪里出问题，后面自己又好了，很奇怪
curl 10.29.2.1:12345/objects/test666 > /tmp/output
  % Total    % Received % Xferd  Average Speed   Time    Time     Time  Current
                                 Dload  Upload   Total   Spent    Left  Speed
  0     0    0     0    0     0      0      0 --:--:-- --:--:-- --:--:--     0
100   97k    0   97k    0     0  4301k      0 --:--:-- --:--:-- --:--:-- 4438k


#比较原文件和断点续传下载的文件
diff -s /tmp/output /tmp/file
檔案 /tmp/output 和 /tmp/file 相同

#使用range下载指定范围的数据
curl 10.29.2.1:12345/objects/test666 -H "range: bytes=32000-" > /tmp/output2
  % Total    % Received % Xferd  Average Speed   Time    Time     Time  Current
                                 Dload  Upload   Total   Spent    Left  Speed
  0     0    0     0    0     0      0      0 --:--:-- --:--:-- --:--:--     0
100 68000    0 68000    0     0  3068k      0 --:--:-- --:--:-- --:--:-- 3162k

#比较原文件和下载文件
 diff -s /tmp/output2 /tmp/second
檔案 /tmp/output2 和 /tmp/second 相同

```


## 7 数据压缩

```shell
#生成一个100MB的测试文件，内容全为0
dd if=/dev/zero of=/tmp/file7 bs=1M count=100
记录了100+0 的读入
记录了100+0 的写出
104857600 bytes (105 MB, 100 MiB) copied, 0.0897956 s, 1.2 GB/s

#计算hash值
openssl dgst -sha256 -binary /tmp/file7|base64
IEkqTQ2E+L6xdn9mFiKfhdRMKCe2S9v7Jg7hL6EQng4=

```
```shell
#上传为test7对象
curl -v 10.29.2.1:12345/objects/test7 -XPUT --data-binary @/tmp/file7 -H "Digest: SHA-256=IEkqTQ2E+L6xdn9mFiKfhdRMKCe2S9v7Jg7hL6EQng4="


*   Trying 10.29.2.1:12345...
* Connected to 10.29.2.1 (10.29.2.1) port 12345 (#0)
> PUT /objects/test7 HTTP/1.1
> Host: 10.29.2.1:12345
> User-Agent: curl/7.87.0
> Accept: */*
> Digest: SHA-256=IEkqTQ2E+L6xdn9mFiKfhdRMKCe2S9v7Jg7hL6EQng4=
> Content-Length: 104857600
> Content-Type: application/x-www-form-urlencoded
> Expect: 100-continue
> 
* Done waiting for 100-continue
[2023-08-17 07:14:14 10.29.2.1:12345:go-distributed-oss/apiServer/objects.putStream(put.go):75]
choose servers:[10.29.1.4:12345 10.29.1.6:12345 10.29.1.5:12345 10.29.1.2:12345 10.29.1.3:12345 10.29.1.1:12345]
* Mark bundle as not supporting multiuse
< HTTP/1.1 100 Continue
* We are completely uploaded and fine
* Mark bundle as not supporting multiuse
< HTTP/1.1 200 OK
< Date: Wed, 16 Aug 2023 23:14:23 GMT
< Content-Length: 0
< 
* Connection #0 to host 10.29.2.1 left intact

#用ls命令查看分片对象大小
# 25514字节（25MB）无压缩的情况下的分片大小
ls -ltr /tmp/?/objects/IE*

-rw-rw-r-- 1 tsingfa tsingfa 25514 8月  17 07:14 '/tmp/4/objects/IEkqTQ2E+L6xdn9mFiKfhdRMKCe2S9v7Jg7hL6EQng4=.0.OUw0XwsMY+5lJiemLu0GkkTTXE1RNOTwfU6rtRr9pH4='
-rw-rw-r-- 1 tsingfa tsingfa 25514 8月  17 07:14 '/tmp/6/objects/IEkqTQ2E+L6xdn9mFiKfhdRMKCe2S9v7Jg7hL6EQng4=.1.OUw0XwsMY+5lJiemLu0GkkTTXE1RNOTwfU6rtRr9pH4='
-rw-rw-r-- 1 tsingfa tsingfa 25514 8月  17 07:14 '/tmp/5/objects/IEkqTQ2E+L6xdn9mFiKfhdRMKCe2S9v7Jg7hL6EQng4=.2.OUw0XwsMY+5lJiemLu0GkkTTXE1RNOTwfU6rtRr9pH4='
-rw-rw-r-- 1 tsingfa tsingfa 25514 8月  17 07:14 '/tmp/2/objects/IEkqTQ2E+L6xdn9mFiKfhdRMKCe2S9v7Jg7hL6EQng4=.3.OUw0XwsMY+5lJiemLu0GkkTTXE1RNOTwfU6rtRr9pH4='
-rw-rw-r-- 1 tsingfa tsingfa 25514 8月  17 07:14 '/tmp/3/objects/IEkqTQ2E+L6xdn9mFiKfhdRMKCe2S9v7Jg7hL6EQng4=.4.OUw0XwsMY+5lJiemLu0GkkTTXE1RNOTwfU6rtRr9pH4='
-rw-rw-r-- 1 tsingfa tsingfa 25514 8月  17 07:14 '/tmp/1/objects/IEkqTQ2E+L6xdn9mFiKfhdRMKCe2S9v7Jg7hL6EQng4=.5.OUw0XwsMY+5lJiemLu0GkkTTXE1RNOTwfU6rtRr9pH4='

```

```shell
#下载test7对象并对比数据
curl -v 10.29.2.1:12345/objects/test7 -o /tmp/output


  % Total    % Received % Xferd  Average Speed   Time    Time     Time  Current
                                 Dload  Upload   Total   Spent    Left  Speed
  0     0    0     0    0     0      0      0 --:--:-- --:--:-- --:--:--     0*   Trying 10.29.2.1:12345...
* Connected to 10.29.2.1 (10.29.2.1) port 12345 (#0)
> GET /objects/test7 HTTP/1.1
> Host: 10.29.2.1:12345
> User-Agent: curl/7.87.0
> Accept: */*
> 
* Mark bundle as not supporting multiuse
< HTTP/1.1 200 OK
< Date: Wed, 16 Aug 2023 23:22:00 GMT
< Content-Type: application/octet-stream
< Transfer-Encoding: chunked
< 
{ [32013 bytes data]
100  100M    0  100M    0     0   121M      0 --:--:-- --:--:-- --:--:--  121M
* Connection #0 to host 10.29.2.1 left intact


#下载test7对象输出（100MB）
ls -ltr  /tmp/output

-rw-rw-r-- 1 tsingfa tsingfa 104857600 8月  17 07:22 /tmp/output

#比较输出文件与原文件
diff -s /tmp/output /tmp/file7

檔案 /tmp/output 和 /tmp/file7 相同

```

```shell
#以gzip压缩的方式下载数据（大小仅98k）
curl -v 10.29.2.1:12345/objects/test7 -H "Accept-Encoding: gzip" -o /tmp/output2.gz

  % Total    % Received % Xferd  Average Speed   Time    Time     Time  Current
                                 Dload  Upload   Total   Spent    Left  Speed
  0     0    0     0    0     0      0      0 --:--:-- --:--:-- --:--:--     0*   Trying 10.29.2.1:12345...
* Connected to 10.29.2.1 (10.29.2.1) port 12345 (#0)
> GET /objects/test7 HTTP/1.1
> Host: 10.29.2.1:12345
> User-Agent: curl/7.87.0
> Accept: */*
> Accept-Encoding: gzip
> 
  0     0    0     0    0     0      0      0 --:--:-- --:--:-- --:--:--     0* Mark bundle as not supporting multiuse
< HTTP/1.1 200 OK
< Content-Encoding: gzip
< Date: Thu, 17 Aug 2023 00:13:57 GMT
< Transfer-Encoding: chunked
< 
{ [3988 bytes data]
100   99k    0   99k    0     0    98k      0 --:--:--  0:00:01 --:--:--   98k
* Connection #0 to host 10.29.2.1 left intact


#解压并对比数据
gunzip /tmp/output2.gz

gzip: /tmp/output2 already exists; do you wish to overwrite (y or n)? y

----
diff -s /tmp/output2 /tmp/file7

檔案 /tmp/output2 和 /tmp/file7 相同


```



## 8 数据维护

### 8.1 上传对象，等待处理

```sh
#先给test8对象上传6个版本
echo -n "this is object test8 version 1"|openssl dgst -sha256 -binary|base64
2IJQkIth94IVsnPQMrsNxz1oqfrsPo0E2ZmZfJLDZnE=

curl 10.29.2.1:12345/objects/test8 -XPUT -d"this is object test8 version 1" -H "Digest: SHA-256=2IJQkIth94IVsnPQMrsNxz1oqfrsPo0E2ZmZfJLDZnE="
------
echo -n "this is object test8 version 2-6"|openssl dgst -sha256 -binary|base64
66WuRH0s0albWDZ9nTmjFo9JIqTTXmB6EiRkhTh1zeA=

curl 10.29.2.1:12345/objects/test8 -XPUT -d"this is object test8 version 2-6" -H "Digest: SHA-256=66WuRH0s0albWDZ9nTmjFo9JIqTTXmB6EiRkhTh1zeA="

curl 10.29.2.1:12345/objects/test8 -XPUT -d"this is object test8 version 2-6" -H "Digest: SHA-256=66WuRH0s0albWDZ9nTmjFo9JIqTTXmB6EiRkhTh1zeA="

curl 10.29.2.1:12345/objects/test8 -XPUT -d"this is object test8 version 2-6" -H "Digest: SHA-256=66WuRH0s0albWDZ9nTmjFo9JIqTTXmB6EiRkhTh1zeA="

curl 10.29.2.1:12345/objects/test8 -XPUT -d"this is object test8 version 2-6" -H "Digest: SHA-256=66WuRH0s0albWDZ9nTmjFo9JIqTTXmB6EiRkhTh1zeA="

curl 10.29.2.1:12345/objects/test8 -XPUT -d"this is object test8 version 2-6" -H "Digest: SHA-256=66WuRH0s0albWDZ9nTmjFo9JIqTTXmB6EiRkhTh1zeA="
------
#查看test8的当前版本
curl 10.29.2.1:12345/versions/test8
{"Name":"test8","Version":1,"Size":30,"Hash":"2IJQkIth94IVsnPQMrsNxz1oqfrsPo0E2ZmZfJLDZnE="}
{"Name":"test8","Version":2,"Size":32,"Hash":"66WuRH0s0albWDZ9nTmjFo9JIqTTXmB6EiRkhTh1zeA="}
{"Name":"test8","Version":3,"Size":32,"Hash":"66WuRH0s0albWDZ9nTmjFo9JIqTTXmB6EiRkhTh1zeA="}
{"Name":"test8","Version":4,"Size":32,"Hash":"66WuRH0s0albWDZ9nTmjFo9JIqTTXmB6EiRkhTh1zeA="}
{"Name":"test8","Version":5,"Size":32,"Hash":"66WuRH0s0albWDZ9nTmjFo9JIqTTXmB6EiRkhTh1zeA="}
{"Name":"test8","Version":6,"Size":32,"Hash":"66WuRH0s0albWDZ9nTmjFo9JIqTTXmB6EiRkhTh1zeA="}

#用ls查看磁盘上的对象文件
ls -l /tmp/?/objects
/tmp/1/objects:
总用量 64
-rw-rw-r-- 1 tsingfa tsingfa    32 8月  17 22:32 '2IJQkIth94IVsnPQMrsNxz1oqfrsPo0E2ZmZfJLDZnE=.1.2eKLvcHfGvzIi+X5HFzgiCJyqjXV9%2F2U08FC6Srcslg='
-rw-rw-r-- 1 tsingfa tsingfa    32 8月  17 22:33 '66WuRH0s0albWDZ9nTmjFo9JIqTTXmB6EiRkhTh1zeA=.5.ih70CdjuiOerAQJiRj5Nnha6at+Rz9A6GKemrqmIDD4='
-rw-rw-r-- 1 tsingfa tsingfa 25514 8月  17 07:14 'IEkqTQ2E+L6xdn9mFiKfhdRMKCe2S9v7Jg7hL6EQng4=.5.OUw0XwsMY+5lJiemLu0GkkTTXE1RNOTwfU6rtRr9pH4='
-rw-rw-r-- 1 tsingfa tsingfa 25000 8月  16 13:55 'soudSwVx4hcqK%2FG1Gkqhs1U%2FEJ87GfsIj+L3voEvjwo=.3.ZIF4Fxs1HfmFMNWTq2vPOgPokMua%2FoAqpnamDeor23M='

/tmp/2/objects:
总用量 64
-rw-rw-r-- 1 tsingfa tsingfa    32 8月  17 22:32 '2IJQkIth94IVsnPQMrsNxz1oqfrsPo0E2ZmZfJLDZnE=.2.WGDnaf+GobSSS3wODa8r0IAgqC1ngM3KTGOklo12P%2Fw='
-rw-rw-r-- 1 tsingfa tsingfa    32 8月  17 22:33 '66WuRH0s0albWDZ9nTmjFo9JIqTTXmB6EiRkhTh1zeA=.0.xPZ9Cf8mShrJsL32FnbSVcayc9W5Y05clRo3GOkLyG0='
-rw-rw-r-- 1 tsingfa tsingfa 25514 8月  17 07:14 'IEkqTQ2E+L6xdn9mFiKfhdRMKCe2S9v7Jg7hL6EQng4=.3.OUw0XwsMY+5lJiemLu0GkkTTXE1RNOTwfU6rtRr9pH4='
-rw-rw-r-- 1 tsingfa tsingfa 25000 8月  16 13:55 'soudSwVx4hcqK%2FG1Gkqhs1U%2FEJ87GfsIj+L3voEvjwo=.2.M5NzdMLTwDMVf62PgH858k8877WH0AM4N1UaD%2FSMA7Q='

/tmp/3/objects:
总用量 64
-rw-rw-r-- 1 tsingfa tsingfa    32 8月  17 22:32 '2IJQkIth94IVsnPQMrsNxz1oqfrsPo0E2ZmZfJLDZnE=.5.ftcCG2hNzSmXh+RRIUPQXg58Kr8zEl9mzZZ6gmrSfH8='
-rw-rw-r-- 1 tsingfa tsingfa    32 8月  17 22:33 '66WuRH0s0albWDZ9nTmjFo9JIqTTXmB6EiRkhTh1zeA=.4.lAxEeRg2CNM7HWbwEUxkrorkqPO9pGI4syLdnOJ6lMI='
-rw-rw-r-- 1 tsingfa tsingfa 25514 8月  17 07:14 'IEkqTQ2E+L6xdn9mFiKfhdRMKCe2S9v7Jg7hL6EQng4=.4.OUw0XwsMY+5lJiemLu0GkkTTXE1RNOTwfU6rtRr9pH4='
-rw-rw-r-- 1 tsingfa tsingfa 25000 8月  16 13:55 'soudSwVx4hcqK%2FG1Gkqhs1U%2FEJ87GfsIj+L3voEvjwo=.0.6HkuURJ+sFWKN7xy+ryhUz2NL5ttEXxDvEtFjZ2jaWI='

/tmp/4/objects:
总用量 64
-rw-rw-r-- 1 tsingfa tsingfa    32 8月  17 22:32 '2IJQkIth94IVsnPQMrsNxz1oqfrsPo0E2ZmZfJLDZnE=.0.xPZ9Cf8mShrJsL32FnbSVcayc9W5Y05clRo3GOkLyG0='
-rw-rw-r-- 1 tsingfa tsingfa    32 8月  17 22:33 '66WuRH0s0albWDZ9nTmjFo9JIqTTXmB6EiRkhTh1zeA=.1.2eKLvcHfGvzIi+X5HFzgiCJyqjXV9%2F2U08FC6Srcslg='
-rw-rw-r-- 1 tsingfa tsingfa 25514 8月  17 07:14 'IEkqTQ2E+L6xdn9mFiKfhdRMKCe2S9v7Jg7hL6EQng4=.0.OUw0XwsMY+5lJiemLu0GkkTTXE1RNOTwfU6rtRr9pH4='
-rw-rw-r-- 1 tsingfa tsingfa 25000 8月  16 13:55 'soudSwVx4hcqK%2FG1Gkqhs1U%2FEJ87GfsIj+L3voEvjwo=.4.reIFPA%2FP1XLjMiJf3esL65cuXHRMbnAZBkv5y+Nlgzo='

/tmp/5/objects:
总用量 64
-rw-rw-r-- 1 tsingfa tsingfa    32 8月  17 22:32 '2IJQkIth94IVsnPQMrsNxz1oqfrsPo0E2ZmZfJLDZnE=.3.qBIQp3Kid8QEkuMld0xIc1494hqIkPLdKzEGl4MBchk='
-rw-rw-r-- 1 tsingfa tsingfa    32 8月  17 22:33 '66WuRH0s0albWDZ9nTmjFo9JIqTTXmB6EiRkhTh1zeA=.3.k7Z7BMDLAqtsm+AnQLO0dwSdXat1CnaUgRyE0f9ZgZ0='
-rw-rw-r-- 1 tsingfa tsingfa 25514 8月  17 07:14 'IEkqTQ2E+L6xdn9mFiKfhdRMKCe2S9v7Jg7hL6EQng4=.2.OUw0XwsMY+5lJiemLu0GkkTTXE1RNOTwfU6rtRr9pH4='
-rw-rw-r-- 1 tsingfa tsingfa 25000 8月  16 13:55 'soudSwVx4hcqK%2FG1Gkqhs1U%2FEJ87GfsIj+L3voEvjwo=.5.eBclQDufpQWd1sdapeomN0HG9E4haX74mnjrTG1Ns6I='

/tmp/6/objects:
总用量 64
-rw-rw-r-- 1 tsingfa tsingfa    32 8月  17 22:32 '2IJQkIth94IVsnPQMrsNxz1oqfrsPo0E2ZmZfJLDZnE=.4.Bg1hMBtSp3uHCiIoNfPdU+UcCtWIe3j8ZdSdM0DUMQ0='
-rw-rw-r-- 1 tsingfa tsingfa    32 8月  17 22:33 '66WuRH0s0albWDZ9nTmjFo9JIqTTXmB6EiRkhTh1zeA=.2.WGDnaf+GobSSS3wODa8r0IAgqC1ngM3KTGOklo12P%2Fw='
-rw-rw-r-- 1 tsingfa tsingfa 25514 8月  17 07:14 'IEkqTQ2E+L6xdn9mFiKfhdRMKCe2S9v7Jg7hL6EQng4=.1.OUw0XwsMY+5lJiemLu0GkkTTXE1RNOTwfU6rtRr9pH4='
-rw-rw-r-- 1 tsingfa tsingfa 25000 8月  16 13:55 'soudSwVx4hcqK%2FG1Gkqhs1U%2FEJ87GfsIj+L3voEvjwo=.1.rsSA6IGLPpvHl6pkE5a92Rk6uD4m+WkkzOboOcgsBVI='

```



### 8.2 测试deleteOldMetadata工具

```shell
#配置相关服务
export RABBITMQ_SERVER=amqp://test:test@localhost:5672
 export ES_SERVER=localhost:9200
 
 #运行deleteOldMetadata工具
 go run ../maintenance/deleteOldMetadata/deleteOldMetadata.go
 
 #再次查看test8版本
 curl 10.29.2.1:12345/versions/test8

{"Name":"test8","Version":2,"Size":32,"Hash":"66WuRH0s0albWDZ9nTmjFo9JIqTTXmB6EiRkhTh1zeA="}
{"Name":"test8","Version":3,"Size":32,"Hash":"66WuRH0s0albWDZ9nTmjFo9JIqTTXmB6EiRkhTh1zeA="}
{"Name":"test8","Version":4,"Size":32,"Hash":"66WuRH0s0albWDZ9nTmjFo9JIqTTXmB6EiRkhTh1zeA="}
{"Name":"test8","Version":5,"Size":32,"Hash":"66WuRH0s0albWDZ9nTmjFo9JIqTTXmB6EiRkhTh1zeA="}
{"Name":"test8","Version":6,"Size":32,"Hash":"66WuRH0s0albWDZ9nTmjFo9JIqTTXmB6EiRkhTh1zeA="}

#这里成功删除的关键是：es里的结构体能够成功被json解析
#注意将结构体的各个字段绑定好对应json
[2023-08-17 23:22:19 :go-distributed-oss/src/lib/es.SearchVersionStatus(es.go):202]
ar:es.aggregateResult{Aggregations:struct { GroupByName struct { Buckets []es.Bucket "json:\"buckets\"" } "json:\"group_by_name\"" }{GroupByName:struct { Buckets []es.Bucket "json:\"buckets\"" }{Buckets:[]es.Bucket{es.Bucket{Key:"test8", DocCount:6, MinVersion:struct { Value float32 "json:\"value\"" }{Value:1}}}}}}
```



### 8.3  测试deleteOrphanObject工具：

```sh
#test8-version1的元信息被删除，该对象现在为无引用的对象
#调用deleteOrphanObject工具清除
#使用delByAPI函数，只需调用一次
 STORAGE_ROOT=/tmp/? LISTEN_ADDRESS=10.29.2.1:12345 go run ../maintenance/deleteOrphanObject/deleteOrphanObject.go

> return:
[2023-08-18 01:26:13 10.29.2.1:12345:main.delByAPI(deleteOrphanObject.go):47]
delete: 2IJQkIth94IVsnPQMrsNxz1oqfrsPo0E2ZmZfJLDZnE=
2023/08/18 01:26:13 http: superfluous response.WriteHeader call from go-distributed-oss/apiServer/objects.Handler (handler.go:34)

[2023-08-18 01:26:13 10.29.2.1:12345:main.delByAPI(deleteOrphanObject.go):47]
delete: 2IJQkIth94IVsnPQMrsNxz1oqfrsPo0E2ZmZfJLDZnE=
2023/08/18 01:26:14 http: superfluous response.WriteHeader call from go-distributed-oss/apiServer/objects.Handler (handler.go:34)

[2023-08-18 01:26:14 10.29.2.1:12345:main.delByAPI(deleteOrphanObject.go):47]
delete: 2IJQkIth94IVsnPQMrsNxz1oqfrsPo0E2ZmZfJLDZnE=
2023/08/18 01:26:15 http: superfluous response.WriteHeader call from go-distributed-oss/apiServer/objects.Handler (handler.go:34)

[2023-08-18 01:26:15 10.29.2.1:12345:main.delByAPI(deleteOrphanObject.go):47]
delete: 2IJQkIth94IVsnPQMrsNxz1oqfrsPo0E2ZmZfJLDZnE=
2023/08/18 01:26:16 http: superfluous response.WriteHeader call from go-distributed-oss/apiServer/objects.Handler (handler.go:34)

[2023-08-18 01:26:16 10.29.2.1:12345:main.delByAPI(deleteOrphanObject.go):47]
delete: 2IJQkIth94IVsnPQMrsNxz1oqfrsPo0E2ZmZfJLDZnE=
2023/08/18 01:26:17 http: superfluous response.WriteHeader call from go-distributed-oss/apiServer/objects.Handler (handler.go:34)

[2023-08-18 01:26:17 10.29.2.1:12345:main.delByAPI(deleteOrphanObject.go):47]
delete: 2IJQkIth94IVsnPQMrsNxz1oqfrsPo0E2ZmZfJLDZnE=
2023/08/18 01:26:19 http: superfluous response.WriteHeader call from go-distributed-oss/apiServer/objects.Handler (handler.go:34)

#查看目录变化
ls -l /tmp/?/objects

> return:
/tmp/1/objects:
总用量 60
-rw-rw-r-- 1 tsingfa tsingfa    32 8月  17 22:33 '66WuRH0s0albWDZ9nTmjFo9JIqTTXmB6EiRkhTh1zeA=.5.ih70CdjuiOerAQJiRj5Nnha6at+Rz9A6GKemrqmIDD4='
-rw-rw-r-- 1 tsingfa tsingfa 25514 8月  17 07:14 'IEkqTQ2E+L6xdn9mFiKfhdRMKCe2S9v7Jg7hL6EQng4=.5.OUw0XwsMY+5lJiemLu0GkkTTXE1RNOTwfU6rtRr9pH4='
-rw-rw-r-- 1 tsingfa tsingfa 25000 8月  16 13:55 'soudSwVx4hcqK%2FG1Gkqhs1U%2FEJ87GfsIj+L3voEvjwo=.3.ZIF4Fxs1HfmFMNWTq2vPOgPokMua%2FoAqpnamDeor23M='

/tmp/2/objects:
总用量 60
-rw-rw-r-- 1 tsingfa tsingfa    32 8月  17 22:33 '66WuRH0s0albWDZ9nTmjFo9JIqTTXmB6EiRkhTh1zeA=.0.xPZ9Cf8mShrJsL32FnbSVcayc9W5Y05clRo3GOkLyG0='
-rw-rw-r-- 1 tsingfa tsingfa 25514 8月  17 07:14 'IEkqTQ2E+L6xdn9mFiKfhdRMKCe2S9v7Jg7hL6EQng4=.3.OUw0XwsMY+5lJiemLu0GkkTTXE1RNOTwfU6rtRr9pH4='
-rw-rw-r-- 1 tsingfa tsingfa 25000 8月  16 13:55 'soudSwVx4hcqK%2FG1Gkqhs1U%2FEJ87GfsIj+L3voEvjwo=.2.M5NzdMLTwDMVf62PgH858k8877WH0AM4N1UaD%2FSMA7Q='

/tmp/3/objects:
总用量 60
-rw-rw-r-- 1 tsingfa tsingfa    32 8月  17 22:33 '66WuRH0s0albWDZ9nTmjFo9JIqTTXmB6EiRkhTh1zeA=.4.lAxEeRg2CNM7HWbwEUxkrorkqPO9pGI4syLdnOJ6lMI='
-rw-rw-r-- 1 tsingfa tsingfa 25514 8月  17 07:14 'IEkqTQ2E+L6xdn9mFiKfhdRMKCe2S9v7Jg7hL6EQng4=.4.OUw0XwsMY+5lJiemLu0GkkTTXE1RNOTwfU6rtRr9pH4='
-rw-rw-r-- 1 tsingfa tsingfa 25000 8月  16 13:55 'soudSwVx4hcqK%2FG1Gkqhs1U%2FEJ87GfsIj+L3voEvjwo=.0.6HkuURJ+sFWKN7xy+ryhUz2NL5ttEXxDvEtFjZ2jaWI='

/tmp/4/objects:
总用量 60
-rw-rw-r-- 1 tsingfa tsingfa    32 8月  17 22:33 '66WuRH0s0albWDZ9nTmjFo9JIqTTXmB6EiRkhTh1zeA=.1.2eKLvcHfGvzIi+X5HFzgiCJyqjXV9%2F2U08FC6Srcslg='
-rw-rw-r-- 1 tsingfa tsingfa 25514 8月  17 07:14 'IEkqTQ2E+L6xdn9mFiKfhdRMKCe2S9v7Jg7hL6EQng4=.0.OUw0XwsMY+5lJiemLu0GkkTTXE1RNOTwfU6rtRr9pH4='
-rw-rw-r-- 1 tsingfa tsingfa 25000 8月  16 13:55 'soudSwVx4hcqK%2FG1Gkqhs1U%2FEJ87GfsIj+L3voEvjwo=.4.reIFPA%2FP1XLjMiJf3esL65cuXHRMbnAZBkv5y+Nlgzo='

/tmp/5/objects:
总用量 60
-rw-rw-r-- 1 tsingfa tsingfa    32 8月  17 22:33 '66WuRH0s0albWDZ9nTmjFo9JIqTTXmB6EiRkhTh1zeA=.3.k7Z7BMDLAqtsm+AnQLO0dwSdXat1CnaUgRyE0f9ZgZ0='
-rw-rw-r-- 1 tsingfa tsingfa 25514 8月  17 07:14 'IEkqTQ2E+L6xdn9mFiKfhdRMKCe2S9v7Jg7hL6EQng4=.2.OUw0XwsMY+5lJiemLu0GkkTTXE1RNOTwfU6rtRr9pH4='
-rw-rw-r-- 1 tsingfa tsingfa 25000 8月  16 13:55 'soudSwVx4hcqK%2FG1Gkqhs1U%2FEJ87GfsIj+L3voEvjwo=.5.eBclQDufpQWd1sdapeomN0HG9E4haX74mnjrTG1Ns6I='

/tmp/6/objects:
总用量 60
-rw-rw-r-- 1 tsingfa tsingfa    32 8月  17 22:33 '66WuRH0s0albWDZ9nTmjFo9JIqTTXmB6EiRkhTh1zeA=.2.WGDnaf+GobSSS3wODa8r0IAgqC1ngM3KTGOklo12P%2Fw='
-rw-rw-r-- 1 tsingfa tsingfa 25514 8月  17 07:14 'IEkqTQ2E+L6xdn9mFiKfhdRMKCe2S9v7Jg7hL6EQng4=.1.OUw0XwsMY+5lJiemLu0GkkTTXE1RNOTwfU6rtRr9pH4='
-rw-rw-r-- 1 tsingfa tsingfa 25000 8月  16 13:55 'soudSwVx4hcqK%2FG1Gkqhs1U%2FEJ87GfsIj+L3voEvjwo=.1.rsSA6IGLPpvHl6pkE5a92Rk6uD4m+WkkzOboOcgsBVI='

```



### 8.4 测试objectScanner工具

定期执行以检查并修复所有对象

```sh
# 模拟分片丢失
rm /tmp/1/objects/66WuRH0s0albWDZ9nTmjFo9JIqTTXmB6EiRkhTh1zeA=*

# 模拟分片损坏（数据变化）
echo some_garbage > /tmp/2/objects/66WuRH0s0albWDZ9nTmjFo9JIqTTXmB6EiRkhTh1zeA=.*

#执行检查并修复
 STORAGE_ROOT=/tmp/3 go run ../maintenance/objectScanner/objectScanner.go
 
 > return :
[2023-08-18 16:54:21 :main.verify(objectScanner.go):24]
verify:66WuRH0s0albWDZ9nTmjFo9JIqTTXmB6EiRkhTh1zeA=
[2023-08-18 16:54:21 10.29.1.2:12345:go-distributed-oss/dataServer/objects.sendFile(get.go):86]
gzip: invalid header
[2023-08-18 16:54:21 10.29.1.2:12345:go-distributed-oss/dataServer/objects.getFile(get.go):66]
object hash mismatch, remove/tmp/2/objects/66WuRH0s0albWDZ9nTmjFo9JIqTTXmB6EiRkhTh1zeA=.1.2eKLvcHfGvzIi+X5HFzgiCJyqjXV9%2F2U08FC6Srcslg=
[2023-08-18 16:54:21 10.29.1.1:12345:go-distributed-oss/dataServer/objects.getFile(get.go):57]
object shard 66WuRH0s0albWDZ9nTmjFo9JIqTTXmB6EiRkhTh1zeA=.2 not found in 10.29.1.1:12345, just found []

#再次执行检查
STORAGE_ROOT=/tmp/3 go run ../maintenance/objectScanner/objectScanner.go

> return :
[2023-08-18 16:55:50 :main.verify(objectScanner.go):24]
verify:66WuRH0s0albWDZ9nTmjFo9JIqTTXmB6EiRkhTh1zeA=
```



```sh
# 分片丢失/损坏过多（无法恢复）
rm /tmp/1/objects/66WuRH0s0albWDZ9nTmjFo9JIqTTXmB6EiRkhTh1zeA=*
echo some_garbage > /tmp/2/objects/66WuRH0s0albWDZ9nTmjFo9JIqTTXmB6EiRkhTh1zeA=.*
echo some_garbage > /tmp/3/objects/66WuRH0s0albWDZ9nTmjFo9JIqTTXmB6EiRkhTh1zeA=.*

> return :
[2023-08-18 16:57:17 :main.verify(objectScanner.go):24]
verify:66WuRH0s0albWDZ9nTmjFo9JIqTTXmB6EiRkhTh1zeA=
[2023-08-18 16:57:17 10.29.1.3:12345:go-distributed-oss/dataServer/objects.sendFile(get.go):86]
gzip: invalid header		# 数据异常，dataSever解压时未通过校验
[2023-08-18 16:57:17 10.29.1.3:12345:go-distributed-oss/dataServer/objects.getFile(get.go):66]
object hash mismatch, remove/tmp/3/objects/66WuRH0s0albWDZ9nTmjFo9JIqTTXmB6EiRkhTh1zeA=.0.xPZ9Cf8mShrJsL32FnbSVcayc9W5Y05clRo3GOkLyG0=
[2023-08-18 16:57:17 10.29.1.2:12345:go-distributed-oss/dataServer/objects.sendFile(get.go):86]
gzip: invalid header		# 数据异常，dataSever解压时未通过校验
[2023-08-18 16:57:17 10.29.1.2:12345:go-distributed-oss/dataServer/objects.getFile(get.go):66]
object hash mismatch, remove/tmp/2/objects/66WuRH0s0albWDZ9nTmjFo9JIqTTXmB6EiRkhTh1zeA=.1.2eKLvcHfGvzIi+X5HFzgiCJyqjXV9%2F2U08FC6Srcslg=
[2023-08-18 16:57:17 10.29.1.1:12345:go-distributed-oss/dataServer/objects.getFile(get.go):57]
object shard 66WuRH0s0albWDZ9nTmjFo9JIqTTXmB6EiRkhTh1zeA=.2 not found in 10.29.1.1:12345, just found []	# 分片丢失
[2023-08-18 16:57:17 :main.verify(objectScanner.go):37]
object hash mismatch,calculated=47DEQpj8HBSa+/TImW+5JCeuQeRkm5NMpJWZG3hSuFU=,requested=66WuRH0s0albWDZ9nTmjFo9JIqTTXmB6EiRkhTh1zeA=
[2023-08-18 16:57:17 10.29.1.3:12345:go-distributed-oss/dataServer/temp.put(put.go):48]
actual size mismatch,expect:8,but actual:0	#尝试恢复，但是恢复失败
[2023-08-18 16:57:17 10.29.1.2:12345:go-distributed-oss/dataServer/temp.put(put.go):48]
actual size mismatch,expect:8,but actual:0
[2023-08-18 16:57:17 10.29.1.1:12345:go-distributed-oss/dataServer/temp.put(put.go):48]
actual size mismatch,expect:8,but actual:0
```



