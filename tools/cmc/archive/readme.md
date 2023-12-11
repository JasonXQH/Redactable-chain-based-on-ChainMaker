# cmc 归档中心命令

## cmc归档数据命令
```powershell
./cmc archive dump 100 --sdk-conf-path ./testdata/sdk_config.yml --mode quick

```
## cmc查询归档中心当前已归档数据状态命令
```powershell
./cmc archive query archived-status --sdk-conf-path ./testdata/sdk_config.yml
```

## cmc查询链上当前已归档数据状态命令

```powershell
./cmc archive query chain-archived-status --sdk-conf-path ./testdata/sdk_config.yml
```


## cmc查询归档中心,根据高度查区块命令
```powershell
./cmc archive query block-by-height 20  --sdk-conf-path ./testdata/sdk_config.yml
```

## cmc查询归档中心,根据txId查询区块命令
```powershell
./cmc archive query block-by-txid 17221b132a25209aca52fdfc07218265e4377ef0099d46a49edfd032001fc2be --sdk-conf-path ./testdata/sdk_config.yml
```

## cmc查询归档中心,根据txId查询交易命令
```powershell
./cmc archive query tx 17221b132a25209aca52fdfc07218265e4377ef0099d46a49edfd032001fc2be  --sdk-conf-path ./testdata/sdk_config.yml
```

## cmc查询归档中心,根据block hash查询区块命令

```powershell
./cmc archive query block-by-hash 17221b132a25209aca52fdfc07218265e4377ef0099d46a49edfd032001fc2be --sdk-conf-path ./testdata/sdk_config.yml
```

## cmc清理链上区块数据命令
```powershell
./cmc archive archive 10000 --timeout 20 --sdk-conf-path ./testdata/sdk_config.yml 
```  

## restore 恢复链上区块数据命令,恢复到指定高度
```powershell
./cmc archive restore 3000 --timeout 20 --sdk-conf-path ./testdata/sdk_config.yml 
```