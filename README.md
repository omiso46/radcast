# radcast

radikoを録音し、podcast配信する

radicastを少し、自分好みに改造しました
* 無劣化録音
* 05時開始番組が録音できないときがある
* チャンネルのアイコン設定
* 番組のアイコン取得
* 他

ORIGINAL By https://github.com/soh335/radicast


## 必要パッケージ

* rtmpdump
* <s>swftools</s> ← 2020/12のradiko仕様変更の対応で不要に
* ffmpeg or avconv

## インストール

```
$ go get github.com/omiso46/radcast
```

## 使い方

### 設定ファイル

```
$ radcast --setup > config.json
$ vim config.json
$ cat config.json

{
  "FMT": [
    "0 0 17 * * *"
  ]
}
```

cron specification is [here](https://godoc.org/github.com/robfig/cron#hdr-CRON_Expression_Format)

### 設定ファイルのリロード

```
$ kill -HUP nnn
```

## LICENSE

* MIT
