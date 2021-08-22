#!/bin/bash
echo -n "Build to the server"
ssh root@192.168.0.2 systemctl stop ForkServer.service
echo -n "."
GOARCH=arm go build -o /run/user/1000/gvfs/smb-share:server=oms,share=fs/ForkServer.daemon -ldflags="-s -w"
echo -n "."
ssh root@192.168.0.2 chmod +x /opt/ForkServer/ForkServer.daemon
echo -n "."
ssh root@192.168.0.2 systemctl start ForkServer.service
echo "done!"