# 使用docker启动，发送交易，停止docker的测试

echo "mkdir tmp_solo"
mkdir -p tmp_solo/data
mkdir -p tmp_solo/log

echo
echo "run docker"
docker run   \
 -itd     \
 -p 11301:11301   \
 -p 12301:12301   \
 -e TZ=Asia/Shanghai   \
 -v ./config/solo/wx-org1.chainmaker.org:/chainmaker-go/config/wx-org1.chainmaker.org \
 -v ./tmp_solo/data:/chainmaker-go/data \
 -v ./tmp_solo/log:/chainmaker-go/log \
 --privileged=true  \
 --name csolo  \
 chainmakerofficial/chainmaker:v2.2.1_arm \
 bash -c "./chainmaker start -c ../config/wx-org1.chainmaker.org/chainmaker.yml > panic.log"


echo
echo "sendTx"
./sendTx.sh

echo "docker stop csolo"
docker stop csolo