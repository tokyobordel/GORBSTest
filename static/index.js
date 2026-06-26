// Элементы для взаимодействия с веб-сокетами
// Телеметрия
const telemetryWebSocket = new WebSocket('ws://localhost:8080/telemetry');
const cpuFreqBlock = document.getElementById("cpuFreq");
const telemetryStatusSpan = document.getElementById("telemetryStatus");
const ramBlock = document.getElementById("ram");
// Эхо
const echoWebSocket = new WebSocket('ws://localhost:8080/echo');
const echoStatusSpan = document.getElementById("echoStatus");
const echoInputField = document.getElementById("echoInput");
const echoSubmitBtn = document.getElementById("echoSubmit");
const echoResponseBlock = document.getElementById("echoResponse");

// Вспомогательные функции
function setEchoWebSocketStatus(isConnected) {
    echoStatusSpan.classList.remove("open");
    echoStatusSpan.classList.remove("error");
    echoStatusSpan.classList.add(isConnected ? "open" : "error");
}

function setTelemetryWebSocketStatus(isConnected) {
    telemetryStatusSpan.classList.remove("open");
    telemetryStatusSpan.classList.remove("error");
    telemetryStatusSpan.classList.add(isConnected ? "open" : "error");
}

function setTelemetryWebSocketError() {
    cpuFreqBlock.innerHTML = "ошибка"
    ramBlock.innerHTML = "ошибка"
}

// Инит вебсокетов
// telemetry
telemetryWebSocket.onopen = () => {
    setTelemetryWebSocketStatus(true);
    setInterval(async () => {
        if (telemetryWebSocket.readyState !== WebSocket.OPEN) {
            setTelemetryWebSocketStatus(false)
            return;
        }
        telemetryWebSocket.send('ping')
    }, 1000)
}
telemetryWebSocket.onmessage = (event) => {
    const data = JSON.parse(event.data);
    cpuFreqBlock.innerHTML = data.cpu_freq + " МГц"
    ramBlock.innerHTML = data.ram + " Мб"
};
telemetryWebSocket.onclose = () => {
    setTelemetryWebSocketStatus(false)
}
telemetryWebSocket.onerror = () => {
    setTelemetryWebSocketError()
}

// echo
echoWebSocket.onopen = () => {
    setEchoWebSocketStatus(true)
}
echoWebSocket.onmessage = (event) => {
    const data = JSON.parse(event.data);
    echoResponseBlock.innerHTML = data.text + ", " + data.timestamp;
};
echoWebSocket.onclose = () => {
    setEchoWebSocketStatus(false)
}
echoWebSocket.onerror = () => {

}

// Обработчик нажатия на кнопку отправки сообщения
echoSubmitBtn.addEventListener('click', async () => {
    if (echoWebSocket.readyState !== WebSocket.OPEN) {
        setEchoWebSocketStatus(false);
        return;
    }

    const payload = JSON.stringify({text: echoInputField.value});
    echoWebSocket.send(payload)
});