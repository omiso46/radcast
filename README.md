# radcast

■2021/01のradikoの仕様変更（HLSストリーミング）に対応しました。

radikoを録音し、podcast配信する

radicastを少し、自分好みに改造しました
* 無劣化録音
* チャンネルのアイコン設定
* 番組のアイコン取得
* 05時開始番組が録音できないときがある
* 他

ORIGINAL By https://github.com/soh335/radicast


## 必要パッケージ

* ffmpeg
* <s>or avconv</s> ← avconvは検証環境がないので対象外に
* <s>rtmpdump</s> ← HLSストリーミングの対応で不要に
* <s>swftools</s> ← flash廃止の対応で不要に

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
    "00 17 * * *"
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

## お約束
録音ファイルは個人使用の範囲内で。絶対。
