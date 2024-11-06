# プロトコルの実装について

このドキュメントでは、プロトコルの実装について説明します。  
ここで指すプロトコルとは、技術レイヤーのプロトコルではなく、人狼知能のエージェントとサーバ間のやり取りの際の文字列としてのプロトコルです。  
また、従来の人狼知能対戦接続システムでは、エージェント側がサーバとして待ち受けするため、ゲームマスタ側をサーバと呼ばずに対戦接続システムと表記していましたが、WebSocketを使用したこのシステムでは、ゲームマスタ側がサーバとして待ち受けをするため、ゲームマスタ側をサーバ、エージェント側をクライアントと表記します。

## プロトコルの概要

サーバがエージェントに送るメッセージは、すべてJSON形式の文字列です。  
それに対して、エージェントがサーバに送るメッセージは、すべて生の文字列です。  
また、このドキュメントではサーバがエージェントに送るメッセージをリクエスト、エージェントがサーバに送るメッセージをレスポンスと表記します。

### リクエストの概要

リクエストの種類は、`model/request.go` に定義されています。  
- 名前リクエスト `NAME`
- ゲーム開始リクエスト `INITIALIZE`
- トークリクエスト `TALK`
- 囁きリクエスト `WHISPER`
- 投票リクエスト `VOTE`
- 占いリクエスト `DIVINE`
- 護衛リクエスト `GUARD`
- 襲撃リクエスト `ATTACK`
- 昼開始リクエスト `DAILY_INITIALIZE`
- 昼終了リクエスト `DAILY_FINISH`
- ゲーム終了リクエスト `FINISH`

リクエストの種類によって、リクエストに含まれる情報が異なり、レスポンスを返す必要があるかどうかも異なります。

### レスポンスの概要

レスポンスは、トークや囁きリクエストに対してエージェントが発する自然言語を返す場合（例: `こんにちは`）と、投票や占いリクエストなどに対して対象のエージェントのインデックス付き文字列（例: `Agent[01]`）を返す２種類があります。

## プロトコルの詳細

各リクエストについて、実際の例を示しながら説明します。

### 名前リクエスト (NAME)

名前リクエストは、エージェントがサーバに接続した際に送信されるリクエストです。  
エージェントは、このリクエストを受信した際に、自身の名前を返す必要があります。  
複数エージェントを接続する場合、後ろにユニークな数字をつける必要があります。  
例えば、`dpetektq` という名前を返す場合、`dpetektq1`、`dpetektq2` などとします。  
後ろの数字を除いた名前は、エージェントのチーム名として扱われます。

```
2024/11/06 04:53:29 INFO NAMEパケットを送信しました remote_addr=127.0.0.1:41072
    dummy_client.go:49: recv: {"request":"NAME"}
    dummy_client.go:68: send: dpetektq
2024/11/06 04:53:29 INFO クライアントが接続しました team=dpetektq name=dpetektq remote_addr=127.0.0.1:41056
2024/11/06 04:53:29 INFO 新しいクライアントが待機部屋に追加されました team=dpetektq remote_addr=127.0.0.1:41056
```

### ゲーム開始リクエスト (INITIALIZE)

ゲーム開始リクエストは、ゲームが開始された際に送信されるリクエストです。  
エージェントは、このリクエストを受信した際に、何も返す必要はありません。

**ゲームの現状態を示す情報 (info)**
- statusMap: 各エージェントの生存状態を示すマップ
- roleMap: 各エージェントの役職を示すマップ (自分以外のエージェントの役職は見えません)
- remainTalkMap: 現リクエスト時における各エージェントの残り発言数を示すマップ
- remainWhisperMap: 現リクエスト時における各エージェントの残り囁き数を示すマップ
- day: 現在の日数
- agent: 自分のエージェントインデックス

**ゲームの設定を示す情報 (setting)**
- roleNumMap: 各役職の人数を示すマップ
- maxTalk: 1日あたりの1エージェントの最大発言数 (トーク)
- maxTalkTurn: 1日あたりの全体の発言回数 (トーク)
- maxWhisper: 1日あたりの1エージェントの最大囁き数
- maxWhisperTurn: 1日あたりの全体の囁き回数
- maxSkip: 1日あたりの全体のスキップ回数 (トークと囁きのスキップ回数は区別してカウントされる)
- isEnableNoAttack: 襲撃なしの日を許可するか
- isVoteVisible: 投票の結果を公開するか
- isTalkOnFirstDay: 1日目の発言を許可するか
- responseTimeout: エージェントのアクションのタイムアウト時間
- actionTimeout: エージェントの生存確認のタイムアウト時間
- maxRevote: 1位タイの場合の最大再投票回数
- maxAttackRevote: 1位タイの場合の最大襲撃再投票回数

```
2024/11/06 05:22:04 INFO パケットを送信しました agent=Agent[02] packet="{Request:INITIALIZE Info:0xc000184210 Settings:0xc0000b0430 TalkHistory:[] WhisperHistory:[]}"
    dummy_client.go:49: recv: {"request":"INITIALIZE","info":{"statusMap":{"Agent[01]":"ALIVE","Agent[02]":"ALIVE","Agent[03]":"ALIVE","Agent[04]":"ALIVE","Agent[05]":"ALIVE"},"roleMap":{"Agent[02]":"SEER"},"remainTalkMap":{},"remainWhisperMap":{},"day":0,"agent":"Agent[02]"},"setting":{"roleNumMap":{"BODYGUARD":0,"MEDIUM":0,"POSSESSED":0,"SEER":1,"VILLAGER":3,"WEREWOLF":1},"maxTalk":3,"maxTalkTurn":15,"maxWhisper":3,"maxWhisperTurn":15,"maxSkip":3,"isEnableNoAttack":true,"isVoteVisible":false,"isTalkOnFirstDay":true,"responseTimeout":90000,"actionTimeout":60000,"maxRevote":1,"maxAttackRevote":1}}
```

### 昼開始リクエスト (DAILY_INITIALIZE)

昼開始リクエストは、昼が開始された際、つまり次の日が始まった際に送信されるリクエストです。  
エージェントは、このリクエストを受信した際に、何も返す必要はありません。  
各キーについては、ゲーム開始リクエストと同様です。

```
2024/11/06 05:22:04 INFO 昼を開始します id=01JBZYREPRWF177KSQD6KJFF9P day=0
2024/11/06 05:22:04 INFO パケットを送信しました agent=Agent[01] packet="{Request:DAILY_INITIALIZE Info:0xc000185130 Settings:0xc0000b0430 TalkHistory:[] WhisperHistory:[]}"
    dummy_client.go:49: recv: {"request":"DAILY_INITIALIZE","info":{"statusMap":{"Agent[01]":"ALIVE","Agent[02]":"ALIVE","Agent[03]":"ALIVE","Agent[04]":"ALIVE","Agent[05]":"ALIVE"},"roleMap":{"Agent[01]":"WEREWOLF"},"remainTalkMap":{},"remainWhisperMap":{},"day":0,"agent":"Agent[01]"},"setting":{"roleNumMap":{"BODYGUARD":0,"MEDIUM":0,"POSSESSED":0,"SEER":1,"VILLAGER":3,"WEREWOLF":1},"maxTalk":3,"maxTalkTurn":15,"maxWhisper":3,"maxWhisperTurn":15,"maxSkip":3,"isEnableNoAttack":true,"isVoteVisible":false,"isTalkOnFirstDay":true,"responseTimeout":90000,"actionTimeout":60000,"maxRevote":1,"maxAttackRevote":1}}
```

