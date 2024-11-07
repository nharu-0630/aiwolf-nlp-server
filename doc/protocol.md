# プロトコルの実装について

このドキュメントでは、プロトコルの実装について説明します。  
ここで指すプロトコルとは、技術レイヤーのプロトコルではなく、人狼知能のエージェントとサーバ間のやり取りの際の文字列としてのプロトコルです。  
また、従来の人狼知能対戦接続システムでは、エージェント側がサーバとして待ち受けするため、ゲームマスタ側をサーバと呼ばずに対戦接続システムと表記していましたが、WebSocketを使用したこのシステムでは、ゲームマスタ側がサーバとして待ち受けをするため、ゲームマスタ側をサーバ、エージェント側をクライアントと表記します。

## プロトコルの概要

サーバがエージェントに送るメッセージは、すべてJSON形式の文字列です。  
それに対して、エージェントがサーバに送るメッセージは、すべて生の文字列です。  
また、このドキュメントではサーバがエージェントに送るメッセージをリクエスト、エージェントがサーバに送るメッセージをレスポンスと表記します。

### リクエストの概要

- 名前リクエスト `NAME`
- ゲーム開始リクエスト `INITIALIZE`
- 昼開始リクエスト `DAILY_INITIALIZE`
- 囁きリクエスト `WHISPER`
- トークリクエスト `TALK`
- 昼終了リクエスト `DAILY_FINISH`
- 占いリクエスト `DIVINE`
- 護衛リクエスト `GUARD`
- 投票リクエスト `VOTE`
- 襲撃リクエスト `ATTACK`
- ゲーム終了リクエスト `FINISH`

リクエストの種類によって、リクエストに含まれる情報が異なり、レスポンスを返す必要があるかどうかも異なります。  
詳細な実装については、[request.go](../model/request.go)と[packet.go](../model/packet.go)を参照してください。

### レスポンスの概要

レスポンスは、トークや囁きリクエストに対してエージェントが発する自然言語を返す場合 (例: `こんにちは`) と、投票や占いリクエストなどに対して対象のエージェントのインデックス付き文字列 (例: `Agent[01]`) を返す２種類があります。

## リクエストの構造

各リクエストについて、実際の例を示しながら説明します。  
リクエストには必ず `request` というキーが含まれ、その値がリクエストの種類を示します。  
他のキーは、以下の通りです。
- info: ゲームの現状態を示す情報
- setting: ゲームの設定を示す情報
- talkHistory: トークの履歴を示す情報
- whisperHistory: 囁きの履歴を示す情報

### ゲームの現状態を示す情報 (info)

- day: 現在の日数
- agent: 自分のエージェントのインデックス付き文字列
- mediumResult: 霊能者の結果 (エージェントの役職が霊媒師であるかつ霊能結果が設定されている場合のみ)
- divineResult: 占い師の結果 (エージェントの役職が占い師であるかつ占い結果が設定されている場合のみ)
- executedAgent: 昨日の追放結果 (エージェントが追放された場合のみ)
- attackedAgent: 昨夜の襲撃結果 (エージェントが襲撃された場合のみ)
- voteList: 投票の結果 (投票結果が公開されている場合のみ)
- attackVoteList: 襲撃の投票結果 (エージェントの役職が人狼かつ襲撃投票結果が公開されている場合のみ)
- statusMap: 各エージェントの生存状態を示すマップ
- roleMap: 各エージェントの役職を示すマップ (自分以外のエージェントの役職は見えません)
- remainTalkMap: 現リクエスト時における各エージェントの残り発言数を示すマップ
- remainWhisperMap: 現リクエスト時における各エージェントの残り囁き数を示すマップ

### ゲームの設定を示す情報 (setting)

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

### 会話の履歴を示す情報 (talkHistory / whisperHistory)

- idx: 会話のインデックス
- day: 会話が行われた日数
- turn: 会話が行われたターン数
- agent: 会話を行ったエージェント
- text: 会話の内容

## リクエストの種類

リクエストの種類ごとの詳細な説明を以下に示します。

### 名前リクエスト (NAME)

名前リクエストは、エージェントがサーバに接続した際に送信されるリクエストです。  
エージェントは、このリクエストを受信した際に、自身の名前を返す必要があります。  
複数エージェントを接続する場合、後ろにユニークな数字をつける必要があります。  
例えば、 `dpetektq` という名前を返す場合、 `dpetektq1`, `dpetektq2` などとします。  
後ろの数字を除いた名前は、エージェントのチーム名として扱われます。

```
2024/11/07 06:42:27 INFO NAMEパケットを送信しました remote_addr=127.0.0.1:36266
    dummy_client.go:50: recv: {"request":"NAME"}
    dummy_client.go:69: send: dpetektq
2024/11/07 06:42:27 INFO クライアントが接続しました team=dpetektq name=dpetektq remote_addr=127.0.0.1:36266
2024/11/07 06:42:27 INFO 新しいクライアントが待機部屋に追加されました team=dpetektq remote_addr=127.0.0.1:36266
```

### ゲーム開始リクエスト (INITIALIZE)

ゲーム開始リクエストは、ゲームが開始された際に送信されるリクエストです。  
エージェントは、このリクエストを受信した際に、何も返す必要はありません。

**ゲームの現状態を示す情報 (info)**  
**ゲームの設定を示す情報 (setting)**

```
2024/11/06 05:22:04 INFO パケットを送信しました agent=Agent[02] packet="{Request:INITIALIZE Info:0xc000184210 Settings:0xc0000b0430 TalkHistory:[] WhisperHistory:[]}"
    dummy_client.go:49: recv: {"request":"INITIALIZE","info":{"statusMap":{"Agent[01]":"ALIVE","Agent[02]":"ALIVE","Agent[03]":"ALIVE","Agent[04]":"ALIVE","Agent[05]":"ALIVE"},"roleMap":{"Agent[02]":"SEER"},"remainTalkMap":{},"remainWhisperMap":{},"day":0,"agent":"Agent[02]"},"setting":{"roleNumMap":{"BODYGUARD":0,"MEDIUM":0,"POSSESSED":0,"SEER":1,"VILLAGER":3,"WEREWOLF":1},"maxTalk":3,"maxTalkTurn":15,"maxWhisper":3,"maxWhisperTurn":15,"maxSkip":3,"isEnableNoAttack":true,"isVoteVisible":false,"isTalkOnFirstDay":true,"responseTimeout":90000,"actionTimeout":60000,"maxRevote":1,"maxAttackRevote":1}}
```

### 昼開始リクエスト (DAILY_INITIALIZE)

昼開始リクエストは、昼が開始された際、つまり次の日が始まった際に送信されるリクエストです。  
エージェントは、このリクエストを受信した際に、何も返す必要はありません。  

**ゲームの現状態を示す情報 (info)**  
**ゲームの設定を示す情報 (setting)**

```
2024/11/07 06:42:27 INFO パケットを送信しました agent=Agent[01] packet="{Request:DAILY_INITIALIZE Info:0xc0000b16b0 Settings:0xc0000ed020 TalkHistory:<nil> WhisperHistory:<nil>}"
    dummy_client.go:50: recv: {"request":"DAILY_INITIALIZE","info":{"statusMap":{"Agent[01]":"ALIVE","Agent[02]":"ALIVE","Agent[03]":"ALIVE","Agent[04]":"ALIVE","Agent[05]":"ALIVE"},"roleMap":{"Agent[01]":"VILLAGER"},"remainTalkMap":{},"remainWhisperMap":{},"day":0,"agent":"Agent[01]"},"setting":{"roleNumMap":{"BODYGUARD":1,"MEDIUM":0,"POSSESSED":0,"SEER":1,"VILLAGER":2,"WEREWOLF":1},"maxTalk":3,"maxTalkTurn":15,"maxWhisper":3,"maxWhisperTurn":15,"maxSkip":3,"isEnableNoAttack":true,"isVoteVisible":false,"isTalkOnFirstDay":true,"responseTimeout":90000,"actionTimeout":60000,"maxRevote":1,"maxAttackRevote":1}}
```

### 囁きリクエスト (WHISPER) / トークリクエスト (TALK)

囁きリクエストとトークリクエストは、それぞれ囁きとトークが要求された際に送信されるリクエストです。  
囁きリクエストについては、人狼の役職が2人以上生存している場合に、人狼のみに送信されます。  
エージェントは、このリクエストを受信した際に、囁きやトークの自然言語の文字列を返す必要があります。  
サーバ側が送信する履歴は、前回のエージェントに対する送信の差分のみであり、全ての履歴を送信するわけではありません。

**トークの履歴を示す情報 (talkHistory)**  
**囁きの履歴を示す情報 (whisperHistory)**

```
2024/11/07 06:42:27 INFO パケットを送信しました agent=Agent[01] packet="{Request:TALK Info:<nil> Settings:<nil> TalkHistory:0xc000528138 WhisperHistory:<nil>}"
    dummy_client.go:50: recv: {"request":"TALK","talkHistory":[{"idx":0,"day":0,"turn":0,"agent":"Agent[04]","text":"58ef003db5f9e64550b1fc32782ed9d3"}]}
    dummy_client.go:69: send: c778b90c7405d605a385c0931e347cae
2024/11/07 06:42:27 INFO レスポンスを受信しました agent=Agent[01] response=c778b90c7405d605a385c0931e347cae
2024/11/07 06:42:27 INFO 発言がオーバーもしくはスキップではないため、スキップ回数をリセットしました id=01JC2NRBXKKASYNDHS84FTYHXA agent=Agent[01]
2024/11/07 06:42:27 INFO 発言を受信しました id=01JC2NRBXKKASYNDHS84FTYHXA agent=Agent[01] text=c778b90c7405d605a385c0931e347cae
2024/11/07 06:42:27 INFO パケットを送信しました agent=Agent[02] packet="{Request:TALK Info:<nil> Settings:<nil> TalkHistory:0xc000012990 WhisperHistory:<nil>}"
    dummy_client.go:50: recv: {"request":"TALK","talkHistory":[{"idx":0,"day":0,"turn":0,"agent":"Agent[04]","text":"58ef003db5f9e64550b1fc32782ed9d3"},{"idx":1,"day":0,"turn":0,"agent":"Agent[01]","text":"c778b90c7405d605a385c0931e347cae"}]}
    dummy_client.go:69: send: bcf2a95dbd72aa885bafe5e54d067a5b
2024/11/07 06:42:27 INFO レスポンスを受信しました agent=Agent[02] response=bcf2a95dbd72aa885bafe5e54d067a5b
```

### 昼終了リクエスト (DAILY_FINISH)

昼終了リクエストは、昼が終了された際、つまりその日の夜が始まった際に送信されるリクエストです。  
エージェントは、このリクエストを受信した際に、何も返す必要はありません。  
直前までの会話の履歴が送信されます。  
ゲーム全体の人狼の役職が2人未満で囁きフェーズが存在しない場合においても、人狼の役職に対しては、囁きの履歴が送信されます。

**トークの履歴を示す情報 (talkHistory)**  
**囁きの履歴を示す情報 (whisperHistory)**

```
2024/11/07 06:42:27 INFO パケットを送信しました agent=Agent[03] packet="{Request:DAILY_FINISH Info:<nil> Settings:<nil> TalkHistory:0xc0004b23f0 WhisperHistory:0xc0004b2408}"
    dummy_client.go:50: recv: {"request":"DAILY_FINISH","talkHistory":[{"idx":14,"day":0,"turn":2,"agent":"Agent[03]","text":"Over"}],"whisperHistory":[]}
```

### 占いリクエスト (DIVINE)

占いリクエストは、占いが要求された際に送信されるリクエストです。  
占い師のみに送信されます。  
エージェントは、このリクエストを受信した際に、占いの対象となるエージェントのインデックス付き文字列を返す必要があります。

```
2024/11/07 06:42:27 INFO パケットを送信しました agent=Agent[04] packet="{Request:DIVINE Info:<nil> Settings:<nil> TalkHistory:<nil> WhisperHistory:<nil>}"
    dummy_client.go:50: recv: {"request":"DIVINE"}
    dummy_client.go:69: send: Agent[03]
2024/11/07 06:42:27 INFO レスポンスを受信しました agent=Agent[04] response=Agent[03]
```

### 護衛リクエスト (GUARD)

護衛リクエストは、護衛が要求された際に送信されるリクエストです。  
騎士のみに送信されます。
エージェントは、このリクエストを受信した際に、護衛の対象となるエージェントのインデックス付き文字列を返す必要があります。

```
2024/11/07 06:42:27 INFO パケットを送信しました agent=Agent[05] packet="{Request:GUARD Info:<nil> Settings:<nil> TalkHistory:<nil> WhisperHistory:<nil>}"
    dummy_client.go:50: recv: {"request":"GUARD"}
    dummy_client.go:69: send: Agent[05]
2024/11/07 06:42:27 INFO レスポンスを受信しました agent=Agent[05] response=Agent[05]
```

### 投票リクエスト (VOTE)

投票リクエストは、追放するエージェントを投票する際に送信されるリクエストです。  
エージェントは、このリクエストを受信した際に、投票の対象となるエージェントのインデックス付き文字列を返す必要があります。

```
2024/11/07 06:42:27 INFO パケットを送信しました agent=Agent[01] packet="{Request:VOTE Info:<nil> Settings:<nil> TalkHistory:<nil> WhisperHistory:<nil>}"
    dummy_client.go:50: recv: {"request":"VOTE"}
    dummy_client.go:69: send: Agent[01]
2024/11/07 06:42:27 INFO レスポンスを受信しました agent=Agent[01] response=Agent[01]
```

### 襲撃リクエスト (ATTACK)

襲撃リクエストは、襲撃するエージェントを投票する際に送信されるリクエストです。  
人狼のみに送信されます。  
エージェントは、このリクエストを受信した際に、襲撃の対象となるエージェントのインデックス付き文字列を返す必要があります。  
直前までの会話の履歴が送信されます。  
ゲーム全体の人狼の役職が2人未満で囁きフェーズが存在しない場合においても、人狼の役職に対しては、囁きの履歴が送信されます。

**会話の履歴を示す情報 (whisperHistory)**

```
2024/11/07 06:42:27 INFO パケットを送信しました agent=Agent[03] packet="{Request:ATTACK Info:<nil> Settings:<nil> TalkHistory:<nil> WhisperHistory:0xc0004b2228}"
    dummy_client.go:50: recv: {"request":"ATTACK","whisperHistory":[]}
    dummy_client.go:69: send: Agent[02]
2024/11/07 06:42:27 INFO レスポンスを受信しました agent=Agent[03] response=Agent[02]
```

### ゲーム終了リクエスト (FINISH)

ゲーム終了リクエストは、ゲームが終了された際に送信されるリクエストです。  
エージェントは、このリクエストを受信した際に、何も返す必要はありません。  
各キーについては、ゲーム開始リクエストと同様です。ゲーム開始リクエストとは異なり、 `setting` は送信されません。

**ゲームの現状態を示す情報 (info)**  
なお、`roleMap` は自分以外も含めたすべてのエージェントの役職が含まれます。

```
2024/11/07 06:42:27 INFO パケットを送信しました agent=Agent[01] packet="{Request:FINISH Info:0xc0000b1130 Settings:<nil> TalkHistory:<nil> WhisperHistory:<nil>}"
    dummy_client.go:50: recv: {"request":"FINISH","info":{"statusMap":{"Agent[01]":"DEAD","Agent[02]":"DEAD","Agent[03]":"ALIVE","Agent[04]":"DEAD","Agent[05]":"DEAD"},"roleMap":{"Agent[01]":"VILLAGER","Agent[02]":"VILLAGER","Agent[03]":"WEREWOLF","Agent[04]":"SEER","Agent[05]":"BODYGUARD"},"remainTalkMap":{},"remainWhisperMap":{},"day":4,"agent":"Agent[01]","executedAgent":"Agent[01]"}}
```