#!/bin/bash

for i in `seq 1 6`
do
    mkdir -p /tmp/$i/objects
    mkdir -p /tmp/$i/temp
    mkdir -p /tmp/$i/garbage
done

#设置8个虚拟地址
sudo ifconfig lo:1 10.29.1.1/16
sudo ifconfig lo:2 10.29.1.2/16
sudo ifconfig lo:3 10.29.1.3/16
sudo ifconfig lo:4 10.29.1.4/16
sudo ifconfig lo:5 10.29.1.5/16
sudo ifconfig lo:6 10.29.1.6/16
sudo ifconfig lo:7 10.29.2.1/16
sudo ifconfig lo:8 10.29.2.2/16