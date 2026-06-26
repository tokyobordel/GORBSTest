// Элементы для взаимодействия с веб-сокетами
// Отправка сообщения
const chatWebSocket = new WebSocket('ws://localhost:8080/chat');
const chatTextarea = document.getElementById("chatTextarea")
const chatSendBtn = document.getElementById("chatSend")

// Выгрузка истории чата
const chatHistoryWebSocket = new WebSocket('ws://localhost:8080/chat_history');
const chatHistoryBlock = document.getElementById("chatHistory")

// Вспомогательные функции
function addMessage(message) {
    let html = "<div class='message' id='msg_" + message.id + "'>";
    html += "<div class='content'>" + message.content + "</div>"
    html += "<div class='createdAt'>" + message.created_at + "</div>"
    html += "</div>";
    chatHistoryBlock.insertAdjacentHTML('beforeend', html);

    document.getElementById("msg_" + message.id).scrollIntoView({
        behavior: "smooth",
        block: "start"
    });
}

function displayMessages(data) {
    const messages = JSON.parse(data);
    chatHistoryBlock.innerHTML = null;
    if(messages) {
        for (const message of messages) {
            addMessage(message)
        }
    }
}

// Инит вебсокетов
// chat
chatWebSocket.onopen = () => {

}
chatWebSocket.onmessage = (event) => {
    displayMessages(event.data)
};
chatWebSocket.onclose = () => {

}
chatWebSocket.onerror = () => {

}

// chat_history
chatHistoryWebSocket.onopen = () => {
    chatHistoryWebSocket.send("ping")
}
chatHistoryWebSocket.onmessage = (event) => {
    displayMessages(event.data)
};
chatHistoryWebSocket.onclose = () => {

}
chatHistoryWebSocket.onerror = () => {

}

// Обработчик нажатия на кнопку отправки сообщения
chatSendBtn.addEventListener('click', async () => {
    if(chatWebSocket.readyState !== WebSocket.OPEN) {
        return;
    }

    const payload = JSON.stringify({content: chatTextarea.value});
    chatWebSocket.send(payload)
});