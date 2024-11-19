# aiwolf-nlp-server

人狼知能コンテスト2024冬季 国内大会（自然言語部門） のゲームサーバです。  
従来の対戦接続システムでは、対戦接続システムが実行中のサーバに対してSSH接続を確立した後、TCP通信を行う必要がありましたが、新しいゲームサーバでは、WebSocketを使用して直接ゲームサーバと通信を行います。

サンプルエージェントについては、[kano-lab/aiwolf-nlp-agent](https://github.com/kano-lab/aiwolf-nlp-agent) を参考にしてください。
大会の詳細ならびに参加登録については、[AIWolfDial2024WinterJp](https://sites.google.com/view/aiwolfdial2024winterjp/) を参考にしてください。

## ドキュメント

従来の対戦接続システムのロジックを参考に実装していますが、一部仕様が変更されています。  
また、開発中であるため、一部に不具合やバグがある可能性があります。

- [プロトコルの実装について](./doc/protocol.md)
- [ゲームロジックの実装について](./doc/logic.md)

## 実行方法

人狼知能コンテスト2024冬季 国内大会（自然言語部門） で使用予定の設定でサーバを起動する方法になります。

デフォルトのサーバアドレスは `127.0.0.1:8080` です。エージェントプログラムの接続先には、このアドレスを指定してください。  
同じチーム名のエージェント同士のみをマッチングさせる自己対戦モードは、デフォルトで有効になっています。そのため、異なるチーム名のエージェント同士をマッチングさせる場合は、設定ファイルを変更してください。

### Linux

```bash
curl -LJO https://github.com/kano-lab/aiwolf-nlp-server/releases/latest/download/aiwolf-nlp-server-linux-amd64
curl -LJO https://github.com/kano-lab/aiwolf-nlp-server/releases/latest/download/default.yml
chmod u+x ./aiwolf-nlp-server-linux-amd64
./aiwolf-nlp-server-linux-amd64
```

### Windows

```bash
curl -LJO https://github.com/kano-lab/aiwolf-nlp-server/releases/latest/download/aiwolf-nlp-server-windows-amd64.exe
curl -LJO https://github.com/kano-lab/aiwolf-nlp-server/releases/latest/download/default.yml
.\aiwolf-nlp-server-windows-amd64.exe
```

### macOS (Intel)

> [!NOTE]
> 開発元が不明なアプリケーションとしてブロックされる場合があります。  
> 下記サイトを参考に、実行許可を与えてください  
> https://support.apple.com/ja-jp/guide/mac-help/mh40616/mac

> [!WARNING]
> 動作確認が取れていないため、動作しない可能性があります。

```bash
curl -LJO https://github.com/kano-lab/aiwolf-nlp-server/releases/latest/download/aiwolf-nlp-server-darwin-amd64
curl -LJO https://github.com/kano-lab/aiwolf-nlp-server/releases/latest/download/default.yml
chmod u+x ./aiwolf-nlp-server-darwin-amd64
./aiwolf-nlp-server-darwin-amd64
```

### macOS (Apple Silicon)

> [!NOTE]
> 開発元が不明なアプリケーションとしてブロックされる場合があります。  
> 下記サイトを参考に、実行許可を与えてください  
> https://support.apple.com/ja-jp/guide/mac-help/mh40616/mac

```bash
curl -LJO https://github.com/kano-lab/aiwolf-nlp-server/releases/latest/download/aiwolf-nlp-server-darwin-arm64
curl -LJO https://github.com/kano-lab/aiwolf-nlp-server/releases/latest/download/default.yml
chmod u+x ./aiwolf-nlp-server-darwin-arm64
./aiwolf-nlp-server-darwin-arm64
```
