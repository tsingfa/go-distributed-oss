# 分布式对象存储系统--测试记录

## 1 单机对象存储





## 2 可扩展的分布式系统





## 3 元数据服务





##  4 数据校验和去重









## 5 数据冗余和即时修复



```shell
# 获取sha256摘要
echo -n "this object will be separate to 4+2 shards"|openssl dgst -sha256 -binary|base64

>  return:
MBMxWHrPMsuOBaVYHkwScZQRyTRMQyiKp2oelpLZza8=
```



```shell
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
this object will be separate to 4+2 shards

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