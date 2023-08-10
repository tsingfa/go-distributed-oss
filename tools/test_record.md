# 分布式对象存储系统--测试记录

## 1





## 2





## 3





##  4









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

