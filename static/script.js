const namePrompt = document.getElementById('namePrompt');
const chat = document.getElementById('chat');
const messages = document.getElementById('messages');
const nameInput = document.getElementById('nameInput');
const enterButton = document.getElementById('enterButton');
const messageInput = document.getElementById('messageInput');
const sendButton = document.getElementById('sendButton');

let ws;
let userName = '';

function initializeWebSocket() {
    ws = new WebSocket('ws://localhost:8080/ws');

    ws.onopen = () => {
        console.log('Подключено к WebSocket серверу');
    };

    ws.onmessage = (event) => {
        const message = document.createElement('div');
        message.className = 'message received';
        message.textContent = event.data;
        messages.appendChild(message);
        chat.scrollTop = chat.scrollHeight; // Автопрокрутка до конца чата
    };

    ws.onclose = () => {
        console.log('Отключено от WebSocket сервера');
    };

    ws.onerror = (error) => {
        console.error('Ошибка WebSocket:', error);
    };
}

enterButton.addEventListener('click', () => {
    const name = nameInput.value.trim();
    if (name) {
        userName = name;
        initializeWebSocket();
        ws.onopen = () => {
            ws.send(userName); // Отправляем имя серверу
            namePrompt.style.display = 'none';
            chat.style.display = 'flex';
        };
    } else {
        alert('Введите ваше имя!');
    }
});

sendButton.addEventListener('click', () => {
    const message = messageInput.value;
    if (message) {
        const messageElement = document.createElement('div');
        messageElement.className = 'message sent';
        messageElement.textContent = `Сообщение от ${userName}: ${message}`;
        messages.appendChild(messageElement);
        chat.scrollTop = chat.scrollHeight; // Автопрокрутка до конца чата
        ws.send(message);
        messageInput.value = '';
    }
});

messageInput.addEventListener('keydown', (event) => {
    if (event.key === 'Enter') {
        sendButton.click();
    }
});
