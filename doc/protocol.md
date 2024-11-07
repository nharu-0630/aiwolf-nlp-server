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

### レスポンスの概要

レスポンスは、トークや囁きリクエストに対してエージェントが発する自然言語を返す場合 (例: `こんにちは`) と、投票や占いリクエストなどに対して対象のエージェントのインデックス付き文字列 (例: `Agent[01]`) を返す２種類があります。

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

### 囁きリクエスト (WHISPER) / トークリクエスト (TALK)

囁きリクエストとトークリクエストは、それぞれ囁きとトークが要求された際に送信されるリクエストです。  
囁きリクエストについては、人狼の役職が2人以上生存している場合に、人狼 (`WEREWOLF`) のみに送信されます。  
エージェントは、このリクエストを受信した際に、囁きやトークの自然言語の文字列を返す必要があります。  
サーバ側が送信する履歴は、前回のエージェントに対する送信の差分のみであり、全ての履歴を送信するわけではありません。

**会話の履歴を示す情報 (talkHistory / whisperHistory)**
- idx: 会話のインデックス
- day: 会話が行われた日数
- turn: 会話が行われたターン数
- agent: 会話を行ったエージェント
- text: 会話の内容

```
2024/11/06 06:41:47 INFO 発言フェーズを開始します id=01JC03ADVECM4H51QAMWC997E7 day=0
2024/11/06 06:41:47 INFO パケットを送信しました agent=Agent[03] packet="{Request:TALK Info:<nil> Settings:<nil> TalkHistory:0xc000122630 WhisperHistory:<nil>}"
    dummy_client.go:49: recv: {"request":"TALK","talkHistory":[]}
    dummy_client.go:68: send: 5dbdf1ef06e756c2e1f9309cc37b90b4
2024/11/06 06:41:47 INFO レスポンスを受信しました agent=Agent[03] response=5dbdf1ef06e756c2e1f9309cc37b90b4
2024/11/06 06:41:47 INFO 発言がオーバーもしくはスキップではないため、スキップ回数をリセットしました id=01JC03ADVECM4H51QAMWC997E7 agent=Agent[03]
2024/11/06 06:41:47 INFO 発言を受信しました id=01JC03ADVECM4H51QAMWC997E7 agent=Agent[03] text=5dbdf1ef06e756c2e1f9309cc37b90b4
2024/11/06 06:41:47 INFO パケットを送信しました agent=Agent[04] packet="{Request:TALK Info:<nil> Settings:<nil> TalkHistory:0xc000012228 WhisperHistory:<nil>}"
    dummy_client.go:49: recv: {"request":"TALK","talkHistory":[{"idx":0,"day":0,"turn":0,"agent":"Agent[03]","text":"5dbdf1ef06e756c2e1f9309cc37b90b4"}]}
    dummy_client.go:68: send: 3f9d747c94c51311c0ca66226d832d30
2024/11/06 06:41:47 INFO レスポンスを受信しました agent=Agent[04] response=3f9d747c94c51311c0ca66226d832d30
2024/11/06 06:41:47 INFO 発言がオーバーもしくはスキップではないため、スキップ回数をリセットしました id=01JC03ADVECM4H51QAMWC997E7 agent=Agent[04]
2024/11/06 06:41:47 INFO 発言を受信しました id=01JC03ADVECM4H51QAMWC997E7 agent=Agent[04] text=3f9d747c94c51311c0ca66226d832d30
2024/11/06 06:41:47 INFO パケットを送信しました agent=Agent[02] packet="{Request:TALK Info:<nil> Settings:<nil> TalkHistory:0xc0001226c0 WhisperHistory:<nil>}"
    dummy_client.go:49: recv: {"request":"TALK","talkHistory":[{"idx":0,"day":0,"turn":0,"agent":"Agent[03]","text":"5dbdf1ef06e756c2e1f9309cc37b90b4"},{"idx":1,"day":0,"turn":0,"agent":"Agent[04]","text":"3f9d747c94c51311c0ca66226d832d30"}]}
    dummy_client.go:68: send: 5315831399858025fae5d766c0f5196e
2024/11/06 06:41:47 INFO レスポンスを受信しました agent=Agent[02] response=5315831399858025fae5d766c0f5196e
2024/11/06 06:41:47 INFO 発言がオーバーもしくはスキップではないため、スキップ回数をリセットしました id=01JC03ADVECM4H51QAMWC997E7 agent=Agent[02]
2024/11/06 06:41:47 INFO 発言を受信しました id=01JC03ADVECM4H51QAMWC997E7 agent=Agent[02] text=5315831399858025fae5d766c0f5196e
2024/11/06 06:41:47 INFO パケットを送信しました agent=Agent[05] packet="{Request:TALK Info:<nil> Settings:<nil> TalkHistory:0xc0004b22e8 WhisperHistory:<nil>}"
    dummy_client.go:49: recv: {"request":"TALK","talkHistory":[{"idx":0,"day":0,"turn":0,"agent":"Agent[03]","text":"5dbdf1ef06e756c2e1f9309cc37b90b4"},{"idx":1,"day":0,"turn":0,"agent":"Agent[04]","text":"3f9d747c94c51311c0ca66226d832d30"},{"idx":2,"day":0,"turn":0,"agent":"Agent[02]","text":"5315831399858025fae5d766c0f5196e"}]}
    dummy_client.go:68: send: 4458a27ebf237f7a6608c8ae5a589a91
2024/11/06 06:41:47 INFO レスポンスを受信しました agent=Agent[05] response=4458a27ebf237f7a6608c8ae5a589a91
```

### 昼終了リクエスト (DAILY_FINISH)

昼終了リクエストは、昼が終了された際、つまりその日の夜が始まった際に送信されるリクエストです。  
エージェントは、このリクエストを受信した際に、何も返す必要はありません。  
直前までの会話の履歴が送信されます。  
ゲーム全体の人狼の役職が2人未満で囁きフェーズが存在しない場合においても、人狼の役職に対しては、囁きの履歴 (`whisperHistory`) が送信されます。

```
2024/11/06 14:55:05 INFO パケットを送信しました agent=Agent[01] packet="{Request:DAILY_FINISH Info:<nil> Settings:<nil> TalkHistory:0xc00048e390 WhisperHistory:0xc00048e3a8}"
    dummy_client.go:49: recv: {"request":"DAILY_FINISH","talkHistory":[{"idx":13,"day":0,"turn":2,"agent":"Agent[01]","text":"Over"},{"idx":14,"day":0,"turn":2,"agent":"Agent[04]","text":"Over"}],"whisperHistory":[]}
```

### 占いリクエスト (DIVINE)

占いリクエストは、占いが要求された際に送信されるリクエストです。  
占い師 (`SEER`) のみに送信されます。  
エージェントは、このリクエストを受信した際に、占いの対象となるエージェントのインデックス付き文字列を返す必要があります。

> [!WARNING]
> 下記のリクエスト例では、`info` 内に `talkList` が含まれていますが、削除される予定です。

```
2024/11/06 14:55:05 INFO パケットを送信しました agent=Agent[02] packet="{Request:DIVINE Info:0xc0002af6b0 Settings:<nil> TalkHistory:<nil> WhisperHistory:<nil>}"
    dummy_client.go:49: recv: {"request":"DIVINE","info":{"statusMap":{"Agent[01]":"ALIVE","Agent[02]":"ALIVE","Agent[03]":"ALIVE","Agent[04]":"ALIVE","Agent[05]":"ALIVE"},"roleMap":{"Agent[02]":"SEER"},"remainTalkMap":{"Agent[01]":0,"Agent[02]":0,"Agent[03]":0,"Agent[04]":0,"Agent[05]":0},"remainWhisperMap":{},"day":0,"agent":"Agent[02]","talkList":[{"idx":0,"day":0,"turn":0,"agent":"Agent[02]","text":"c6fcc2d3e3ce644aacfc424ec60ec514"},{"idx":1,"day":0,"turn":0,"agent":"Agent[05]","text":"4148eb46c5f1ed1c279a855bfc5ecb39"},{"idx":2,"day":0,"turn":0,"agent":"Agent[03]","text":"2eda40dd4f3c96e12e17834df4d0b9eb"},{"idx":3,"day":0,"turn":0,"agent":"Agent[01]","text":"726b712eb941cac9c8ada4056cad84bd"},{"idx":4,"day":0,"turn":0,"agent":"Agent[04]","text":"f36fbe999cea731323261d8bbbf9eb5b"},{"idx":5,"day":0,"turn":1,"agent":"Agent[02]","text":"171c1cfd36abbd51cec735ec12d3c71b"},{"idx":6,"day":0,"turn":1,"agent":"Agent[05]","text":"2dbc95b6c735f052bc73b4946db94018"},{"idx":7,"day":0,"turn":1,"agent":"Agent[03]","text":"1b306c22512c2e412cc0eb4cd685a136"},{"idx":8,"day":0,"turn":1,"agent":"Agent[01]","text":"d562fd0e9b397d0b896bcad2f81cb0cb"},{"idx":9,"day":0,"turn":1,"agent":"Agent[04]","text":"17a060c1204801816d45a464eb8bf9e1"},{"idx":10,"day":0,"turn":2,"agent":"Agent[02]","text":"Over"},{"idx":11,"day":0,"turn":2,"agent":"Agent[05]","text":"Over"},{"idx":12,"day":0,"turn":2,"agent":"Agent[03]","text":"Over"},{"idx":13,"day":0,"turn":2,"agent":"Agent[01]","text":"Over"},{"idx":14,"day":0,"turn":2,"agent":"Agent[04]","text":"Over"}]}}
    dummy_client.go:68: send: Agent[01]
2024/11/06 14:55:05 INFO レスポンスを受信しました agent=Agent[02] response=Agent[01]
```

### 護衛リクエスト (GUARD)

護衛リクエストは、護衛が要求された際に送信されるリクエストです。  
騎士 (`BODYGUARD`) のみに送信されます。
エージェントは、このリクエストを受信した際に、護衛の対象となるエージェントのインデックス付き文字列を返す必要があります。

> [!WARNING]
> 下記のリクエスト例では、`info` 内に `talkList` が含まれていますが、削除される予定です。

```
2024/11/07 02:55:47 INFO パケットを送信しました agent=Agent[01] packet="{Request:GUARD Info:0xc00033edc0 Settings:<nil> TalkHistory:<nil> WhisperHistory:<nil>}"
    dummy_client.go:49: recv: {"request":"GUARD","info":{"statusMap":{"Agent[01]":"DEAD","Agent[02]":"ALIVE","Agent[03]":"ALIVE","Agent[04]":"ALIVE","Agent[05]":"ALIVE"},"roleMap":{"Agent[01]":"BODYGUARD"},"remainTalkMap":{"Agent[01]":0,"Agent[02]":0,"Agent[03]":0,"Agent[04]":0,"Agent[05]":0},"remainWhisperMap":{},"day":1,"agent":"Agent[01]","talkList":[{"idx":0,"day":1,"turn":0,"agent":"Agent[05]","text":"7cc13d2972f5e2a7c8969ee723ce88c4"},{"idx":1,"day":1,"turn":0,"agent":"Agent[02]","text":"b42dcbc57089bea4638a91d5cc05748a"},{"idx":2,"day":1,"turn":0,"agent":"Agent[04]","text":"6573c92c1be5b20d7569ab71f1e86417"},{"idx":3,"day":1,"turn":0,"agent":"Agent[01]","text":"6b80bc3bcc97509152e9fa8718ba63fe"},{"idx":4,"day":1,"turn":0,"agent":"Agent[03]","text":"8d0edc5a648bda1e727ae8377e07ee7b"},{"idx":5,"day":1,"turn":1,"agent":"Agent[05]","text":"456f740a91a45562ce04b522aa544b93"},{"idx":6,"day":1,"turn":1,"agent":"Agent[02]","text":"ac437b239eb530e47234a13331cc7c64"},{"idx":7,"day":1,"turn":1,"agent":"Agent[04]"
    dummy_client.go:68: send: Agent[02]
2024/11/07 02:55:47 INFO レスポンスを受信しました agent=Agent[01] response=Agent[02]
```

テストのログ出力の仕様により、リクエスト例の文字列が途切れていますが、実際のリクエストは以下のようになります。

```
{"request":"GUARD","info":{"statusMap":{"Agent[01]":"DEAD","Agent[02]":"ALIVE","Agent[03]":"ALIVE","Agent[04]":"ALIVE","Agent[05]":"ALIVE"},"roleMap":{"Agent[03]":"BODYGUARD"},"remainTalkMap":{"Agent[01]":0,"Agent[02]":0,"Agent[03]":0,"Agent[04]":0,"Agent[05]":0},"remainWhisperMap":{},"day":1,"agent":"Agent[03]","talkList":[{"idx":0,"day":1,"turn":0,"agent":"Agent[03]","text":"52489609be6aeda2f583c7328b21f983"},{"idx":1,"day":1,"turn":0,"agent":"Agent[04]","text":"e1a38d894a24ee9db735c7abcc82495b"},{"idx":2,"day":1,"turn":0,"agent":"Agent[02]","text":"22f7930001d6b0ea2400d05e198fa7b5"},{"idx":3,"day":1,"turn":0,"agent":"Agent[01]","text":"0b76445d4c2c0dbe639a058cd651007d"},{"idx":4,"day":1,"turn":0,"agent":"Agent[05]","text":"8bde19e1b8c0d1ff510531308827583e"},{"idx":5,"day":1,"turn":1,"agent":"Agent[03]","text":"32b1d92a6671f043548c3725bb00c1f6"},{"idx":6,"day":1,"turn":1,"agent":"Agent[04]","text":"641743ee5d1988731759e7ec546c21d1"},{"idx":7,"day":1,"turn":1,"agent":"Agent[02]","text":"3914eb8c0c45e2fc9321a7248a9ad9a3"},{"idx":8,"day":1,"turn":1,"agent":"Agent[01]","text":"36414701fb89403b38c2eb13fef57e55"},{"idx":9,"day":1,"turn":1,"agent":"Agent[05]","text":"782c9839bced0e1248a79d96a8b64d95"},{"idx":10,"day":1,"turn":2,"agent":"Agent[03]","text":"Over"},{"idx":11,"day":1,"turn":2,"agent":"Agent[04]","text":"Over"},{"idx":12,"day":1,"turn":2,"agent":"Agent[02]","text":"Over"},{"idx":13,"day":1,"turn":2,"agent":"Agent[01]","text":"Over"},{"idx":14,"day":1,"turn":2,"agent":"Agent[05]","text":"Over"}]}}
```

### 投票リクエスト (VOTE)

投票リクエストは、追放するエージェントを投票する際に送信されるリクエストです。  
エージェントは、このリクエストを受信した際に、投票の対象となるエージェントのインデックス付き文字列を返す必要があります。

> [!WARNING]
> 下記のリクエスト例では、`info` 内に `talkList` が含まれていますが、削除される予定です。

```
2024/11/07 02:55:47 INFO パケットを送信しました agent=Agent[01] packet="{Request:VOTE Info:0xc0002af290 Settings:<nil> TalkHistory:<nil> WhisperHistory:<nil>}"
    dummy_client.go:49: recv: {"request":"VOTE","info":{"statusMap":{"Agent[01]":"ALIVE","Agent[02]":"ALIVE","Agent[03]":"ALIVE","Agent[04]":"ALIVE","Agent[05]":"ALIVE"},"roleMap":{"Agent[01]":"BODYGUARD"},"remainTalkMap":{"Agent[01]":0,"Agent[02]":0,"Agent[03]":0,"Agent[04]":0,"Agent[05]":0},"remainWhisperMap":{},"day":1,"agent":"Agent[01]","talkList":[{"idx":0,"day":1,"turn":0,"agent":"Agent[05]","text":"7cc13d2972f5e2a7c8969ee723ce88c4"},{"idx":1,"day":1,"turn":0,"agent":"Agent[02]","text":"b42dcbc57089bea4638a91d5cc05748a"},{"idx":2,"day":1,"turn":0,"agent":"Agent[04]","text":"6573c92c1be5b20d7569ab71f1e86417"},{"idx":3,"day":1,"turn":0,"agent":"Agent[01]","text":"6b80bc3bcc97509152e9fa8718ba63fe"},{"idx":4,"day":1,"turn":0,"agent":"Agent[03]","text":"8d0edc5a648bda1e727ae8377e07ee7b"},{"idx":5,"day":1,"turn":1,"agent":"Agent[05]","text":"456f740a91a45562ce04b522aa544b93"},{"idx":6,"day":1,"turn":1,"agent":"Agent[02]","text":"ac437b239eb530e47234a13331cc7c64"},{"idx":7,"day":1,"turn":1,"agent":"Agent[04]"
    dummy_client.go:68: send: Agent[01]
2024/11/07 02:55:47 INFO レスポンスを受信しました agent=Agent[01] response=Agent[01]
```

リクエスト例の文字列が途切れていますが、護衛リクエスト同様にテストのログ出力の仕様によりなります。

### 襲撃リクエスト (ATTACK)

襲撃リクエストは、襲撃するエージェントを投票する際に送信されるリクエストです。  
人狼 (`WEREWOLF`) のみに送信されます。  
エージェントは、このリクエストを受信した際に、襲撃の対象となるエージェントのインデックス付き文字列を返す必要があります。

```
2024/11/07 02:55:47 INFO パケットを送信しました agent=Agent[03] packet="{Request:ATTACK Info:0xc00038ed10 Settings:<nil> TalkHistory:<nil> WhisperHistory:<nil>}"
    dummy_client.go:49: recv: {"request":"ATTACK","info":{"statusMap":{"Agent[01]":"DEAD","Agent[02]":"ALIVE","Agent[03]":"ALIVE","Agent[04]":"ALIVE","Agent[05]":"ALIVE"},"roleMap":{"Agent[03]":"WEREWOLF"},"remainTalkMap":{"Agent[01]":0,"Agent[02]":0,"Agent[03]":0,"Agent[04]":0,"Agent[05]":0},"remainWhisperMap":{"Agent[01]":0,"Agent[02]":0,"Agent[03]":15,"Agent[04]":0,"Agent[05]":0},"day":1,"agent":"Agent[03]","talkList":[{"idx":0,"day":1,"turn":0,"agent":"Agent[05]","text":"7cc13d2972f5e2a7c8969ee723ce88c4"},{"idx":1,"day":1,"turn":0,"agent":"Agent[02]","text":"b42dcbc57089bea4638a91d5cc05748a"},{"idx":2,"day":1,"turn":0,"agent":"Agent[04]","text":"6573c92c1be5b20d7569ab71f1e86417"},{"idx":3,"day":1,"turn":0,"agent":"Agent[01]","text":"6b80bc3bcc97509152e9fa8718ba63fe"},{"idx":4,"day":1,"turn":0,"agent":"Agent[03]","text":"8d0edc5a648bda1e727ae8377e07ee7b"},{"idx":5,"day":1,"turn":1,"agent":"Agent[05]","text":"456f740a91a45562ce04b522aa544b93"},{"idx":6,"day":1,"turn":1,"agent":"Agent[02]","text":"ac437b239e
    dummy_client.go:68: send: Agent[02]
2024/11/07 02:55:47 INFO レスポンスを受信しました agent=Agent[03] response=Agent[02]
```

リクエスト例の文字列が途切れていますが、護衛リクエスト同様にテストのログ出力の仕様によりなります。

### ゲーム終了リクエスト (FINISH)

ゲーム終了リクエストは、ゲームが終了された際に送信されるリクエストです。  
エージェントは、このリクエストを受信した際に、何も返す必要はありません。  
各キーについては、ゲーム開始リクエストと同様です。ゲーム開始リクエストとは異なり、`setting` は送信されません。

```
2024/11/07 02:55:47 INFO パケットを送信しました agent=Agent[01] packet="{Request:FINISH Info:0xc000537970 Settings:<nil> TalkHistory:<nil> WhisperHistory:<nil>}"
    dummy_client.go:49: recv: {"request":"FINISH","info":{"statusMap":{"Agent[01]":"DEAD","Agent[02]":"DEAD","Agent[03]":"DEAD","Agent[04]":"ALIVE","Agent[05]":"ALIVE"},"roleMap":{"Agent[01]":"BODYGUARD","Agent[02]":"SEER","Agent[03]":"WEREWOLF","Agent[04]":"VILLAGER","Agent[05]":"VILLAGER"},"remainTalkMap":{},"remainWhisperMap":{},"day":3,"agent":"Agent[01]","executedAgent":"Agent[03]"}}
```