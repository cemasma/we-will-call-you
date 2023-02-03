let socket;

const ACTION_MAPPING = {
    0: openControlPanel,
    1: gameStarted,
    2: displayJob,
    3: openHand,
    4: cardPlayed,
    5: updatePlayerList,
};

const PLAYER_TYPES = {
    INTERVIEWER: 1,
    INTERVIEWEE: 2,
};

const ELEMENT_MAPPING = {
    [PLAYER_TYPES.INTERVIEWER]: 'interviewer-card-text',
    [PLAYER_TYPES.INTERVIEWEE]: 'interviewee-card-text',
};

const PROTOCOL_MAPPING = {
  'http:': 'ws:',
  'https:': 'wss:'
};

function openControlPanel() {
    console.log('[openControlPanel] control panel opened.');
    document.getElementById('commands').style.visibility = 'visible';
}

function gameStarted(data) {
    document.getElementById('hand').style.visibility = 'hidden';

    [...Object.values(ELEMENT_MAPPING), 'hand-card-text'].forEach((className) => {
        const elements = document.getElementsByClassName(className);

        for (const element of elements) {
            element.innerText = '?';
            element.style.textDecoration = 'none';
        }
    });

    if (data?.interviewer && data?.interviewee) {
        console.log('[gameStarted]', data);
        document.getElementById("interviewerName").innerText = data.interviewer;
        document.getElementById("intervieweeName").innerText = data.interviewee;
    }
}

function displayJob(data) {
    if (data?.job) {
        console.log('[displayJob]', data.job);

        document.getElementById('job').innerText = data.job;
    }
}

function openHand(data) {
    if (data?.cards) {
        console.log('[openHand]', data.cards);
        (data.cards || []).forEach((card, index) => {
           const cardElements = document.getElementsByClassName("hand-card-text");
           cardElements[index].innerText = card;
        });
        document.getElementById('hand').style.visibility = 'visible';

    }
}

function cardPlayed(data) {
    if (data?.playerType && data?.word) {
        console.log('[cardPlayed]', data);

        const elements = document.getElementsByClassName(ELEMENT_MAPPING[data.playerType]);
        for (const element of elements) {
            if (element.innerText === '?') {
                element.innerText = data.word;
                break;
            }
        }
    }
}

function updatePlayerList(data) {
    if (data?.playerList) {
        console.log('[updatePlayerList]', data.playerList);

        const playerList = document.getElementById('playerList');

        playerList.innerHTML = '';

        (data.playerList || []).forEach((playerName) => {
            const playerNameElement = document.createElement('li');
            playerNameElement.style = 'color: white;';
            playerNameElement.innerText = playerName;
            playerList.appendChild(playerNameElement);
        });
    }
}

window.onload = function () {
    const roomId = document.getElementById('roomId').value;

    let name = document.getElementById('name').value;
    const language = document.getElementById('language').value;

    while (!name) {
        name = prompt('Enter your name:');
    }

    console.log(`${roomId} trying to connect.`);
    socket = new WebSocket(`${PROTOCOL_MAPPING[window.location.protocol]}//${window.location.host}/ws/${roomId}?name=${name}&language=${language}`);

    socket.onopen = function (e) {
        console.log("[open] Connection established");
    };

    socket.onmessage = function (event) {
        console.log(`[message] Data received from server: ${event.data}`);

        const data = JSON.parse(event.data);

        if (typeof ACTION_MAPPING[data?.action] === 'function') {
            ACTION_MAPPING[data.action](data);
        }
    };

    socket.onclose = function (event) {
        if (event.wasClean) {
            console.log(`[close] Connection closed cleanly, code=${event.code} reason=${event.reason}`);
        } else {
            // e.g. server process killed or network down
            // event.code is usually 1006 in this case
            console.log('[close] Connection died');
        }
    };

    socket.onerror = function (error) {
        console.log(`[error] ${error.message}`);
    };
};

function start() {
    console.log('[start]');
    socket.send('/start');
}

function dealCards() {
    console.log('[dealCards]');
    socket.send('/dealcards');
}

function nextInterviewee() {
    console.log('[nextInterviewee]');
    socket.send('/nextInterviewee')
}

function playCard(element) {
    console.log('[playCard]', element, element.innerText);
    socket.send(`/card:${element.innerText}`);
    element.style.textDecoration = 'line-through';
}