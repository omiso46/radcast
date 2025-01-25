# radcast
■ 公開停止<br>
2025/01/20以降radikoの録音ができません<br>
メンテの気力が湧かないので公開停止をすることにしました<br>
<br>
■ 2021/01のradikoの仕様変更（HLSストリーミング）に対応しました。<br>
<br><br>
radikoを録音し、podcast配信する
<br>
radicastを少し、自分好みに改造しました<br>
* 無劣化録音<br>
* チャンネルのアイコン設定<br>
* 番組のアイコン取得<br>
* 05時開始番組が録音できないときがある<br>
* 他<br>
<br>
ORIGINAL By https://github.com/soh335/radicast<br>
<br><br>
## 必要パッケージ

* ffmpeg
* sh ← なぜかgolangからffmpegが直接起動できない（sh経由で起動させる）
* <s>or avconv</s> ← avconvは検証環境がないので対象外に
* <s>rtmpdump</s> ← HLSストリーミングの対応で不要に
* <s>swftools</s> ← flash廃止の対応で不要に

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
