package websocket

import (
	"encoding/json"
	"fmt"
	"sync"

	"ec.com/models"
	"github.com/gofiber/contrib/socketio"
)

type ClientListMessage struct {
	Event string   `json:"event"`
	Data  []string `json:"data"` // Lista de UserIDs conectados
}

var Clients = make(map[string]string)
var clientsLock sync.RWMutex

// getClientUserIDs: Coleta todas as chaves (User IDs) do mapa Clients de forma segura.
func getClientUserIDs() []string {
	clientsLock.RLock()
	defer clientsLock.RUnlock()

	userIDs := make([]string, 0, len(Clients))
	for userID := range Clients {
		userIDs = append(userIDs, userID)
	}
	return userIDs
}

// broadcastClientList: Empacota a lista de IDs em JSON e a transmite para todos os clientes.
func broadcastClientList(kws *socketio.Websocket) {
	userList := getClientUserIDs()

	// 1. Crie a estrutura de mensagem JSON
	message := ClientListMessage{
		Event: "ACTIVE_CLIENTS_UPDATE",
		Data:  userList,
	}

	// 2. Serializa a estrutura completa
	listJSON, err := json.Marshal(message)
	if err != nil {
		fmt.Println("Erro ao serializar lista de clientes:", err)
		return
	}

	// 3. Transmite a lista para todos os clientes, incluindo a conexão atual (true).
	kws.Broadcast(listJSON, true, socketio.TextMessage)
	fmt.Println("Lista de clientes atualizada e transmitida.")
}

func sendClientListTo(kws *socketio.Websocket) {
	userList := getClientUserIDs()

	message := ClientListMessage{
		Event: "ACTIVE_CLIENTS_UPDATE",
		Data:  userList,
	}

	jsonMsg, _ := json.Marshal(message)

	// envia apenas para o cliente que acabou de conectar
	kws.Emit(jsonMsg, socketio.TextMessage)
}

// --- Handlers de Conexão ---

// WSHandler: Função principal chamada no momento do upgrade HTTP para WebSocket.
func WebSocketHandler(kws *socketio.Websocket) {
	userId := kws.Params("id")

	// 1. REGISTRO SEGURO: Adiciona o cliente ao mapa
	clientsLock.Lock()
	Clients[userId] = kws.UUID
	clientsLock.Unlock()

	// 2. Configura
	kws.SetAttribute("user_id", userId)

	// Envia a lista SOMENTE para o cliente recém conectado
	sendClientListTo(kws)

	// Agora broadcast normal para atualizar todos os outros
	broadcastClientList(kws)
}

func SendMessageToClient(userID string, message models.SocketMessage) {
	jsonMsg, _ := json.Marshal(message)
	if uuid, ok := Clients[userID]; ok {
		socketio.EmitTo(uuid, jsonMsg, socketio.TextMessage)
	}
}

func InitSocket() {
	// EventConnect: Apenas loga o novo cliente. O registro real ocorre no WSHandler.
	socketio.On(socketio.EventConnect, func(ep *socketio.EventPayload) {
		fmt.Printf("Connection event - User: %s\n", ep.Kws.GetStringAttribute("user_id"))

		clientsLock.RLock() // Leitura segura para logging
		fmt.Println("Clients: ", Clients)
		clientsLock.RUnlock()
	})

	socketio.On("CUSTOM_EVENT", func(ep *socketio.EventPayload) {
		// ... (código existente para CUSTOM_EVENT)
	})

	// EventMessage: Trata mensagens de retransmissão e eventos internos.
	socketio.On(socketio.EventMessage, func(ep *socketio.EventPayload) {
		message := models.SocketMessage{}
		if err := json.Unmarshal(ep.Data, &message); err != nil {
			fmt.Println("Erro ao desserializar mensagem recebida:", err)
			return
		}

		// 1. Dispara evento interno (se definido na mensagem)
		if message.Event != "" {
			ep.Kws.Fire(message.Event, []byte(message.Data))
		}

		// 2. Envia a mensagem original (JSON) para o destinatário (message.To)
		destUUID := Clients[message.To]
		if destUUID != "" {
			if err := ep.Kws.EmitTo(destUUID, ep.Data, socketio.TextMessage); err != nil {
				fmt.Println("Erro ao emitir mensagem para", message.To, ":", err)
			}
		} else {
			fmt.Println("Destinatário", message.To, "não encontrado no mapa Clients.")
		}
	})

	// EventDisconnect: Disparo quando o cliente fecha a conexão de forma limpa.
	socketio.On(socketio.EventDisconnect, func(ep *socketio.EventPayload) {
		clientsLock.Lock()
		delete(Clients, ep.Kws.GetStringAttribute("user_id"))
		clientsLock.Unlock()

		fmt.Printf("Disconnection event - User: %s\n", ep.Kws.GetStringAttribute("user_id"))
		broadcastClientList(ep.Kws)
	})

	// EventClose: Disparo em caso de erro ou fechamento forçado da conexão.
	socketio.On(socketio.EventClose, func(ep *socketio.EventPayload) {
		clientsLock.Lock()
		delete(Clients, ep.Kws.GetStringAttribute("user_id"))
		clientsLock.Unlock()

		fmt.Printf("Close event - User: %s\n", ep.Kws.GetStringAttribute("user_id"))
		broadcastClientList(ep.Kws)
	})

	// EventError: Apenas loga o erro.
	socketio.On(socketio.EventError, func(ep *socketio.EventPayload) {
		fmt.Printf("Error event - User: %s\n", ep.Kws.GetStringAttribute("user_id"))
	})

	socketio.On("REQUEST_ACTIVE_CLIENTS", func(ep *socketio.EventPayload) {
		fmt.Println("REQUEST_ACTIVE_CLIENTS recebido de", ep.Kws.GetStringAttribute("user_id"))

		broadcastClientList(ep.Kws)
	})
}
