const telemetryURL = 'http://localhost:8080/telemetry';
const cpuFreqBlock = document.getElementById("cpuFreq");
const ramBlock = document.getElementById("ram");

const echoWebSocket = new WebSocket('ws://localhost:8080/echo');
const echoStatusSpan = document.getElementById("echoStatus");
const echoInputField = document.getElementById("echoInput");
const echoSubmitBtn = document.getElementById("echoSubmit");
const echoResponseBlock = document.getElementById("echoResponse");

const statsInterval = setInterval(async () => {
    try {
        const response = await fetch(telemetryURL);
        const json = await response.json();
        cpuFreqBlock.innerHTML = json.cpu_freq;
        ramBlock.innerHTML = json.ram;
    } catch (error) {
        cpuFreqBlock.innerHTML = 'ошибка';
        ramBlock.innerHTML = 'ошибка';
    }
}, 1000);

function setEchoWebSocketStatus(isConnected) {
    echoStatusSpan.innerHTML = isConnected ? "подключено" : "отключено"
}

function setEchoWebSocketError() {
    echoStatusSpan.innerHTML = "ошибка"
}

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
    setEchoWebSocketError()
}

echoSubmitBtn.addEventListener('click', async (event) => {
    if (echoWebSocket.readyState !== WebSocket.OPEN) {
        setEchoWebSocketStatus(false);
        return;
    }

    const payload = JSON.stringify({text: echoInputField.value});
    echoWebSocket.send(payload)
});