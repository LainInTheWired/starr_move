package main

import (
	"log"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
	"github.com/rs/xid"
)
type Move struct{
	x int
	y int 
}

var clients = make(map[*websocket.Conn]string) // 接続されるクライアント
var stackMove []Move
var id string
var clientkey string

// アップグレーダ
var upgrader = websocket.Upgrader{}

func main(){
	// ファイルサーバーを立ち上げる
	fs := http.FileServer(http.Dir("./public"))
	http.Handle("/", fs)

	guid := xid.New()

	clientkey = guid.String()
	//websocketへのルーティング
	http.HandleFunc("/ws/test/kimu", handleConnections)
	go handleMessages()
	log.Println("http server started on :8000")
	err := http.ListenAndServe(":8000", nil)
	// エラーがあった場合ロギングする
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
func handleConnections(w http.ResponseWriter,r * http.Request){
	//送られてきたgetリクエストをwebsocetにアップグレード
	ws, err := upgrader.Upgrade(w, r, nil)

	if err != nil {
        log.Fatal(err)
    }

	//関数が終わったらwebsocektのコネクションをクローズ
	defer ws.Close()

	//クライアントを新しく登録
	clients[ws] = clientkey

	for {
		var cmove Move

		err := ws.ReadJSON(&cmove)
		if err != nil {
			log.Printf("error:%v",err)
			delete(clients,ws)
			break
		}
		//受信されたデータをスタック
		stackMove = append(stackMove, cmove)
	}
}
func handleMessages(){
	t := time.NewTicker(10 * time.Second)
	for {
		select {
		case <-t.C:
			//stackMoveのブロードキャストする
			for client := range clients {
				for _,m := range stackMove{
					err := client.WriteJSON(m)
					if err != nil {
						log.Printf("error: %v",err)
						client.Close()
						delete(clients,client)
					}
				}
			}
			//stackMoveの中を削除
			stackMove = stackMove[:0]
		}
	}
}