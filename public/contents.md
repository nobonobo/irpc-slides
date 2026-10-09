<!-- .slide: class="top-page" data-background-video="dist/drive.mp4" data-background-video-loop data-background-video-muted data-background-size="cover" -->
## TinyGoとGoをつなぐRPC
### 発表者: NoboNobo 
### 所属: (株)スイッチサイエンス

Note:
メモ書き

---

## 「irpc」の紹介

### https://github.com/marben/irpc

---

## 「irpc」の特徴

- 実行時リフレクション依存なし
- 双方向通信路ならなんでも
- IDLのかわりをGo記述で
- 双方向RPCライブラリ
- gRPCより軽量で高速

---

## 実行時リフレクション依存なし

<ul>
<li>事前にコード生成を行う</li>
<li>生成コードにはreflectは含まれない</li>
<li class="fragment">つまり<strong>「TinyGoでも使える！」</strong></li>
</ul>

---

## 双方向通信路ならなんでも

- TCP-Conn
- Unix-Socket
- WebSocketのラップ
- WebRTCのDataChannelのラップ
- シリアルポート(シングルのみ)
- Stdin/Stdoutのラップ(シングルのみ)
- I2C/SPIなどのラップ(シングルのみ)

その他、io.ReadWriteCloserを実装<br/>さえしていれば何でも使える <!-- .element: class="fragment" -->

---

## IDLのかわりをGo記述で

```bash
go install github.com/marben/irpc/cmd/irpc@latest
```
でジェネレーターをインストール

```go
package chat
//go:generate irpc $GOFILE
type Message struct {
	Name string
	Text string
}
type Event interface {
	OnMessage(msg Message)
}
type Session interface {
	Send(msg Message) error
}
```
Let's Go Generate! <!-- .element: class="fragment" -->

---

## irpc 生成コード

インターフェースごとに `New<Name>IrpcService` と `New<Name>IrpcClient` という名のコンストラクタが生成される

| コンストラクタ  | 主な処理 |
|----------|----------|
| `NewEventIrpcService(impl Event) *EventIrpcService` | ユーザ実装からサーバ側サービスを作成 |
| `NewEventIrpcClient(endpoint irpcgen.Endpoint) (*EventIrpcClient, error)` | エンドポイントに紐づくクライアントを作成 |
| `NewSessionIrpcService(impl Session) *SessionIrpcService` | ユーザ実装からサーバ側サービスを作成 |
| `NewSessionIrpcClient(endpoint irpcgen.Endpoint) (*SessionIrpcClient, error)` | エンドポイントに紐づくクライアントを作成 |
<!-- .element: style="font-size: 60%" -->

---

## サーバー実装

```go
import "irpc-demo/services/chat" // 生成されたコードのパッケージ名

func main() {
  impl := SessionImpl{} // chat.Sessionインターフェースを実装
  service := chat.NewSessionIrpcService(impl)
  irpcServer := irpc.NewServer(irpc.WithServices(service))
  tcpListener, err := net.Listen("tcp", "127.0.0.1:8888")
  if err != nil {
    log.Fatalf("error opening tcp listener: %v", err)
  }
  if err := irpcServer.Serve(tcpListener); err != nil {
    log.Println("irpcServer.Serve():", err)
  }
}
```

---

## クライアント実装

```go
import "irpc-demo/services/chat" // 生成されたコードのパッケージ名

func main() {
  tcpConn, _ := net.Dial("tcp", "127.0.0.1:8888")
  ep := irpc.NewEndpoint(tcpConn)
  client, _ := chat.NewSessionIrpcClient(ep)
  if err := client.Send(chat.Message{
    Name: "Nobo",
    Text: "Hello",
  }); err != nil {
    // エラーハンドリング
  }
}
```

---

## 双方向RPCについて

### 1. Event実装
サーバ側がクライアント側を呼び出す。

### 2. Session実装
クライアント側がサーバ側を呼び出す。

---

### 双方向サーバー実装

```go
server := irpc.NewServer(
  irpc.WithOnConnect(func(ep *irpc.Endpoint) {
    event, _ := chat.NewEventIrpcClient(ep)
    // Session実装の中でeventクライアントを利用
    service := &impl{event:event}
    ep.RegisterService(
      chat.NewSessionIrpcService(),
    )
  }),
)
server.Serve(listener)
```

---

### 双方向クライアント実装

```go
conn, _ := net.Dial("tcp", "127.0.0.1:8080")
// Eventサービス実装のサービス化
event := chat.NewEventIrpcService(&impl{})
ep := irpc.NewEndpoint(conn, irpc.WithEndpointServices(event))
client, _ := chat.NewSessionIrpcClient(ep)
if err := client.Send(chat.Message{
  Name: "Nobo",
  Text: "Hello",
}); err != nil {
  // エラーハンドリング
}
// epを閉じないで待機しておくとEventサービスが利用可能
```

---

## Chat Demo

リポジトリ: https://github.com/nobonobo/irpc-demo

#### サーバービルド＆起動
```bash
go install github.com/marben/irpc/cmd/irpc@latest
go generate ./services/chat/
go build ./cmd/chat/server && ./server
```

#### クライアント実行
```bash
go run ./cmd/chat/client
```

---

## Chat Demo

---

## DIYハンコンの設定GUI

- DIYハンコンはTinyGo実装
- irpcサービス/TinyGoを可動
- 通信路はUSB-CDC接続を利用
- マイコンはstdioとして扱う
- GUI/Goアプリ側はシリアルポートとして扱う
- GUI/Goアプリからirpcサービスを利用

---

## Go-TinyGo / irpc

![](/mermaid-diagram.png)<!-- .element: class="r-stretch" -->

---

## サービス定義

<small>https://github.com/nobonobo/diy-controller/blob/main/service/service.go</small>

GUIが利用している定義部分
```Go
type Service interface {
	Gains() map[string]int32
	SetGains(s map[string]int32)
	Settings() map[string]int32
	SetSettings(s map[string]int32)
	Store() error
	Load() error
	Reset() error
}
```

---

## irpc メモリリークIssue

- 初期バージョンではメモリリークに悩まされてた
- マイコンではより早く破綻した
- よく実装を追うとリソースを閉じれていないパスがあった

https://github.com/marben/irpc/pull/3 → マージ済み

---

## Demonstration

展示でデモするので見に来て！

---

## まとめ

- IDLを別途学ぶ必要なし
- gRPCより軽快に動く
- 適用環境を選ばない
  - PC/WASM/Microcontroller
- 通信物理層も自由に選択
  - TCP/UnixSocket/WebSocket/Serial/StdIO/WebRTC-DC/I2C/SPI/etc.
- 唯一の欠点は他の言語処理系とのインターオペラビリティが無い<!-- .element: class="fragment" -->

---

## 使い分け

- 他の処理系とのインターオペラビリティが必要ならJSON-RPCやgRPCなど
- TinyGo側とGo側だけで良いならirpcでいいんじゃないか

---

# おわり