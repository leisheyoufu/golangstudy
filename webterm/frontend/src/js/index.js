var Terminal = require('xterm').Terminal;
Terminal.applyAddon(require('xterm/lib/addons/fit'));

class Session {
    constructor() {
        this.url = 'ws://' + window.location.host + window.location.pathname + 'term';
        this.ws = new WebSocket(this.url);
        this.term = this.openTerm();
        console.log(this.term);
        this.ws.onmessage = (event) => {
            this.term.write(event.data);
        };
        this.term.on('key', (key, ev) => {
            this.term.fit();
            this.ws.send(key);
        });
    }
    openTerm() {
        var terminalContainer = document.getElementById('term-box');
        var term = new Terminal({
            cursorBlink: true
        });
        term.open(terminalContainer)
        term.fit();
        return term;
    }
}

if (document.readyState === 'complete' || document.readyState !== 'loading') {
    new Session();
} else {
    document.addEventListener('DOMContentLoaded', new Session());
}