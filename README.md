## radcast
### radikoを録音し、podcast配信する

radicastを少し、自分好みに改造しました
* 無劣化録音
* チャンネルのアイコン設定
* 番組のアイコン取得
* 他

ORIGINAL By https://github.com/soh335/radicast

## 必要パッケージ
* ffmpeg

## インストール
```
$ go install github.com/omiso46/radcast@latest
```

## 使い方
### 設定ファイル
```
$ radcast --setup > config.json
$ vim config.json

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
録音ファイルは個人使用の範囲内で。絶対！
